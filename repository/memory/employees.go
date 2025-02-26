package memory_repository

import (
	"errors"
	"sync"
	"time"
	"valighita/bookings-ai-agent/repository"
)

type employeeMemoryRepository struct {
	mu                 sync.RWMutex
	employees          map[uint]*repository.Employee
	bookingsRepository repository.BookingRepository
	serviceRepository  repository.ServiceRepository
}

func NewEmployeeMemoryRepository(bookingRepository repository.BookingRepository, serviceRepository repository.ServiceRepository, data map[uint]*repository.Employee) repository.EmployeeRepository {
	return &employeeMemoryRepository{
		employees:          data,
		bookingsRepository: bookingRepository,
		serviceRepository:  serviceRepository,
	}
}

func (r *employeeMemoryRepository) GetEmployees() ([]*repository.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employees := make([]*repository.Employee, 0, len(r.employees))
	for _, employee := range r.employees {
		employees = append(employees, employee)
	}

	return employees, nil
}

func (r *employeeMemoryRepository) GetEmployeeById(id uint) (*repository.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employee, ok := r.employees[id]
	if !ok {
		return nil, errors.New("employee not found")
	}

	return employee, nil
}

func (r *employeeMemoryRepository) GetEmployeeByName(name string) (*repository.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, employee := range r.employees {
		if employee.Name == name {
			return employee, nil
		}
	}

	return nil, errors.New("employee not found")
}

func (r *employeeMemoryRepository) CheckAvailability(employeeId uint, serviceId uint, bookingDate string, bookingTime string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, err := r.serviceRepository.GetServiceById(serviceId)
	if err != nil {
		return false, err
	}

	// Use the info in booking repository to check availability
	dayBookings, err := r.bookingsRepository.GetBookingsByDateAndEmployee(bookingDate, employeeId)
	if err != nil {
		return false, err
	}

	checkTime, err := time.Parse("2006-01-02 15:04:05", bookingDate+" "+bookingTime)
	if err != nil {
		return false, err
	}
	checkEndTime := checkTime.Add(time.Duration(service.Duration) * time.Minute)

	for _, booking := range dayBookings {
		bookingEndTime := booking.BookingDateTime.Add(time.Duration(service.Duration) * time.Minute)
		if checkTime.Before(bookingEndTime) && booking.BookingDateTime.Before(checkEndTime) {
			return false, nil // There is an overlap
		}
	}

	return true, nil
}

func (r *employeeMemoryRepository) GetServicesByEmployeeId(employeeId uint) ([]*repository.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employee, err := r.GetEmployeeById(employeeId)
	if err != nil {
		return nil, err
	}

	if employee == nil {
		return nil, errors.New("employee not found")
	}

	employeeServices := make([]*repository.Service, 0, len(employee.ServicesIds))
	for _, serviceId := range employee.ServicesIds {
		service, err := r.serviceRepository.GetServiceById(serviceId)
		if err != nil {
			return nil, err
		}
		employeeServices = append(employeeServices, service)
	}

	return employeeServices, nil
}
