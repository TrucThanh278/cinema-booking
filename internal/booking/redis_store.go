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

func (s *RedisStore) Book(b Booking) error {
	session, err := s.hold(b)
	if err != nil {
		return err
	}

	log.Printf("Session Booked: %v", session)
	return nil
}

func (r *RedisStore) hold(b Booking) (Booking, error) {
	id := uuid.New().String()
	now := time.Now()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)
	b.ID = id
	val, _ := json.Marshal(b)

	res := r.rdb.SetArgs(ctx, key, val, redis.SetArgs{
		Mode: "NX", // Only set the key if it does not exist. If it does not exist, return true. If it exists, return false.
		TTL:  DefaultHoldTTL,
	})

	ok := res.Val() == "OK"

	if !ok {
		return Booking{}, ErrSeatAlreadyTaken
	}

	r.rdb.Set(ctx, sessionKey(id), key, DefaultHoldTTL)

	return Booking{
		ID:       id,
		MovieID:  b.MovieID,
		SeatID:   b.SeatID,
		Status:   "held",
		UserID:   b.UserID,
		ExpireAt: now.Add(DefaultHoldTTL),
	}, nil
}

func (r *RedisStore) ListBookings(moveID string) []Booking {
	return []Booking{}
}
