package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hekmon/tavily"
	tavilytools "github.com/hekmon/tavily/tools"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	OAIClient  *openai.Client
	TavilyTool tavilytools.OpenAISearchTool
)

func main() {
	// Init clients
	OAIClient = openai.NewClient(
		option.WithAPIKey(llmKey),
		option.WithBaseURL(baseURL),
	)
	TavilyTool.TavilyClient = tavily.NewClient(tavilyKey, nil)

	// Use them
	err := startConversation("What is Tavily? What do they offer? Be specific and do not omit anything. Did they announce something recently?")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func startConversation(question string) (err error) {
	// Prepare
	ctx := context.TODO()
	fmt.Println("Question:", question)
	fmt.Println("---------8<----------")
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(tavilytools.SystemPrompt),
		openai.UserMessage(question),
	}
	var chatCompletion *openai.ChatCompletion
	// Start conversation
	for {
		// Send messages
		if chatCompletion, err = newChatCompletion(ctx, OAIClient, messages); err != nil {
			return fmt.Errorf("failed to create a new chat completion: %w", err)
		}
		// Handle response
		if len(chatCompletion.Choices) != 1 {
			return fmt.Errorf("unexpected number of choices: %d", len(chatCompletion.Choices))
		}
		response := chatCompletion.Choices[0]
		messages = append(messages, response.Message)
		// Act based on response
		switch response.FinishReason {
		case openai.ChatCompletionChoicesFinishReasonToolCalls:
			fmt.Println("Received", len(response.Message.ToolCalls), "tool call request(s)")
			// TODO: parallelize to speed up time to first token for the user
			for _, tool := range response.Message.ToolCalls {
				fmt.Println("tool call:", tool.Function.Name, tool.Function.Arguments, tool.ID)
				switch tool.Function.Name {
				case tavilytools.OpenAISearchToolName:
					msg, err := TavilyTool.ActivateTool(ctx, tool.ID, tool.Function.Arguments)
					if err != nil {
						return fmt.Errorf("failed to activate Tavily OpenAISearchTool: %w", err)
					}
					messages = append(messages, msg)
					fmt.Println("tavily answer:", msg.Content)
					fmt.Println()
				default:
					return fmt.Errorf("failed to handle OpenAISearchTool: %w", err)
				}
			}
		case openai.ChatCompletionChoicesFinishReasonStop:
			fmt.Println("---------8<----------")
			fmt.Println("Response:")
			fmt.Println(response.Message.Content)
			return
		default:
			return fmt.Errorf("unexpected finish reason: %s", response.FinishReason)
		}
	}
}

func newChatCompletion(ctx context.Context, client *openai.Client, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	return client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Model:       openai.F("Qwen2.5-72B"),
			Messages:    openai.F(messages),
			Tools:       openai.F(availableTools()),
			N:           openai.F(int64(1)),
			Temperature: openai.F(0.7), // recommended by Qwen2.5
			TopP:        openai.F(0.8), // recommended by Qwen2.5
		},
		option.WithJSONSet("repetition_penalty", strconv.FormatFloat(1.05, 'f', -1, 64)), // recommended by Qwen2.5
	)
}

func availableTools() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		TavilyTool.GetToolParam(),
	}
}
