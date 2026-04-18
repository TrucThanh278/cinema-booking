package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const DefaultHoldTTL = 2 * time.Minute

// RedisStore implement session-base seat booking backed by Redis.
//
// Key design:
//
// seat:{movieID}:{seatID}	->	session JSON (TTL: held, no TTL: confirmed
// session:{sessionID}		->	seat key (reverse lookup)

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(redisClient *redis.Client) *RedisStore {
	return &RedisStore{
		rdb: redisClient,
	}
}

func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

func (s *RedisStore) Book(b Booking) (Booking, error) {
	session, err := s.hold(b)
	if err != nil {
		return Booking{}, err
	}

	log.Printf("Session Booked: %v", session)
	return session, nil
}

func (r *RedisStore) hold(b Booking) (Booking, error) {
	id := uuid.New().String()
	now := time.Now()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)
	b.ID = id
	val, _ := json.Marshal(b)

	res := r.rdb.SetArgs(ctx, key, val, redis.SetArgs{
		Mode: "NX", // Only set the key if it does not exist. If it does not exist, return "OK". If it exists, return false.
		TTL:  DefaultHoldTTL,
	})

	ok := res.Val() == "OK"

	if !ok {
		return Booking{}, ErrSeatAlreadyTaken
	}

	r.rdb.Set(ctx, sessionKey(id), key, DefaultHoldTTL)

	return Booking{
		ID:        id,
		MovieID:   b.MovieID,
		SeatID:    b.SeatID,
		Status:    "held",
		UserID:    b.UserID,
		ExpiresAt: now.Add(DefaultHoldTTL),
	}, nil
}

// Get all key of redis => get value base on key => add to list book and return
func (r *RedisStore) ListBookings(moveID string) []Booking {
	pattern := fmt.Sprintf("seat:%s:*", moveID)
	var sessions []Booking
	ctx := context.Background()

	iter := r.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		val, err := r.rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		session, err := parseSession(val)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions
}

func (s *RedisStore) getSession(ctx context.Context, sessionID string, userID string) (Booking, string, error) {
	sk, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		return Booking{}, "", err
	}

	val, err := s.rdb.Get(ctx, sk).Result()
	if err != nil {
		return Booking{}, "", err
	}

	session, err := parseSession(val)
	if err != nil {
		return Booking{}, "", err
	}

	return session, sk, nil //sk is session key
}

// Confirm converts a held session into a permanent booking.
// Removes the TTL (PERSIST) so the key never expires.
func (s *RedisStore) Confirm(ctx context.Context, sessionID string, userID string) (Booking, error) {
	session, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return Booking{}, err
	}

	s.rdb.Persist(ctx, sk)
	s.rdb.Persist(ctx, sessionKey(sessionID))

	session.Status = "confirmed"
	data := Booking{
		ID:      session.ID,
		MovieID: string(session.MovieID),
		SeatID:  session.SeatID,
		UserID:  session.UserID,
		Status:  "confirmed",
	}
	val, _ := json.Marshal(data)
	s.rdb.Set(ctx, sk, val, 0)

	return session, nil
}

func (s *RedisStore) Release(ctx context.Context, sessionID string, userID string) error {
	_, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	s.rdb.Del(ctx, sk, sessionKey(sessionID))
	return nil
}

func parseSession(val string) (Booking, error) {
	var data Booking

	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return Booking{}, err
	}

	return Booking{
		ID:      data.ID,
		MovieID: data.MovieID,
		SeatID:  data.SeatID,
		UserID:  data.UserID,
		Status:  data.Status,
	}, nil
}
