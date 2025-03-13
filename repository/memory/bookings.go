package memory_repository

import (
	"errors"
	"sync"
	"time"

	"valighita/bookings-ai-agent/repository"
)

type bookingsMemoryRepository struct {
	mu sync.RWMutex
	// map that stores the bookings indexed by date
	bookings map[string][]*repository.Booking
	nextID   uint
}

func NewBookingsMemoryRepository() repository.BookingRepository {
	return &bookingsMemoryRepository{
		bookings: make(map[string][]*repository.Booking),
		nextID:   1,
	}
}

func (r *bookingsMemoryRepository) GetBookingsByDateAndEmployee(date string, employeeId uint) ([]*repository.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	dateBookings, ok := r.bookings[date]
	if !ok {
		return nil, nil
	}

	var bookings []*repository.Booking
	for _, booking := range dateBookings {
		if booking.EmployeeID == employeeId && booking.BookingDateTime.Format("2006-01-02") == parsedDate.Format("2006-01-02") {
			bookings = append(bookings, booking)
		}
	}

	return bookings, nil
}

func (r *bookingsMemoryRepository) SaveBooking(booking *repository.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if booking.ID == 0 {
		booking.ID = r.nextID
		r.nextID++
	}

	if booking.BookingDateTime.Before(time.Now()) {
		return errors.New("booking time is in the past")
	}

	date := booking.BookingDateTime.Format("2006-01-02")
	bookings, ok := r.bookings[date]
	if !ok {
		r.bookings[date] = []*repository.Booking{booking}
	} else {
		r.bookings[date] = append(bookings, booking)
	}

	return nil
}
