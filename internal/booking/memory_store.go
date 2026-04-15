package booking

type MemoryStore struct {
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}

func (s *MemoryStore) Book(b Booking) error {
	if _, existed := s.bookings[b.SeatID]; existed {
		return ErrSeatAlreadyTaken
	}
	s.bookings[b.SeatID] = b
	return nil
}

func (s *MemoryStore) ListBookings(movieID string) []Booking {
	var result = []Booking{}
	for _, val := range s.bookings {
		if val.MovieID == movieID {
			result = append(result, val)
		}
	}
	return result
}
