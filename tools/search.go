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
	OpenAISearchToolName                          = "tavily_web_search"
	OpenAISearchToolParamQuery                    = "query"
	OpenAISearchToolParamCategory                 = "category"
	OpenAISearchToolParamDepth                    = "depth"
	OpenAISearchToolParamResultFormat             = "result_format"
	OpenAISearchToolParamResultFormatValueSummary = "summary"
	OpenAISearchToolParamResultFormatValueRanked  = "ranked"
)

const (
	SystemPrompt = `You are a helpful assistant. Your primary goal is to answer user queries to the best of your capacities, focusing on providing accurate, relevant, and useful information.
If you search the web and use one of the results, always link back to the user the original URL to back your claims. Avoid including data that are not directly relevant to the query.
If there is no relevant data in the search results, simply state it clearly and concisely. Engage in a conversational manner, asking follow-up questions to clarify or deepen the discussion.
Follow ethical guidelines, ensuring that your responses are not harmful, misleading, or biased. If you are uncertain about an answer or lack sufficient information, clearly state this and suggest ways to find more accurate information.`
)

var (
	OpenAISearchToolParamCategoryNewsDays = 7
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
							"The category of the search can be set to \"%s\" to perform a search solely on news articles for the past %d days. The default category is \"%s\".",
							tavily.SearchQueryTopicNews, OpenAISearchToolParamCategoryNewsDays, tavily.SearchQueryTopicGeneral,
						),
						"enum": []string{string(tavily.SearchQueryTopicGeneral), string(tavily.SearchQueryTopicNews)},
					},
					OpenAISearchToolParamDepth: map[string]interface{}{
						"type": "string",
						"description": fmt.Sprintf(
							"The level of depth you want for your search. Use \"%s\" for better results if the query or its subject is complex. The default level is \"%s\".",
							tavily.SearchQueryDepthAdvanced, tavily.SearchQueryDepthBasic,
						),
						"enum": []string{string(tavily.SearchQueryDepthBasic), string(tavily.SearchQueryDepthAdvanced)},
					},
					OpenAISearchToolParamResultFormat: map[string]interface{}{
						"type": "string",
						"description": fmt.Sprintf(
							"Determines the format of the search results. Use \"%s\" for complex queries to get all results as XML in a ranked order with scores and original URLs. The default is \"%s\" to get a concise top result.",
							OpenAISearchToolParamResultFormatValueRanked, OpenAISearchToolParamResultFormatValueSummary,
						),
						"enum": []string{OpenAISearchToolParamResultFormatValueSummary, OpenAISearchToolParamResultFormatValueRanked},
					},
				},
				"required": []string{OpenAISearchToolParamQuery},
			}),
		}),
	}
}

func (oaist OpenAISearchTool) ActivateTool(ctx context.Context, toolCallID, params string) (toolResultMsg openai.ChatCompletionToolMessageParam, err error) {
	// First parse the parameters
	parsedParams := make(map[string]string, 4)
	if err = json.Unmarshal([]byte(params), &parsedParams); err != nil {
		err = fmt.Errorf("failed to parse parameters: %w", err)
		return
	}
	var summarize bool
	if value, ok := parsedParams[OpenAISearchToolParamResultFormat]; (ok && value == OpenAISearchToolParamResultFormatValueSummary) || !ok {
		summarize = true
	}
	var newsDays int
	if _, ok := parsedParams[OpenAISearchToolParamCategory]; ok && parsedParams[OpenAISearchToolParamCategory] == string(tavily.SearchQueryTopicNews) {
		newsDays = OpenAISearchToolParamCategoryNewsDays
	}
	// Execute the search
	resp, err := oaist.TavilyClient.Search(ctx, tavily.SearchQuery{
		Query:         parsedParams[OpenAISearchToolParamQuery],
		SearchDepth:   tavily.SearchQueryDepth(parsedParams[OpenAISearchToolParamDepth]),
		Topic:         tavily.SearchQueryTopic(parsedParams[OpenAISearchToolParamCategory]),
		Days:          newsDays,
		IncludeAnswer: summarize,
		// MaxResults:    tavily.SearchMaxPossibleResults,
	})
	if err != nil {
		err = fmt.Errorf("failed to perform tavily search: %w", err)
		return
	}
	// Return the answer
	var results []openai.ChatCompletionContentPartTextParam
	if summarize {
		//// short format
		results = []openai.ChatCompletionContentPartTextParam{
			{
				Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
				Text: openai.F(*resp.Answer),
			},
		}
	} else {
		//// long format
		results = make([]openai.ChatCompletionContentPartTextParam, len(resp.Results))
		for i, result := range resp.Results {
			results[i] = openai.ChatCompletionContentPartTextParam{
				Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
				Text: openai.F(fmt.Sprintf("<result><title>%s</title><url>%s</url><score>%f</score><content>%s</content></result>",
					result.Title, result.URL, result.Score, result.Content,
				)),
			}
		}
	}
	toolResultMsg = openai.ChatCompletionToolMessageParam{
		Role:       openai.F(openai.ChatCompletionToolMessageParamRoleTool),
		Content:    openai.F(results),
		ToolCallID: openai.F(toolCallID),
	}
	return
}
