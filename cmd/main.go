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
	servicesRepository := memory_repository.NewServicesMemoryRepository(map[uint]*repository.Service{
		1: {
			ID:       1,
			Name:     "Men's Haircut",
			Duration: 30,
			Price:    50,
		},
		2: {
			ID:       2,
			Name:     "Hair coloring",
			Duration: 60,
			Price:    100,
		},
		3: {
			ID:       3,
			Name:     "Hair styling",
			Duration: 45,
			Price:    75,
		},
		4: {
			ID:       4,
			Name:     "Men's beard trimming",
			Duration: 15,
			Price:    30,
		},
		5: {
			ID:       5,
			Name:     "Women's Haircut",
			Duration: 45,
			Price:    70,
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
	})

	debugMode := os.Getenv("DEBUG_MODE") == "true"
	agentTools := agent.GetAgentTools(bookingsRepository, servicesRepository, employeeRepository, debugMode)
	agentFactory := agent.NewOpenaiAgentFactory(agentTools, debugMode)

	server.RunHttpServer(agentFactory)
}
