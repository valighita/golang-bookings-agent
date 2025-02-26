package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"valighita/bookings-ai-agent/repository"

	"github.com/prathyushnallamothu/swarmgo"
)

func makeResult(data interface{}, errorMessage string, err error) swarmgo.Result {
	if err != nil {
		log.Printf("%s: %w\n", errorMessage, err)
		return swarmgo.Result{
			Data: fmt.Sprintf("Error: %s", errorMessage),
		}
	}

	result, err := json.Marshal(data)
	if err != nil {
		log.Println("Failed to marshal data: ", err)
		return swarmgo.Result{
			Data: fmt.Sprintf("Error: %s", errorMessage),
		}
	}

	return swarmgo.Result{
		Data: string(result),
	}
}

func GetAgentTools(bookingsRepository repository.BookingRepository, servicesRepository repository.ServiceRepository, employeeRepository repository.EmployeeRepository, debug bool) []swarmgo.AgentFunction {
	debugPrintf := func(format string, args ...interface{}) {
		if debug {
			log.Printf(format, args...)
		}
	}

	return []swarmgo.AgentFunction{
		{
			Name:        "getServices",
			Description: "Get the list of services and their details (duration and price) offered by business.",
			Function: func(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
				debugPrintf("getServices called with args=%v ; contextVars=%v\n", args, contextVariables)
				services, err := servicesRepository.GetServices()
				return makeResult(services, "Failed to get services", err)
			},
		},
		{
			Name:        "getEmployees",
			Description: "Get the list of employees and the services they offer.",
			Function: func(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
				debugPrintf("getEmployees called with args=%v ; contextVars=%v\n", args, contextVariables)
				employees, err := employeeRepository.GetEmployees()
				return makeResult(employees, "Failed to get employees", err)
			},
		},
		{
			Name:        "getServicesForEmployee",
			Description: "Get the list of services offered by a specific employee.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"employee": map[string]interface{}{
						"type":        "string",
						"description": "The name of the employee",
					},
				},
				"required": []interface{}{"employee"},
			},
			Function: func(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
				debugPrintf("getServicesForEmployee called with args=%v ; contextVars=%v\n", args, contextVariables)
				employeeArg, ok := args["employee"].(string)
				if !ok || employeeArg == "" {
					return makeResult(nil, "invalid employee argument", fmt.Errorf("employee is not a string"))
				}
				employee, err := employeeRepository.GetEmployeeByName(employeeArg)
				if err != nil || employee == nil {
					return makeResult(nil, "employee not found", err)
				}

				services, err := employeeRepository.GetServicesByEmployeeId(employee.ID)
				return makeResult(services, "Failed to get services for employee", err)
			},
		},
		{
			Name:        "getEmployeesForService",
			Description: "Get the list of employees who perform a specific service.",
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
			Function: func(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
				debugPrintf("getEmployeesForService called with args=%v ; contextVars=%v\n", args, contextVariables)
				service, ok := args["service"].(string)
				if !ok || service == "" {
					return makeResult(nil, "invalid service argument", fmt.Errorf("service is not a string"))
				}
				serv, err := servicesRepository.GetServiceByName(service)
				if err != nil || serv == nil {
					return makeResult(nil, "service not found", err)
				}

				employees, err := employeeRepository.GetEmployeesForServiceId(serv.ID)
				return makeResult(employees, "Failed to get employees for service", err)
			},
		},
		{
			Name:        "checkAvailability",
			Description: "Check if a employee is available for a booking.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"description": "The name of the service",
					},
					"employee": map[string]interface{}{
						"type":        "string",
						"description": "The name of the employee",
					},
					"date": map[string]interface{}{
						"type":        "string",
						"description": "The date to check, in the format YYYY-MM-DD",
					},
					"time": map[string]interface{}{
						"type":        "string",
						"description": "The time to check, in the format HH:MM",
					},
				},
				"required": []interface{}{"employee", "service", "date", "time"},
			},
			Function: func(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
				debugPrintf("checkAvailability called with args=%v ; contextVars=%v\n", args, contextVariables)
				employeeArg, ok := args["employee"].(string)
				if !ok || employeeArg == "" {
					return makeResult(nil, "invalid employee argument", fmt.Errorf("employee is not a string"))
				}
				serviceArg, ok := args["service"].(string)
				if !ok || serviceArg == "" {
					return makeResult(nil, "invalid service argument", fmt.Errorf("service is not a string"))
				}
				employee, err := employeeRepository.GetEmployeeByName(employeeArg)
				if err != nil || employee == nil {
					return makeResult(nil, "employee not found", err)
				}
				service, err := servicesRepository.GetServiceByName(serviceArg)
				if err != nil || service == nil {
					return makeResult(nil, "service not found", err)
				}

				date, ok := args["date"].(string)
				if !ok {
					return makeResult(nil, "invalid date argument", fmt.Errorf("date is not a string"))
				}
				time, ok := args["time"].(string)
				if !ok {
					return makeResult(nil, "invalid time argument", fmt.Errorf("time is not a string"))
				}

				debugPrintf("checking availability for employeeId: %d serviceId: %d date: %s time: %s",
					employee.ID, service.ID, date, time)
				available, err := employeeRepository.CheckAvailability(employee.ID, service.ID, date, time)
				return makeResult(available, "Failed to check availability", err)
			},
		},
		{
			Name:        "bookAppointment",
			Description: "Book an appointment with a employee for a specific service, date, and time.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"description": "The name of the service",
					},
					"employee": map[string]interface{}{
						"type":        "string",
						"description": "The name of the employee",
					},
					"date": map[string]interface{}{
						"type":        "string",
						"description": "The date to book at, in the format YYYY-MM-DD",
					},
					"time": map[string]interface{}{
						"type":        "string",
						"description": "The time to book at, in the format HH:MM",
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
				"required": []interface{}{"employee", "service", "date", "time", "name", "phone"},
			},
			Function: func(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
				debugPrintf("bookAppointment called with args: %v ; %v\n", args, contextVariables)

				employeeArg, ok := args["employee"].(string)
				if !ok || employeeArg == "" {
					return makeResult(nil, "invalid employee argument", fmt.Errorf("employee is not a string"))
				}
				serviceArg, ok := args["service"].(string)
				if !ok || serviceArg == "" {
					return makeResult(nil, "invalid service argument", fmt.Errorf("service is not a string"))
				}
				employee, err := employeeRepository.GetEmployeeByName(employeeArg)
				if err != nil || employee == nil {
					return makeResult(nil, "service not found", err)
				}
				service, err := servicesRepository.GetServiceByName(serviceArg)
				if err != nil || service == nil {
					return makeResult(nil, "service not found", err)
				}

				date, ok := args["date"].(string)
				if !ok {
					return makeResult(nil, "invalid date argument", fmt.Errorf("date is not a string"))
				}
				bookingTime, ok := args["time"].(string)
				if !ok {
					return makeResult(nil, "invalid time argument", fmt.Errorf("time is not a string"))
				}
				name, ok := args["name"].(string)
				if !ok || name == "" {
					return makeResult(nil, "invalid name argument", fmt.Errorf("name is not a string"))
				}
				phone, ok := args["phone"].(string)
				if !ok || phone == "" {
					return makeResult(nil, "invalid phone argument", fmt.Errorf("phone is not a string"))
				}

				if ok, err := employeeRepository.CheckAvailability(employee.ID, service.ID, date, bookingTime); !ok || err != nil {
					return makeResult(nil, "service is not available", err)
				}

				dateTime, err := time.Parse("2006-01-02 15:04", date+" "+bookingTime)
				if err != nil {
					return makeResult(nil, "invalid date and time", err)
				}

				debugPrintf("booking appointment for employeeId: %d serviceId: %d date: %s time: %s name: %s phone: %s",
					employee.ID, service.ID, date, bookingTime, name, phone)

				booking := repository.Booking{
					ServiceID:       service.ID,
					EmployeeID:      employee.ID,
					BookingDateTime: dateTime,
					CustomerName:    name,
					CustomerPhone:   phone,
				}

				err = bookingsRepository.SaveBooking(&booking)
				return makeResult("ok", "Failed to save booking", err)
			},
		},
	}
}
