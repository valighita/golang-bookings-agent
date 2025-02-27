package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	customswarmgo "valighita/bookings-ai-agent/agent/customswarmgo"

	"github.com/prathyushnallamothu/swarmgo"
	"github.com/prathyushnallamothu/swarmgo/llm"
)

const (
	defaultLlmModel = "gpt-4o-mini"
	defaultMaxTurns = 5
	contextPrompt   = "You are a helpful booking assistant for a dental clinic, helping clients book appointments. " +
		"The clinic has multiple employees, each performing different services with different duration and prices." +
		"You can use multiple tools.  Always use service and employee names, never ids." +
		"Bookings can be made at multiple of 15 minutes, never anything else." +
		"Clients can book appointments with one of them and they need to specify a service, a date and a time, a name and a phone number." +
		"It's important to only answer relevant questions about the services provided, do not provide information about unrelated topics." +
		"Ask the name and phone number as the final info if not already provided. Ask for confirmation before performing the final booking."
)

type AgentFactory interface {
	CreateAgent() (Agent, error)
}

type Agent interface {
	GetCompletion(message string) (string, error)
}

type agentConfig struct {
	llmModel  string
	maxTurns  int
	debugMode bool
}

type openAIAgentFactory struct {
	client      *customswarmgo.Swarm
	agentTools  []swarmgo.AgentFunction
	agentConfig *agentConfig
}

type openAIAgent struct {
	config   *agentConfig
	client   *customswarmgo.Swarm
	agent    *swarmgo.Agent
	messages []llm.Message
}

func NewOpenaiAgentFactory(agentTools []swarmgo.AgentFunction, debugMode bool) AgentFactory {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatalf("OPENAI_API_KEY is required")
	}

	client := customswarmgo.NewSwarm(openAIKey, llm.OpenAI)

	llmModel := os.Getenv("LLM_MODEL")
	if llmModel == "" {
		llmModel = defaultLlmModel
	}

	maxTurnsStr := os.Getenv("MAX_AGENT_TURNS")
	maxTurns := defaultMaxTurns
	if maxTurnsStr != "" {
		var err error
		maxTurns, err = strconv.Atoi(maxTurnsStr)
		if err != nil || maxTurns <= 0 {
			log.Fatalf("MAX_AGENT_TURNS must be a positive integer")
		}
	}

	return &openAIAgentFactory{
		client:     client,
		agentTools: agentTools,
		agentConfig: &agentConfig{
			llmModel:  llmModel,
			maxTurns:  maxTurns,
			debugMode: debugMode,
		},
	}
}

func (f *openAIAgentFactory) CreateAgent() (Agent, error) {
	agent := swarmgo.NewAgent("Booking Agent", f.agentConfig.llmModel, llm.OpenAI)
	agent.Functions = f.agentTools
	return &openAIAgent{
		client: f.client,
		agent:  agent,
		messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: contextPrompt + ". Current time is " + time.Now().Format("2006-01-02 15:04:05, Monday"),
			},
		},
		config: f.agentConfig,
	}, nil
}

func (a *openAIAgent) GetCompletion(message string) (string, error) {
	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleUser,
		Content: message,
	})

	ctx := context.Background()
	response, err := a.client.Run(ctx, a.agent, a.messages, nil, "", false, a.config.debugMode, 10, true)
	if err != nil {
		return "", err
	}

	if len(response.Messages) == 0 {
		return "", fmt.Errorf("Can't process request.")
	}

	msg := response.Messages[len(response.Messages)-1]
	a.messages = append(a.messages, msg)

	return msg.Content, nil
}
