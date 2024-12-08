# OpenAI Tool helper

This package can help you integrate the Tavily client as an OpenAI tool.

## Exemple

### Result

```
Question: What is Tavily? What do they offer? Be specific and do not omit anything. Did they announce something recently?
---------8<----------
tool call: tavily_web_search {"query": "What is Tavily", "category": "general", "depth": "advanced"} chatcmpl-tool-45993a6a234c413a926b8f6e0719bd7a
tavily answer: [{Tavily is an AI-powered search engine and research tool specifically designed for AI agents, particularly Large Language Models (LLMs). It provides fast, accurate, and comprehensive results from trusted sources, optimized for Retrieval-Augmented Generation (RAG) purposes. Tavily offers a specialized Search API that developers can integrate into their applications, allowing them to perform complex queries and retrieve information efficiently. It enhances AI capabilities by aggregating information from multiple sources and is recognized globally by AI leaders for its effectiveness in powering AI applications and autonomous agents. text}]

tool call: tavily_web_search {"query": "Tavily recent announcements", "category": "news", "depth": "basic"} chatcmpl-tool-9b6bbef355d44be1a00d134bf570fac1
tavily answer: [{Recent announcements from Tavily are not available in the provided data sources. The sources focus on various companies such as Novo Nordisk, Amazon, and Oura but do not include specific information regarding Tavily. text}]

---------8<----------
Response:
Tavily is an AI-powered search engine and research tool specifically designed for AI agents, particularly Large Language Models (LLMs). It provides fast, accurate, and comprehensive results from trusted sources, optimized for Retrieval-Augmented Generation (RAG) purposes. Tavily offers a specialized Search API that developers can integrate into their applications, allowing them to perform complex queries and retrieve information efficiently. This tool enhances AI capabilities by aggregating information from multiple sources and is recognized globally by AI leaders for its effectiveness in powering AI applications and autonomous agents.

Regarding recent announcements, I couldn't find any specific news about Tavily in the available data sources. If you have a particular interest in their latest developments, it might be helpful to check their official website or social media channels for the most up-to-date information.
```

### Code

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hekmon/tavily"
	tavilytools "github.com/hekmon/tavily/tools"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	OAIClient    *openai.Client
	TavilyClient *tavily.Client
	TavilyTool   tavilytools.OpenAISearchTool
)

func main() {

	OAIClient = openai.NewClient(
		option.WithAPIKey(llmKey),
		option.WithBaseURL(baseURL),
	)

	TavilyClient = tavily.NewClient(tavilyKey, nil)
	TavilyTool.TavilyClient = TavilyClient

	err := startConversation()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func startConversation() (err error) {
	// Prepare
	ctx := context.TODO()
	question := "What is Tavily? What do they offer? Be specific and do not omit anything. Did they announce something recently?"
	fmt.Println("Question:", question)
	fmt.Println("---------8<----------")
	messages := []openai.ChatCompletionMessageParamUnion{
		// openai.SystemMessage(systemPrompt),
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
		case openai.ChatCompletionChoicesFinishReasonStop:
			fmt.Println("---------8<----------")
			fmt.Println("Response:")
			fmt.Println(response.Message.Content)
			return
		case openai.ChatCompletionChoicesFinishReasonToolCalls:
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
		default:
			return fmt.Errorf("unexpected finish reason: %s", chatCompletion.Choices[0].FinishReason)
		}
	}
}

func newChatCompletion(ctx context.Context, client *openai.Client, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	return client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Model:       openai.F(model),
			Messages:    openai.F(messages),
			Tools:       openai.F(availableTools()),
			N:           openai.F(int64(1)),
			Temperature: openai.F(temperature),
			TopP:        openai.F(topP),
		},
		option.WithJSONSet("repetition_penalty", strconv.FormatFloat(repetitionPenalty, 'f', -1, 64)), // extra param for Qwen2.5
	)
}

func availableTools() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		TavilyTool.GetToolParam(),
	}
}
```
