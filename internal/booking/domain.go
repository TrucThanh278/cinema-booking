package booking

import (
	"errors"
	"time"
)

var ErrSeatAlreadyTaken = errors.New("seat is already taken")

// Booking represents a confirmed seat reservation.
type Booking struct {
	ID       string
	MovieID  string
	SeatID   string
	UserID   string
	Status   string
	ExpireAt time.Time
}

type BookingStore interface {
	Book(b Booking) error
	ListBookings(movieID string) []Booking
}
