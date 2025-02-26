package memory_repository

import (
	"errors"
	"slices"
	"sync"

	"valighita/bookings-ai-agent/repository"
)

type servicesMemoryRepository struct {
	mu                 sync.RWMutex
	services           map[uint]*repository.Service
	employeeRepository repository.EmployeeRepository
}

func NewServicesMemoryRepository(employeeRepository repository.EmployeeRepository, data map[uint]*repository.Service) repository.ServiceRepository {
	return &servicesMemoryRepository{
		services:           data,
		employeeRepository: employeeRepository,
	}
}

func (r *servicesMemoryRepository) GetServices() ([]*repository.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*repository.Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}

	return services, nil
}

func (r *servicesMemoryRepository) GetServiceById(id uint) (*repository.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, service := range r.services {
		if service.ID == id {
			return service, nil
		}
	}

	return nil, nil
}

func (r *servicesMemoryRepository) GetServiceByName(name string) (*repository.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, service := range r.services {
		if service.Name == name {
			return service, nil
		}
	}

	return nil, errors.New("service not found")
}

func (r *servicesMemoryRepository) GetEmployeesForServiceId(serviceId uint) ([]*repository.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	employees, err := r.employeeRepository.GetEmployees()
	if err != nil {
		return nil, err
	}

	var employeesForService []*repository.Employee

	for _, employee := range employees {
		if slices.Contains(employee.ServicesIds, serviceId) {
			employeesForService = append(employeesForService, employee)
		}
	}

	return employees, nil
}
