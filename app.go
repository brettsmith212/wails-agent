package main

import (
	"context"
	"log"
	"path/filepath"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"wails-agent/agent"
	"wails-agent/logger"
	"wails-agent/tools"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// App struct
type App struct {
	ctx      context.Context
	client   anthropic.Client
	agent    *agent.Agent
	mutex    sync.Mutex
	messages []Message
}

// NewApp creates a new App application struct
func NewApp() *App {
	logDir := filepath.Join(".", "logs")
	if err := logger.Initialize(logDir); err != nil {
		log.Println("Failed to initialize logger:", err)
	}

	client := anthropic.NewClient()
	return &App{
		client:   client,
		messages: []Message{},
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.agent = agent.NewAgent(
		&a.client,
		nil,
		[]tools.ToolDefinition{
			tools.ReadFileDefinition,
			tools.ListFilesDefinition,
			tools.EditFileDefinition,
			tools.RunShellCommandDefinition,
		},
	)
}

// SendMessage handles a complete user message, including all tool calls
func (a *App) SendMessage(userText string) (string, error) {
	// Append user message
	a.mutex.Lock()
	a.messages = append(a.messages, Message{Role: "user", Content: userText})
	a.mutex.Unlock()

	// Build initial conversation
	conv := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userText)),
	}

	var finalReply string

	// Loop until there are no tool_use blocks
	for {
		resp, err := a.client.Messages.New(a.ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude3_7SonnetLatest,
			MaxTokens: int64(1024),
			Messages:  conv,
			Tools:     a.agent.ToolsParam(),
		})
		if err != nil {
			return "", err
		}

		didTool := false
		var textBuf string

		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				textBuf += block.Text
				// Emit partial text for streaming UI
				runtime.EventsEmit(a.ctx, "streamed_message", map[string]string{
					"role":    "assistant",
					"content": block.Text,
				})
			case "tool_use":
				didTool = true
				// Execute the tool and append its result
				resultBlock := a.agent.ExecuteTool(block.ID, block.Name, block.Input)
				conv = append(conv, anthropic.NewUserMessage(resultBlock))
			}
		}

		if !didTool {
			finalReply = textBuf
			break
		}
	}

	// Save and emit completion
	a.mutex.Lock()
	a.messages = append(a.messages, Message{Role: "assistant", Content: finalReply})
	a.mutex.Unlock()
	runtime.EventsEmit(a.ctx, "message_complete", finalReply)

	return finalReply, nil
}

// GetMessages returns all messages in the conversation
func (a *App) GetMessages() []Message {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	copyMsgs := make([]Message, len(a.messages))
	copy(copyMsgs, a.messages)
	return copyMsgs
}
