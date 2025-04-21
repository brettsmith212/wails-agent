package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"

	"wails-agent/agent"
	"wails-agent/logger"
	"wails-agent/tools"
)

// App struct
type App struct {
	ctx        context.Context
	agent      *agent.Agent
	client     anthropic.Client // Store directly, not as pointer
	messages   []Message
	messageCh  chan string
	responseCh chan string
	mutex      sync.Mutex
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Initialize logger
	logDir := filepath.Join(".", "logs")
	if err := logger.Initialize(logDir); err != nil {
		log.Println("Failed to initialize logger:", err)
	}

	// Create message channels
	messageCh := make(chan string)
	responseCh := make(chan string)

	// Create Anthropic client
	client := anthropic.NewClient()

	// We'll create the agent in startup instead

	return &App{
		client:     client, // Store the client directly, not as pointer
		messageCh:  messageCh,
		responseCh: responseCh,
		messages:   []Message{},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Create a pointer to our client for the agent
	clientPtr := &a.client
	
	// Initialize agent
	a.agent = agent.NewAgent(clientPtr, func() (string, bool) {
		msg, ok := <-a.messageCh
		return msg, ok
	}, []tools.ToolDefinition{
		tools.ReadFileDefinition,
		tools.ListFilesDefinition,
		tools.EditFileDefinition,
		tools.RunShellCommandDefinition,
	})

	// Since SetResponseHandler doesn't exist, we need to monkey patch the Run method
	// Start agent in a modified goroutine that captures responses
	go func() {
		if err := a.customAgentRun(ctx); err != nil {
			log.Printf("Agent error: %s\n", err.Error())
		}
	}()
}

// customAgentRun is our implementation of agent.Run with response capture
func (a *App) customAgentRun(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	readUserInput := true
	for {
		if readUserInput {
			// Get user input from the channel
			userInput, ok := a.agent.GetUserMessage()
			if !ok {
				break
			}

			// Log the user message
			logger.LogMessage("User", userInput)
			
			// Create a user message and add it to the conversation
			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		// Run inference to get a response from the AI
		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		// Process the response from the AI
		toolResults := []anthropic.ContentBlockParamUnion{}
		responseText := ""
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				// Capture text responses
				responseText += content.Text
				logger.LogMessage("Claude", content.Text)
			case "tool_use":
				// Execute tools
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		
		// Send the response back through our channel if there was text
		if responseText != "" {
			a.responseCh <- responseText
		}
		
		// Continue the conversation based on tool results
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

// runInference runs the AI model with the given conversation
func (a *App) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	// Convert our tools to Anthropic's format
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.agent.Tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}
	
	// Call the Anthropic API
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})

	return message, err
}

// executeTool executes a tool and returns the result
func (a *App) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef tools.ToolDefinition
	var found bool
	for _, tool := range a.agent.Tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	
	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}
	
	// Log tool execution
	log.Printf("Executing tool: %s\n", name)
	
	// Execute the tool
	response, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	return anthropic.NewToolResultBlock(id, response, false)
}

// SendMessage sends a message to the agent and returns the response
func (a *App) SendMessage(message string) string {
	a.mutex.Lock()
	a.messages = append(a.messages, Message{Role: "user", Content: message})
	a.mutex.Unlock()

	// Send message to agent
	a.messageCh <- message

	// Wait for response
	response := <-a.responseCh

	a.mutex.Lock()
	a.messages = append(a.messages, Message{Role: "assistant", Content: response})
	a.mutex.Unlock()

	return response
}

// GetMessages returns all messages in the conversation
func (a *App) GetMessages() []Message {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Return a copy of the messages
	msgCopy := make([]Message, len(a.messages))
	copy(msgCopy, a.messages)
	return msgCopy
}

// Greet returns a greeting for the given name (kept for compatibility)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
