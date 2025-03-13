package agent

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	langchaintools "github.com/tmc/langchaingo/tools"
)

const (
	defaultLlmModel = "gpt-4o-mini"
	defaultMaxTurns = 10
	contextPrompt   = "You are a helpful booking assistant for a dental clinic, helping clients book appointments. " +
		"The clinic has multiple employees, each performing different services with different duration and prices." +
		"You can use multiple tools.  Always use service and employee names, never ids." +
		"Bookings can be made at multiple of 15 minutes, never anything else." +
		"Clients can book appointments with one of them and they need to specify a service, a date and a time, a name and a phone number." +
		"It's important to only answer relevant questions about the services provided, do not provide information about unrelated topics." +
		"Ask the name and phone number as the final info if not already provided. Ask for confirmation before performing the final booking.\n\n" +
		"{{.tool_descriptions}}"
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
	llm         *openai.LLM
	agentTools  []langchaintools.Tool
	agentConfig *agentConfig
}

type openAIAgent struct {
	executor *agents.Executor
}

func NewOpenaiAgentFactory(agentTools []langchaintools.Tool, debugMode bool) AgentFactory {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatalf("OPENAI_API_KEY is required")
	}

	llmModel := os.Getenv("LLM_MODEL")
	if llmModel == "" {
		llmModel = defaultLlmModel
	}

	llm, err := openai.New(openai.WithModel(llmModel))
	if err != nil {
		log.Fatalf("Error creating OpenAI LLM: %v", err)
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
		llm:        llm,
		agentTools: agentTools,
		agentConfig: &agentConfig{
			llmModel:  llmModel,
			maxTurns:  maxTurns,
			debugMode: debugMode,
		},
	}
}

func (f *openAIAgentFactory) CreateAgent() (Agent, error) {
	memory := memory.NewConversationBuffer()

	agent := agents.NewConversationalAgent(f.llm,
		f.agentTools,
		agents.WithPromptPrefix(contextPrompt+". Current time is "+time.Now().Format("2006-01-02 15:04:05, Monday")),
		agents.WithMemory(memory),
	)

	executor := agents.NewExecutor(agent,
		agents.WithMaxIterations(f.agentConfig.maxTurns),
		agents.WithMemory(memory),
	)

	return &openAIAgent{
		executor: executor,
	}, nil
}

func (a *openAIAgent) GetCompletion(prompt string) (string, error) {
	return chains.Run(context.Background(), a.executor, prompt)
}
