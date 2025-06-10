package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ToolDefinition struct {
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam   `json:"input_schema"`
	Function    func(app *App, input json.RawMessage) (string, error)
}

type AIAgent struct {
	client       *anthropic.Client
	tools        []ToolDefinition
	conversation []anthropic.MessageParam
	app          *App // Reference to the main App
	ctx          context.Context // For emitting events
}

func NewAIAgent(app *App) *AIAgent {
	client := anthropic.NewClient()
	tools := []ToolDefinition{
		ListSlidesDefinition, 
		ReadSlideDefinition, 
		EditSlideTextDefinition, 
		ExportSlidesDefinition, 
		AddSlideDefinition, 
		DeleteSlideDefinition,
	}
	
	return &AIAgent{
		client:       &client,
		tools:        tools,
		conversation: []anthropic.MessageParam{},
		app:          app,
		ctx:          nil, // Will be set when SendMessage is called
	}
}

func (a *AIAgent) SendMessage(ctx context.Context, userMessage string) error {
	a.ctx = ctx // Store context for event emission
	
	// Log user message
	a.logToFile("USER", userMessage, "")
	
	// Enhance user message with current presentation context
	enhancedMessage := userMessage
	if a.app != nil && a.app.currentPresentationPath != "" {
		enhancedMessage = fmt.Sprintf("Current presentation loaded: %s\n\nUser request: %s", a.app.currentPresentationPath, userMessage)
	}
	
	// Add user message to conversation
	userMsgParam := anthropic.NewUserMessage(anthropic.NewTextBlock(enhancedMessage))
	a.conversation = append(a.conversation, userMsgParam)

	// Run inference
	message, err := a.runInference(context.Background(), a.conversation)
	if err != nil {
		a.logToFile("ERROR", "AI inference failed", err.Error())
		return err
	}
	a.conversation = append(a.conversation, message.ToParam())

	// Process tool results in a loop until no more tool calls
	currentMessage := message
	
	for {
		toolResults := []anthropic.ContentBlockParamUnion{}
		
		// Process current message content
		for _, content := range currentMessage.Content {
			switch content.Type {
			case "text":
				// Emit text content as event
				if content.Text != "" {
					a.emitMessage(content.Text)
				}
			case "tool_use":
				// Emit tool execution status as event
				statusMsg := getToolDisplayName(content.Name)
				a.emitMessage(fmt.Sprintf("*%s...*", statusMsg))
				
				a.logToFile("TOOL_CALL", fmt.Sprintf("Tool: %s", content.Name), string(content.Input))
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		
		// If no tool calls were made, we're done
		if len(toolResults) == 0 {
			break
		}
		
		// Send tool results and get next response
		a.logToFile("DEBUG", fmt.Sprintf("Running inference with %d tool results", len(toolResults)), "")
		a.conversation = append(a.conversation, anthropic.NewUserMessage(toolResults...))
		
		nextMessage, err := a.runInference(context.Background(), a.conversation)
		if err != nil {
			a.logToFile("ERROR", "Follow-up inference failed", err.Error())
			return err
		}
		a.logToFile("DEBUG", "Follow-up inference completed successfully", "")
		a.conversation = append(a.conversation, nextMessage.ToParam())
		
		// Set up for next iteration
		currentMessage = nextMessage
	}

	return nil
}

func (a *AIAgent) emitMessage(message string) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "ai-message", message)
		// Also log for debugging
		a.logToFile("ASSISTANT", message, "")
	}
}

func getToolDisplayName(toolName string) string {
	switch toolName {
	case "list_slides":
		return "📋 Listing slides"
	case "read_slide":
		return "👀 Reading slide content"
	case "edit_slide_text":
		return "✏️ Editing slide text"
	case "export_slides":
		return "📸 Exporting slides"
	case "add_slide":
		return "➕ Adding new slide"
	case "delete_slide":
		return "🗑️ Deleting slide"
	default:
		return fmt.Sprintf("🔧 Executing %s", toolName)
	}
}

func (a *AIAgent) logToFile(msgType, message, details string) {
	// Create slides directory if it doesn't exist
	os.MkdirAll("slides", 0755)
	
	// Open log file for appending
	logPath := filepath.Join("slides", "ai_conversation.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer file.Close()
	
	// Write log entry
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, msgType, message)
	if details != "" {
		logEntry += fmt.Sprintf("Details: %s\n", details)
	}
	logEntry += "---\n"
	
	file.WriteString(logEntry)
}

func (a *AIAgent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(2048),
		Messages:  conversation,
		Tools:     anthropicTools,
	})
	return message, err
}

func (a *AIAgent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		a.logToFile("TOOL_ERROR", fmt.Sprintf("Tool not found: %s", name), "")
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	// Log current presentation path for debugging
	currentPath := "none"
	if a.app != nil && a.app.currentPresentationPath != "" {
		currentPath = a.app.currentPresentationPath
	}
	a.logToFile("TOOL_DEBUG", fmt.Sprintf("Executing %s with current presentation: %s", name, currentPath), string(input))

	fmt.Printf("Executing tool: %s(%s)\n", name, input)
	response, err := toolDef.Function(a.app, input)
	if err != nil {
		a.logToFile("TOOL_ERROR", fmt.Sprintf("Tool %s failed", name), err.Error())
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	
	a.logToFile("TOOL_RESULT", fmt.Sprintf("Tool %s completed", name), response)
	return anthropic.NewToolResultBlock(id, response, false)
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}
