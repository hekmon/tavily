package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	SystemPrompt = `You are a helpful assistant.
Your primary goal is to answer user queries to the best of your capacities, focusing on providing accurate, relevant, and useful information.
If you don't know or if the user query requires up to date informations, use the provided tool to search the web.
If you do use the tool result, always try to link back the result URL if available to back your claims.
Never include data from web search result that is not directly relevant to the query.
If there is no relevant data in the search results, simply state it clearly and concisely.
Engage in a conversational manner, asking follow-up questions to clarify or deepen the discussion.
Follow ethical guidelines, ensuring that your responses are not harmful, misleading, or biased.
If you are uncertain about a search result or lack sufficient information from the user to perform a web search, clearly state this and suggest ways to find more accurate information.`
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
							"Determines the format of the search results. The default is \"%s\" to get a unique string with a summarized result. When detailed results or multiples sources are needed use \"%s\": each result will be return as XML in a ranked order with scores and original URLs. Always pay attention to the results' score to determine the relevancy of its content against other results. A XML result object will have the following format: <result><title></title><url></url><score></score><short_description></short_description></result>",
							OpenAISearchToolParamResultFormatValueSummary, OpenAISearchToolParamResultFormatValueRanked,
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
				Text: openai.F(fmt.Sprintf("<result><title>%s</title><url>%s</url><score>%f</score><short_description>%s</short_description></result>",
					result.Title, result.URL, result.Score, strings.Replace(result.Content, "\n", "\\n", -1),
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
