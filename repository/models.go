package repository

import "time"

type Employee struct {
	ID          uint
	Name        string
	Description string
	ServicesIds []uint
}

type EmployeeRepository interface {
	GetEmployees() ([]*Employee, error)
	GetEmployeeById(id uint) (*Employee, error)
	GetEmployeeByName(name string) (*Employee, error)
	CheckAvailability(employeeId uint, serviceId uint, bookingDate string, bookingTime string) (bool, error)
	GetServicesByEmployeeId(employeeId uint) ([]*Service, error)
	GetEmployeesForServiceId(serviceId uint) ([]*Employee, error)
}

type Service struct {
	ID       uint
	Name     string
	Price    float64
	Duration uint // In minutes
}

type ServiceRepository interface {
	GetServices() ([]*Service, error)
	GetServiceById(id uint) (*Service, error)
	GetServiceByName(name string) (*Service, error)
}

type Booking struct {
	ID              uint
	EmployeeID      uint
	ServiceID       uint
	BookingDateTime time.Time
	CustomerName    string
	CustomerPhone   string
}

type BookingRepository interface {
	GetBookingsByDateAndEmployee(date string, employeeId uint) ([]*Booking, error)
	SaveBooking(booking *Booking) error
}
