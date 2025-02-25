package main

import (
	"context"
	"fmt"
	"log"

	swarmgo "github.com/prathyushnallamothu/swarmgo"
	llm "github.com/prathyushnallamothu/swarmgo/llm"
)

func main() {
	client := swarmgo.NewSwarm("YOUR_OPENAI_API_KEY", llm.OpenAI)

	agent := &swarmgo.Agent{
		Name:         "Agent",
		Instructions: "You are a helpful assistant.",
		Model:        "gpt-3.5-turbo",
	}

	messages := []llm.Message{
		{Role: "user", Content: "Hello!"},
	}

	ctx := context.Background()
	response, err := client.Run(ctx, agent, messages, nil, "", false, false, 5, true)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(response.Messages[len(response.Messages)-1].Content)
}
