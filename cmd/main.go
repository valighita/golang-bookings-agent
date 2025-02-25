package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	customswarm "valighita/bookings-ai-agent/swarm"

	// customswarm "valighita/bookings-ai-agent/swarmgo"

	"github.com/joho/godotenv"
	swarmgo "github.com/prathyushnallamothu/swarmgo"
	llm "github.com/prathyushnallamothu/swarmgo/llm"
)

const (
	llmModel      = "gpt-4o-mini"
	contextPrompt = "You are a helpful booking assistant for a hair salon, helping clients book appointments. " +
		"The salon has multiple stylists, each performing different services with different duration and prices. " +
		"You can use multiple tools. " +
		"Clients can book appointments with one of them and they need to specify a service, a date and a time, a name and a phone number." +
		"It's important to only answer questions related to the hair salon and the services provided, do not provide information about unrelated topics." +
		"Ask the name and phone number as the final info if not already provided. Ask for confirmation before performing the final booking."
)

func getServices(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	fmt.Printf("getServices called with args: %v ; %v\n", args, contextVariables)
	result, _ := json.Marshal([]struct {
		Name        string
		Description string
		Duration    string
		Price       uint
	}{
		{
			Name:        "Men's haircut",
			Description: "A simple haircut for men.",
			Duration:    "30 minutes",
			Price:       20,
		},
		{
			Name:        "Men's beard trimming",
			Description: "Trimming of beard",
			Duration:    "30 minutes",
			Price:       15,
		},
		{
			Name:        "Women's haircut",
			Description: "A simple haircut for women",
			Duration:    "45 minutes",
			Price:       30,
		},
		{
			Name:        "Hair coloring",
			Description: "Coloring of hair",
			Duration:    "1 hour",
			Price:       50,
		},
		{
			Name:        "Hair styling",
			Description: "Styling of hair",
			Duration:    "1 hour",
			Price:       40,
		},
	})

	return swarmgo.Result{
		Data: string(result),
	}
}

var stylists = []struct {
	Name     string
	Services []string
}{
	{
		Name:     "Alice",
		Services: []string{"Women's haircut", "Hair coloring", "Hair styling"},
	},
	{
		Name:     "George",
		Services: []string{"Men's haircut", "Men's beard trimming"},
	},
	{
		Name:     "Emily",
		Services: []string{"Men's haircut", "Women's haircut", "Hair styling"},
	},
}

func getStylists(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	fmt.Printf("getStylists called with args: %v ; %v\n", args, contextVariables)
	result, _ := json.Marshal(stylists)

	return swarmgo.Result{
		Data: string(result),
	}
}

func getServicesForStylist(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	fmt.Printf("getServicesForStylist called with args: %v ; %v\n", args, contextVariables)
	for _, stylist := range stylists {
		if strings.ToLower(stylist.Name) == strings.ToLower(args["stylist"].(string)) {
			fmt.Println("returning stylist services", stylist.Services)

			result, _ := json.Marshal(stylist.Services)

			return swarmgo.Result{
				Data: string(result),
			}
		}
	}
	return swarmgo.Result{
		Data: "Stylist not found",
	}
}

func getStylistsForService(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	fmt.Printf("getStylistsForService called with args: %+v ; %+v\n", args, contextVariables)
	service := args["service"].(string)
	result := []string{}
	for _, stylist := range stylists {
		for _, s := range stylist.Services {
			if strings.ToLower(s) == strings.ToLower(service) {
				result = append(result, stylist.Name)
			}
		}
	}

	if len(result) > 0 {
		res, _ := json.Marshal(result)
		return swarmgo.Result{
			Data: string(res),
		}
	}

	return swarmgo.Result{
		Data: "No stylist found for the service",
	}
}

func checkAvailability(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	fmt.Printf("checkAvailability called with args: %+v ; %+v\n", args, contextVariables)

	// service := args["time"].(string)
	stylist := args["stylist"].(string)

	if strings.ToLower(stylist) == "alice" {
		return swarmgo.Result{
			Data: "true",
		}
	}

	return swarmgo.Result{
		Data: "false",
	}
}

// TODO employee schedule

func getAvailableSlots(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	fmt.Printf("getAvailableSlots called with args: %+v ; %+v\n", args, contextVariables)

	// duration := args["duration"].(string)
	// date := args["date"].(string)

	return swarmgo.Result{
		Data: []string{"10:00", "11:00", "12:00"},
	}
}

func bookAppointment(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	fmt.Printf("bookAppointment called with args: %+v ; %+v\n", args, contextVariables)

	// service := args["service"].(string)
	// stylist := args["stylist"].(string)
	// date := args["date"].(string)
	// time := args["time"].(string)
	// name := args["name"].(string)
	// phone := args["phone"].(string)

	return swarmgo.Result{
		Data: "ok",
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %w", err)
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatalf("OPENAI_API_KEY is required")
	}

	client := customswarm.NewSwarm(openAIKey, llm.OpenAI)
	agent := swarmgo.NewAgent("Booking Agent", llmModel, llm.OpenAI)

	agent.Functions = []swarmgo.AgentFunction{
		{
			Name:        "getServices",
			Description: "Get the list of services and their details (duration and price) offered by the hair salon.",
			Function:    getServices,
		},
		{
			Name:        "getStylists",
			Description: "Get the list of stylists and the services they offer.",
			Function:    getStylists,
		},
		{
			Name:        "getServicesForStylist",
			Description: "Get the list of services offered by a specific stylist.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"stylist": map[string]interface{}{
						"type":        "string",
						"description": "The name of the stylist",
					},
				},
				"required": []interface{}{"stylist"},
			},
			Function: getServicesForStylist,
		},
		{
			Name:        "getStylistsForService",
			Description: "Get the list of stylists who offer a specific service.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"description": "The name of the service",
					},
				},
				"required": []interface{}{"service"},
			},
			Function: getStylistsForService,
		},
		{
			Name:        "checkAvailability",
			Description: "Check if a stylist is available for a booking.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"stylist": map[string]interface{}{
						"type":        "string",
						"description": "The name of the stylist",
					},
					"time": map[string]interface{}{
						"type":        "string",
						"description": "The time to check",
					},
				},
				"required": []interface{}{"stylist", "time"},
			},
			Function: checkAvailability,
		},
		{
			Name:        "getAvailableSlots",
			Description: "Get the available time slots for a stylist on a specific date.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"date": map[string]interface{}{
						"type":        "string",
						"description": "The date to check",
					},
					"duration": map[string]interface{}{
						"type":        "string",
						"description": "The duration of the service",
					},
				},
				"required": []interface{}{"date", "duration"},
			},
			Function: getAvailableSlots,
		},
		{
			Name:        "bookAppointment",
			Description: "Book an appointment with a stylist for a specific service, date, and time.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"description": "The name of the service",
					},
					"stylist": map[string]interface{}{
						"type":        "string",
						"description": "The name of the stylist",
					},
					"date": map[string]interface{}{
						"type":        "string",
						"description": "The date of the appointment",
					},
					"time": map[string]interface{}{
						"type":        "string",
						"description": "The time of the appointment",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The name of the client",
					},
					"phone": map[string]interface{}{
						"type":        "string",
						"description": "The phone number of the client",
					},
				},
				"required": []interface{}{"service", "stylist", "date", "time", "name", "phone"},
			},
			Function: bookAppointment,
		},
	}

	data := make([]byte, 1024)
	idx := 0

	messages := []llm.Message{
		{
			Role:    llm.RoleSystem,
			Content: contextPrompt + ". Current time is " + time.Now().Format("2006-01-02 15:04:05, Monday"),
		},
	}

	for {
		idx += 1
		var n int

		if idx == 0 {
			data = []byte("what are the prices of the services performed by alice?")
			n = len(data)
		} else {
			fmt.Printf("Enter your message: ")
			n, err = os.Stdin.Read(data)
			if err != nil {
				log.Fatalf("Error reading from stdin: %v", err)
			}
			if n == 0 {
				continue
			}
		}

		messages = append(messages, llm.Message{
			Role:    llm.RoleUser,
			Content: string(data[:n]),
		})

		ctx := context.Background()
		response, err := client.Run(ctx, agent, messages, nil, "", false, false, 10, true)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		if len(response.Messages) == 0 {
			fmt.Println("Can't process request.")
			continue
		}

		msg := response.Messages[len(response.Messages)-1]
		fmt.Printf("Response: %s\n", msg.Content)

		messages = append(messages, msg)
	}
}
