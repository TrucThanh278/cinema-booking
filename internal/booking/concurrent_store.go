package booking

import "sync"

type ConcurrentStore struct {
	bookings map[string]Booking
	sync.RWMutex
}

func NewConcurrentStore() *ConcurrentStore {
	return &ConcurrentStore{
		bookings: map[string]Booking{},
	}
}

func (s *ConcurrentStore) Book(b Booking) error {
	s.Lock()
	defer s.Unlock()
	if _, existed := s.bookings[b.SeatID]; existed {
		return ErrSeatAlreadyTaken
	}
	s.bookings[b.SeatID] = b
	return nil
}

func (s *ConcurrentStore) ListBookings(movieID string) []Booking {
	s.RLock()
	defer s.RUnlock()
	var result = []Booking{}
	for _, val := range s.bookings {
		if val.MovieID == movieID {
			result = append(result, val)
		}
	}
	return result
}
