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
	}
}

func (a *AIAgent) SendMessage(userMessage string) (string, error) {
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
		return "", err
	}
	a.conversation = append(a.conversation, message.ToParam())

	// Process tool results
	var response string
	toolResults := []anthropic.ContentBlockParamUnion{}
	
	for _, content := range message.Content {
		switch content.Type {
		case "text":
			response += content.Text
		case "tool_use":
			a.logToFile("TOOL_CALL", fmt.Sprintf("Tool: %s", content.Name), string(content.Input))
			result := a.executeTool(content.ID, content.Name, content.Input)
			toolResults = append(toolResults, result)
		}
	}

	// If there were tool calls, run another inference to get AI's response
	if len(toolResults) > 0 {
		a.conversation = append(a.conversation, anthropic.NewUserMessage(toolResults...))
		
		followUpMessage, err := a.runInference(context.Background(), a.conversation)
		if err != nil {
			a.logToFile("ERROR", "Follow-up inference failed", err.Error())
			return "", err
		}
		a.conversation = append(a.conversation, followUpMessage.ToParam())
		
		for _, content := range followUpMessage.Content {
			if content.Type == "text" {
				response += content.Text
			}
		}
	}

	// Log final response
	a.logToFile("ASSISTANT", response, "")

	return response, nil
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
