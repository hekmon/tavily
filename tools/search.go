package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hekmon/tavily"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

const (
	OpenAISearchToolName          = "tavily_web_search"
	OpenAISearchToolParamQuery    = "query"
	OpenAISearchToolParamCategory = "category"
	OpenAISearchToolParamDepth    = "depth"
)

type OpenAISearchTool struct {
	TavilyClient *tavily.Client
}

func (oaist OpenAISearchTool) GetToolParam() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(shared.FunctionDefinitionParam{
			Name:        openai.F(OpenAISearchToolName),
			Description: openai.F("Search the web with a query and get a live answer"),
			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					OpenAISearchToolParamQuery: map[string]interface{}{
						"type":        "string",
						"description": "The query you want to search for on the web",
					},
					OpenAISearchToolParamCategory: map[string]interface{}{
						"type": "string",
						"description": fmt.Sprintf(
							"The category of the search. Can be set to \"%s\" to perform a search solely on news articles. Default to \"%s\"",
							tavily.SearchQueryTopicNews, tavily.SearchQueryTopicGeneral,
						),
						"enum": []string{string(tavily.SearchQueryTopicGeneral), string(tavily.SearchQueryTopicNews)},
					},
					OpenAISearchToolParamDepth: map[string]interface{}{
						"type": "string",
						"description": fmt.Sprintf(
							"The level of depth your want your search to have. Use \"%s\" for better results if your query is complex. Default to \"%s\".",
							tavily.SearchQueryDepthAdvanced, tavily.SearchQueryDepthBasic,
						),
						"enum": []string{string(tavily.SearchQueryDepthBasic), string(tavily.SearchQueryDepthAdvanced)},
					},
				},
				"required": []string{OpenAISearchToolParamQuery},
			}),
		}),
	}
}

func (oaist OpenAISearchTool) ActivateTool(ctx context.Context, toolCallID, params string) (toolResultMsg openai.ChatCompletionToolMessageParam, err error) {
	// First parse the parameters
	parsedParams := make(map[string]string, 2)
	if err = json.Unmarshal([]byte(params), &parsedParams); err != nil {
		err = fmt.Errorf("failed to parse parameters: %w", err)
		return
	}
	var newsDays int
	if _, ok := parsedParams[OpenAISearchToolParamCategory]; ok && parsedParams[OpenAISearchToolParamCategory] == string(tavily.SearchQueryTopicNews) {
		newsDays = 7
	}
	// Execute the search
	resp, err := oaist.TavilyClient.Search(ctx, tavily.SearchQuery{
		Query:         parsedParams[OpenAISearchToolParamQuery],
		SearchDepth:   tavily.SearchQueryDepth(parsedParams[OpenAISearchToolParamDepth]),
		Topic:         tavily.SearchQueryTopic(parsedParams[OpenAISearchToolParamCategory]),
		Days:          newsDays,
		MaxResults:    tavily.SearchMaxPossibleResults,
		IncludeAnswer: true, // let Tavily summarize the results for us and use it as our final answer
	})
	if err != nil {
		err = fmt.Errorf("failed to perform tavily search: %w", err)
		return
	}
	// Return the answer
	toolResultMsg = openai.ChatCompletionToolMessageParam{
		Role: openai.F(openai.ChatCompletionToolMessageParamRoleTool),
		Content: openai.F([]openai.ChatCompletionContentPartTextParam{
			{
				Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
				Text: openai.F(*resp.Answer),
			},
		}),
		ToolCallID: openai.F(toolCallID),
	}
	return
}
