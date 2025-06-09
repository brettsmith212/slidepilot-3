package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type ToolDefinition struct {
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam   `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

type AIAgent struct {
	client       *anthropic.Client
	tools        []ToolDefinition
	conversation []anthropic.MessageParam
}

func NewAIAgent() *AIAgent {
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
	}
}

func (a *AIAgent) SendMessage(userMessage string) (string, error) {
	// Add user message to conversation
	userMsgParam := anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage))
	a.conversation = append(a.conversation, userMsgParam)

	// Run inference
	message, err := a.runInference(context.Background(), a.conversation)
	if err != nil {
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
			result := a.executeTool(content.ID, content.Name, content.Input)
			toolResults = append(toolResults, result)
		}
	}

	// If there were tool calls, run another inference to get AI's response
	if len(toolResults) > 0 {
		a.conversation = append(a.conversation, anthropic.NewUserMessage(toolResults...))
		
		followUpMessage, err := a.runInference(context.Background(), a.conversation)
		if err != nil {
			return "", err
		}
		a.conversation = append(a.conversation, followUpMessage.ToParam())
		
		for _, content := range followUpMessage.Content {
			if content.Type == "text" {
				response += content.Text
			}
		}
	}

	return response, nil
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
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	fmt.Printf("Executing tool: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
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
