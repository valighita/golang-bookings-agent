package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"valighita/bookings-ai-agent/repository"

	langchaintools "github.com/tmc/langchaingo/tools"
)

func makeResult(data interface{}, errorMessage string, err error) string {
	if err != nil {
		log.Printf("%s: %v\n", errorMessage, err)
		return fmt.Sprintf("Error: %s", errorMessage)
	}

	result, err := json.Marshal(data)
	if err != nil {
		log.Println("Failed to marshal data: ", err)
		return fmt.Sprintf("Error: %s", errorMessage)
	}

	return string(result)
}

type getServicesTool struct {
	servicesRepository repository.ServiceRepository
	logFunc            func(format string, v ...interface{})
}

func (t *getServicesTool) Name() string {
	return "getServices"
}

func (t *getServicesTool) Description() string {
	return "Get the list of services and their details (duration and price) offered by business."
}

func (t *getServicesTool) Call(ctx context.Context, input string) (string, error) {
	t.logFunc("getServices called with ctx=%v ; input=%v\n", ctx, input)
	services, err := t.servicesRepository.GetServices()
	return makeResult(services, "Failed to get services", err), nil
}

type getEmployeesTool struct {
	employeesRepository repository.EmployeeRepository
	logFunc             func(format string, v ...interface{})
}

func (t *getEmployeesTool) Name() string {
	return "getEmployees"
}

func (t *getEmployeesTool) Description() string {
	return "Get the list of employees and the services they offer."
}

func (t *getEmployeesTool) Call(ctx context.Context, input string) (string, error) {
	t.logFunc("getEmployees called with ctx=%v ; input=%v\n", ctx, input)
	employees, err := t.employeesRepository.GetEmployees()
	return makeResult(employees, "Failed to get services", err), nil
}

type getServicesForEmployeeTool struct {
	employeesRepository repository.EmployeeRepository
	logFunc             func(format string, v ...interface{})
}

func (t *getServicesForEmployeeTool) Name() string {
	return "getServicesForEmployee"
}

func (t *getServicesForEmployeeTool) Description() string {
	return "Get the list of services offered by a specific employee." +
		"Input is a JSON object with the following fields: employee."
}

func (t *getServicesForEmployeeTool) Call(ctx context.Context, input string) (string, error) {
	t.logFunc("getServicesForEmployee called with ctx=%v ; input=%v\n", ctx, input)

	var inputMap map[string]string
	err := json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return makeResult(nil, "invalid input", err), nil
	}
	employeeArg, ok := inputMap["employee"]

	if !ok || employeeArg == "" {
		return makeResult(nil, "invalid employee argument", fmt.Errorf("employee is not a string")), nil
	}
	employee, err := t.employeesRepository.GetEmployeeByName(employeeArg)
	if err != nil || employee == nil {
		return makeResult(nil, "employee not found", err), nil
	}

	// TODO return employee service names instead of IDs

	services, err := t.employeesRepository.GetServicesByEmployeeId(employee.ID)
	return makeResult(services, "Failed to get services for employee", err), nil
}

type getEmployeesForServiceTool struct {
	employeesRepository repository.EmployeeRepository
	servicesRepository  repository.ServiceRepository
	logFunc             func(format string, v ...interface{})
}

func (t *getEmployeesForServiceTool) Name() string {
	return "getEmployeesForService"
}

func (t *getEmployeesForServiceTool) Description() string {
	return "Get the list of employees who perform a specific service." +
		"Input is a JSON object with the following fields: service."
}

func (t *getEmployeesForServiceTool) Call(ctx context.Context, input string) (string, error) {
	t.logFunc("getEmployeesForService called with ctx=%v ; input=%v\n", ctx, input)

	var inputMap map[string]string
	err := json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return makeResult(nil, "invalid input", err), nil
	}
	serviceArg, ok := inputMap["service"]

	if !ok || serviceArg == "" {
		return makeResult(nil, "invalid service argument", fmt.Errorf("service is not a string")), nil
	}
	service, err := t.servicesRepository.GetServiceByName(serviceArg)
	if err != nil || service == nil {
		return makeResult(nil, "service not found", err), nil
	}

	employees, err := t.employeesRepository.GetEmployeesForServiceId(service.ID)
	t.logFunc("returning employees for service id %d: %v\n", service.ID, employees)
	return makeResult(employees, "Failed to get employees for service", err), nil
}

//		Function: func(args map[string]interface{}, contextVariables map[string]interface{}) langchaingo.Result {
//			debugPrintf("checkAvailability called with args=%v ; contextVars=%v\n", args, contextVariables)
//
//		},
//	},
type checkAvailabilityTool struct {
	employeesRepository repository.EmployeeRepository
	servicesRepository  repository.ServiceRepository
	logFunc             func(format string, v ...interface{})
}

func (t *checkAvailabilityTool) Name() string {
	return "checkAvailability"
}

func (t *checkAvailabilityTool) Description() string {
	return "Check if an employee is available for a booking at a given time and date." +
		"Input is a JSON object with the following fields: employee, service, date, time." +
		"All fields are required and the date and time should be in the format YYYY-MM-DD and HH:MM"
}

func (t *checkAvailabilityTool) Call(ctx context.Context, input string) (string, error) {
	t.logFunc("checkAvailability called with ctx=%v ; input=%v\n", ctx, input)

	var inputMap map[string]string
	err := json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return makeResult(nil, "invalid input", err), nil
	}

	employeeArg, ok := inputMap["employee"]
	if !ok || employeeArg == "" {
		return makeResult(nil, "invalid employee argument", fmt.Errorf("employee is not a string")), nil
	}
	serviceArg, ok := inputMap["service"]
	if !ok || serviceArg == "" {
		return makeResult(nil, "invalid service argument", fmt.Errorf("service is not a string")), nil
	}
	employee, err := t.employeesRepository.GetEmployeeByName(employeeArg)
	if err != nil || employee == nil {
		return makeResult(nil, "employee not found", err), nil
	}
	service, err := t.servicesRepository.GetServiceByName(serviceArg)
	if err != nil || service == nil {
		return makeResult(nil, "service not found", err), nil
	}

	employeeServices, err := t.employeesRepository.GetServicesByEmployeeId(employee.ID)
	if err != nil || service == nil {
		return makeResult(nil, "could not get services for employee", err), nil
	}
	found := false
	for _, s := range employeeServices {
		if s.ID == service.ID {
			found = true
			break
		}
	}
	if !found {
		return makeResult(nil, "employee does not offer the service", fmt.Errorf("employee does not offer the service")), nil
	}

	date, ok := inputMap["date"]
	if !ok {
		return makeResult(nil, "invalid date argument", fmt.Errorf("date is not a string")), nil
	}
	time, ok := inputMap["time"]
	if !ok {
		return makeResult(nil, "invalid time argument", fmt.Errorf("time is not a string")), nil
	}

	t.logFunc("checking availability for employeeId: %d serviceId: %d date: %s time: %s",
		employee.ID, service.ID, date, time)
	available, err := t.employeesRepository.CheckAvailability(employee.ID, service.ID, date, time)
	return makeResult(available, "Failed to check availability", err), nil
}

type bookAppointmentTool struct {
	employeesRepository repository.EmployeeRepository
	servicesRepository  repository.ServiceRepository
	bookingsRepository  repository.BookingRepository
	logFunc             func(format string, v ...interface{})
}

func (t *bookAppointmentTool) Name() string {
	return "bookAppointment"
}

func (t *bookAppointmentTool) Description() string {
	return "Book an appointment with an employee for a specific service, date, and time." +
		"Input is a JSON object with the following fields: employee, service, date, time, name, phone." +
		"All fields are required and the date and time should be in the format YYYY-MM-DD and HH:MM"
}

func (t *bookAppointmentTool) Call(ctx context.Context, input string) (string, error) {
	t.logFunc("bookAppointment called with args: %v ; %v\n", ctx, input)

	var inputMap map[string]string
	err := json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return makeResult(nil, "invalid input", err), nil
	}

	employeeArg, ok := inputMap["employee"]
	if !ok || employeeArg == "" {
		return makeResult(nil, "invalid employee argument", fmt.Errorf("employee is not a string")), nil
	}
	serviceArg, ok := inputMap["service"]
	if !ok || serviceArg == "" {
		return makeResult(nil, "invalid service argument", fmt.Errorf("service is not a string")), nil
	}
	employee, err := t.employeesRepository.GetEmployeeByName(employeeArg)
	if err != nil || employee == nil {
		return makeResult(nil, "service not found", err), nil
	}
	service, err := t.servicesRepository.GetServiceByName(serviceArg)
	if err != nil || service == nil {
		return makeResult(nil, "service not found", err), nil
	}

	employeeServices, err := t.employeesRepository.GetServicesByEmployeeId(employee.ID)
	if err != nil || service == nil {
		return makeResult(nil, "could not get services for employee", err), nil
	}
	found := false
	for _, s := range employeeServices {
		if s.ID == service.ID {
			found = true
			break
		}
	}
	if !found {
		return makeResult(nil, "employee does not offer the service", fmt.Errorf("employee does not offer the service")), nil
	}

	date, ok := inputMap["date"]
	if !ok {
		return makeResult(nil, "invalid date argument", fmt.Errorf("date is not a string")), nil
	}
	bookingTime, ok := inputMap["time"]
	if !ok {
		return makeResult(nil, "invalid time argument", fmt.Errorf("time is not a string")), nil
	}
	name, ok := inputMap["name"]
	if !ok || name == "" {
		return makeResult(nil, "invalid name argument", fmt.Errorf("name is not a string")), nil
	}
	phone, ok := inputMap["phone"]
	if !ok || phone == "" {
		return makeResult(nil, "invalid phone argument", fmt.Errorf("phone is not a string")), nil
	}

	if ok, err := t.employeesRepository.CheckAvailability(employee.ID, service.ID, date, bookingTime); !ok || err != nil {
		return makeResult(nil, "employee is not available", err), nil
	}

	dateTime, err := time.Parse("2006-01-02 15:04", date+" "+bookingTime)
	if err != nil {
		return makeResult(nil, "invalid date and time", err), nil
	}

	t.logFunc("booking appointment for employeeId: %d serviceId: %d date: %s time: %s name: %s phone: %s",
		employee.ID, service.ID, date, bookingTime, name, phone)

	booking := repository.Booking{
		ServiceID:       service.ID,
		EmployeeID:      employee.ID,
		BookingDateTime: dateTime,
		CustomerName:    name,
		CustomerPhone:   phone,
	}

	err = t.bookingsRepository.SaveBooking(&booking)
	return makeResult("ok", "Failed to save booking", err), nil
}

func GetAgentTools(bookingsRepository repository.BookingRepository, servicesRepository repository.ServiceRepository, employeeRepository repository.EmployeeRepository, debug bool) []langchaintools.Tool {

	logFunc := func(format string, v ...interface{}) {
		if debug {
			log.Printf(format, v...)
		}
	}

	return []langchaintools.Tool{
		&getServicesTool{
			servicesRepository: servicesRepository,
			logFunc:            logFunc,
		},
		&getEmployeesTool{
			employeesRepository: employeeRepository,
			logFunc:             logFunc,
		},
		&getServicesForEmployeeTool{
			employeesRepository: employeeRepository,
			logFunc:             logFunc,
		},
		&getEmployeesForServiceTool{
			employeesRepository: employeeRepository,
			servicesRepository:  servicesRepository,
			logFunc:             logFunc,
		},
		&checkAvailabilityTool{
			employeesRepository: employeeRepository,
			servicesRepository:  servicesRepository,
			logFunc:             logFunc,
		},
		&bookAppointmentTool{
			employeesRepository: employeeRepository,
			servicesRepository:  servicesRepository,
			bookingsRepository:  bookingsRepository,
			logFunc:             logFunc,
		},
	}
}
