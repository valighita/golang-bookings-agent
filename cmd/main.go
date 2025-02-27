package main

import (
	"log"
	"os"
	"valighita/bookings-ai-agent/agent"
	"valighita/bookings-ai-agent/repository"
	memory_repository "valighita/bookings-ai-agent/repository/memory"
	"valighita/bookings-ai-agent/server"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %w", err)
	}

	bookingsRepository := memory_repository.NewBookingsMemoryRepository()

	// services for a dental clinic
	servicesRepository := memory_repository.NewServicesMemoryRepository(map[uint]*repository.Service{
		1: {
			ID:       1,
			Name:     "Dental Cleaning",
			Duration: 30,
			Price:    100,
		},
		2: {
			ID:       2,
			Name:     "Dental Filling",
			Duration: 60,
			Price:    200,
		},
		3: {
			ID:       3,
			Name:     "Dental Crown",
			Duration: 90,
			Price:    300,
		},
		4: {
			ID:       4,
			Name:     "Dental Implant",
			Duration: 120,
			Price:    400,
		},
		5: {
			ID:       5,
			Name:     "Dental Extraction",
			Duration: 45,
			Price:    150,
		},
		6: {
			ID:       6,
			Name:     "Dental X-Ray",
			Duration: 15,
			Price:    50,
		},
	})
	employeeRepository := memory_repository.NewEmployeeMemoryRepository(bookingsRepository, servicesRepository, map[uint]*repository.Employee{
		1: {
			ID:          1,
			Name:        "Alice",
			ServicesIds: []uint{1, 2, 3, 5},
		},
		2: {
			ID:          2,
			Name:        "Bob",
			ServicesIds: []uint{1, 4},
		},
		3: {
			ID:          3,
			Name:        "Charlie",
			ServicesIds: []uint{1, 2, 3, 4},
		},
		4: {
			ID:          4,
			Name:        "David",
			ServicesIds: []uint{1, 4, 5},
		},
		5: {
			ID:          5,
			Name:        "George",
			ServicesIds: []uint{5, 6},
		},
	})

	debugMode := os.Getenv("DEBUG_MODE") == "true"
	agentTools := agent.GetAgentTools(bookingsRepository, servicesRepository, employeeRepository, debugMode)
	agentFactory := agent.NewOpenaiAgentFactory(agentTools, debugMode)

	server.RunHttpServer(agentFactory)
}
