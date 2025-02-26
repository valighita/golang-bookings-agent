package memory_repository

import (
	"errors"
	"strings"
	"sync"

	"valighita/bookings-ai-agent/repository"
)

type servicesMemoryRepository struct {
	mu       sync.RWMutex
	services map[uint]*repository.Service
}

func NewServicesMemoryRepository(data map[uint]*repository.Service) repository.ServiceRepository {
	return &servicesMemoryRepository{
		services: data,
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
		if strings.ToLower(service.Name) == strings.ToLower(name) {
			return service, nil
		}
	}

	return nil, errors.New("service not found")
}
