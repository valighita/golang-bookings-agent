package customswarmgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/prathyushnallamothu/swarmgo"
	"github.com/prathyushnallamothu/swarmgo/llm"
)

// Swarm represents the main structure
type Swarm struct {
	client llm.LLM
}

// NewSwarm initializes a new Swarm instance with an LLM client
func NewSwarm(apiKey string, provider llm.LLMProvider) *Swarm {
	if provider == llm.OpenAI {
		client := llm.NewOpenAILLM(apiKey)
		return &Swarm{
			client: client,
		}
	}
	if provider == llm.Gemini {
		client, err := llm.NewGeminiLLM(apiKey)
		if err != nil {
			log.Fatalf("Failed to create Gemini client: %v", err)
		}
		return &Swarm{
			client: client,
		}
	}
	if provider == llm.Claude {
		client := llm.NewClaudeLLM(apiKey)

		return &Swarm{
			client: client,
		}
	}
	if provider == llm.Ollama {
		client, err := llm.NewOllamaLLM()
		if err != nil {
			log.Fatalf("Failed to create Ollama client: %v", err)
		}
		return &Swarm{
			client: client,
		}
	}
	if provider == llm.DeepSeek {
		client := llm.NewDeepSeekLLM(apiKey)
		return &Swarm{
			client: client,
		}
	}
	return nil
}

// getChatCompletion requests a chat completion from the LLM
func (s *Swarm) getChatCompletion(
	ctx context.Context,
	agent *swarmgo.Agent,
	history []llm.Message,
	contextVariables map[string]interface{},
	modelOverride string,
	stream bool,
	debug bool,
) (llm.ChatCompletionResponse, error) {
	// Prepare the initial system message with agent instructions
	instructions := agent.Instructions
	if agent.InstructionsFunc != nil {
		instructions = agent.InstructionsFunc(contextVariables)
	}
	messages := append([]llm.Message{
		{
			Role:    llm.RoleSystem,
			Content: instructions,
		},
	}, history...)

	// Build tool definitions from agent's functions
	var tools []llm.Tool
	for _, af := range agent.Functions {
		def := swarmgo.FunctionToDefinition(af)
		tools = append(tools, llm.Tool{
			Type: "function",
			Function: &llm.Function{
				Name:        def.Name,
				Description: def.Description,
				Parameters:  def.Parameters,
			},
		})
	}

	// Prepare the chat completion request
	model := agent.Model
	if modelOverride != "" {
		model = modelOverride
	}

	req := llm.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
	}

	if debug {
		log.Printf("Getting chat completion for: %+v\n", messages)
	}

	// Call the LLM to get a chat completion
	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return llm.ChatCompletionResponse{}, err
	}

	return resp, nil
}

// handleToolCall processes a tool call from the chat completion
func (s *Swarm) handleToolCall(
	ctx context.Context,
	toolCall *llm.ToolCall,
	agent *swarmgo.Agent,
	contextVariables map[string]interface{},
	debug bool,
) (swarmgo.Response, error) {
	toolName := toolCall.Function.Name
	argsJSON := toolCall.Function.Arguments

	// Parse the tool call arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return swarmgo.Response{}, err
	}

	if debug {
		log.Printf("Processing tool call: %s with arguments %v\n", toolName, args)
	}

	// Find the corresponding function in the agent's functions
	var functionFound *swarmgo.AgentFunction
	for _, af := range agent.Functions {
		if af.Name == toolName {
			functionFound = &af
			break
		}
	}

	// Handle case where function is not found
	if functionFound == nil {
		errorMessage := fmt.Sprintf("Error: Tool %s not found.", toolName)
		if debug {
			log.Println(errorMessage)
		}
		return swarmgo.Response{
			Messages: []llm.Message{
				{
					Role:    llm.RoleAssistant,
					Content: errorMessage,
				},
			},
		}, nil
	}

	// Execute the function
	result := functionFound.Function(args, contextVariables)

	// Create a message with the tool result
	toolResultMessage := llm.Message{
		Role:    llm.RoleAssistant,
		Content: fmt.Sprintf("%v", result.Data),
	}

	// Return the partial response with the tool result and any agent transfer
	partialResponse := swarmgo.Response{
		Messages:         []llm.Message{toolResultMessage},
		Agent:            result.Agent, // Use the agent from the result if provided
		ContextVariables: contextVariables,
	}

	return partialResponse, nil
}

// Run executes the chat interaction loop with the agent
func (s *Swarm) Run(
	ctx context.Context,
	agent *swarmgo.Agent,
	messages []llm.Message,
	contextVariables map[string]interface{},
	modelOverride string,
	stream bool,
	debug bool,
	maxTurns int,
	executeTools bool,
) (swarmgo.Response, error) {
	activeAgent := agent
	history := make([]llm.Message, len(messages))
	copy(history, messages)
	if contextVariables == nil {
		contextVariables = make(map[string]interface{})
	}

	// Initialize memory if not already initialized
	if activeAgent.Memory == nil {
		activeAgent.Memory = swarmgo.NewMemoryStore(100)
	}

	initLen := len(messages)
	turns := 0

	// Store initial user message as memory if it exists
	if len(messages) > 0 && messages[len(messages)-1].Role == llm.RoleUser {
		activeAgent.Memory.AddMemory(swarmgo.Memory{
			Content:   messages[len(messages)-1].Content,
			Timestamp: time.Now(),
		})
	}

	for turns < maxTurns {
		// Get chat completion from LLM
		resp, err := s.getChatCompletion(ctx, activeAgent, history, contextVariables, modelOverride, stream, debug)
		if err != nil {
			return swarmgo.Response{}, err
		}

		// Process the response
		if len(resp.Choices) == 0 {
			return swarmgo.Response{}, fmt.Errorf("no choices in response")
		}

		choice := resp.Choices[0]

		// Check for tool calls
		if len(choice.Message.ToolCalls) > 0 && executeTools {
			var toolResponses []swarmgo.Response
			var toolResults []swarmgo.ToolResult
			// Add the assistant's message with tool calls
			history = append(history, choice.Message)

			for _, toolCall := range choice.Message.ToolCalls {
				toolResp, err := s.handleToolCall(ctx, &toolCall, activeAgent, contextVariables, debug)
				if err != nil {
					return swarmgo.Response{}, err
				}
				toolResponses = append(toolResponses, toolResp)

				// Create ToolResult entry
				var args interface{}
				_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				toolResults = append(toolResults, swarmgo.ToolResult{
					ToolName: toolCall.Function.Name,
					Args:     args,
					Result: swarmgo.Result{
						Success: true,
						Data:    toolResp.Messages[0].Content,
						Error:   nil,
						Agent:   toolResp.Agent,
					},
				})

				// Add the tool response as a function message
				history = append(history, llm.Message{
					Role:       llm.RoleTool,
					Content:    toolResp.Messages[0].Content,
					Name:       toolCall.Function.Name,
					ToolCallId: toolCall.ID,
				})
				// Update the active agent if the tool result includes an agent transfer
				if toolResp.Agent != nil {
					activeAgent = toolResp.Agent
				}
			}
			turns++

			// Get a follow-up response from the AI after tool execution
			continue
		} else {
			// Add the assistant's message to history
			history = append(history, choice.Message)

			// Return final response only if there are no tool calls
			finalResponse := swarmgo.Response{
				Messages:         history[initLen:],
				Agent:            activeAgent,
				ContextVariables: contextVariables,
				ToolResults:      nil, // No tool calls were made
			}
			return finalResponse, nil
		}
	}

	// Return response with all messages including the follow-up
	return swarmgo.Response{
		Messages:         history[initLen:],
		Agent:            activeAgent,
		ContextVariables: contextVariables,
		ToolResults:      nil,
	}, nil
}
