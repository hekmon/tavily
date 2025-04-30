package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hekmon/tavily"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

const (
	OpenAISearchToolName          = "tavily_web_search"
	OpenAISearchToolParamQuery    = "query"
	OpenAISearchToolParamCategory = "category"
	OpenAISearchToolParamDepth    = "depth"

	OpenAIExtractToolName       = "tavily_web_extract"
	OpenAIExtractToolParamURL   = "url"
	OpenAIExtractToolParamDepth = "depth"

	SystemPrompt = `You are a helpful assistant.
Your primary goal is to answer user queries to the best of your capacities, focusing on providing accurate, relevant, and useful information.

If you don't know or if the user query requires up to date informations, use the provided tools to search the web. You can use the following tools to accomplish this:
- tavily_web_search: This tool allows you to perform a web search. It takes a query and optionally a category and depth as parameters.
- tavily_web_extract: This tool allows you to extract information from a specific URL. It takes a URL and optionally a depth as parameters.

Use these tools wisely and only when necessary. Always provide a clear explanation of why you are using a tool and what you are looking for.
When using the tavily_web_search tool, consider the category and depth parameters to refine your search results.
When using the tavily_web_extract tool, ensure the URL is correct and relevant to the query.

Always provide a summary of the search results or extracted information, and if possible, include the source URL.
If the search results or extracted information are not relevant or do not answer the query, state this clearly and suggest alternative approaches.
If you need more information from the user to perform a search or extraction, ask for it politely and clearly.

Always aim to provide the most accurate and up-to-date information possible.
If you do use the tool result, always try to link back the result URL if available to back your claims.

Never include data from web search result that is not directly relevant to the query.
If there is no relevant data in the search results, simply state it clearly and concisely.

Engage in a conversational manner, asking follow-up questions to clarify or deepen the discussion.
Follow ethical guidelines, ensuring that your responses are not harmful, misleading, or biased.
If you are uncertain about a search result or lack sufficient information from the user to perform a web search, clearly state it and suggest ways to find more accurate information.
`
)

const (
	defaultSearchMaxResults = 5
	defaultSearchNewsDays   = 7
)

func NewTavilyToolsHelper(client tavily.Client) *OpenAITavilyToolsHelper {
	return &OpenAITavilyToolsHelper{
		client:         client,
		MaxResults:     defaultSearchMaxResults,
		NewsSearchDays: defaultSearchNewsDays,
	}
}

type OpenAITavilyToolsHelper struct {
	client         tavily.Client
	MaxResults     int
	NewsSearchDays int
}

/*
	Search
*/

func (oaitth OpenAITavilyToolsHelper) Search(ctx context.Context, toolCallID, params string) (toolResultMsg openai.ChatCompletionToolMessageParam, err error) {
	// First parse the parameters
	parsedParams := make(map[string]string, 2)
	if err = json.Unmarshal([]byte(params), &parsedParams); err != nil {
		err = fmt.Errorf("failed to parse parameters: %w", err)
		return
	}
	var newsDays int
	if _, ok := parsedParams[OpenAISearchToolParamCategory]; ok && parsedParams[OpenAISearchToolParamCategory] == string(tavily.SearchQueryTopicNews) {
		newsDays = oaitth.NewsSearchDays
	}
	// Execute the search
	resp, err := oaitth.client.Search(ctx, tavily.SearchQuery{
		Query:       parsedParams[OpenAISearchToolParamQuery],
		SearchDepth: tavily.SearchQueryDepthAdvanced, // to have a meaningfull content, Advanced is required. Another solution is to query Basic with raw content but this will consome way more tokens.
		Topic:       tavily.SearchQueryTopic(parsedParams[OpenAISearchToolParamCategory]),
		Days:        newsDays,
		// MaxResults:    tavily.SearchMaxPossibleResults,
	})
	if err != nil {
		err = fmt.Errorf("failed to perform tavily search: %w", err)
		return
	}
	// Return the answer
	results := make([]openai.ChatCompletionContentPartTextParam, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = openai.ChatCompletionContentPartTextParam{
			Text: fmt.Sprintf("<result><title>%s</title><url>%s</url><score>%f</score><content>%s</content></result>",
				result.Title, result.URL, result.Score, strings.Replace(result.Content, "\n", " ", -1),
			),
		}
	}
	toolResultMsg = openai.ChatCompletionToolMessageParam{
		Content: openai.ChatCompletionToolMessageParamContentUnion{
			OfArrayOfContentParts: results,
		},
		ToolCallID: toolCallID,
	}
	return
}

/*
	Extract
*/

func (oaitth OpenAITavilyToolsHelper) GetSearchToolParam() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		// Type: constant.Function(""),
		Function: shared.FunctionDefinitionParam{
			Name: OpenAISearchToolName,
			Description: param.Opt[string]{
				Value: "Perform a web search using the specified query. The results will be returned in an XML format with each entry containing a title, URL, score, and the summarized content of the page. The score indicates the relevance of the result, ranging from 0.0 (least relevant) to 1.0 (most relevant).",
			},
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					OpenAISearchToolParamQuery: map[string]any{
						"type":        "string",
						"description": "The search query to be performed on the web.",
					},
					OpenAISearchToolParamCategory: map[string]any{
						"type": "string",
						"description": fmt.Sprintf(
							"The category for the search. Use %q to search for news articles from the past %d days. The default category is %q, which searches across general topics.",
							tavily.SearchQueryTopicNews, oaitth.NewsSearchDays, tavily.SearchQueryTopicGeneral,
						),
						"enum": []string{string(tavily.SearchQueryTopicGeneral), string(tavily.SearchQueryTopicNews)},
					},
				},
				"required": []string{OpenAISearchToolParamQuery},
			},
		},
	}
}

func (oaitth OpenAITavilyToolsHelper) GetExtractToolParam() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		// Type: constant.Function(""),
		Function: shared.FunctionDefinitionParam{
			Name: OpenAISearchToolName,
			Description: param.Opt[string]{
				Value: "Extract content from a given URL",
			},
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					OpenAIExtractToolParamURL: map[string]interface{}{
						"type":        "string",
						"description": "The URL you want to extract content from",
					},
					OpenAIExtractToolParamDepth: map[string]interface{}{
						"type": "string",
						"description": fmt.Sprintf(
							"The level of depth you want for your search. Use %q for better results if the query or its subject is complex. The default level is %q.",
							tavily.SearchQueryDepthAdvanced, tavily.SearchQueryDepthBasic,
						),
						"enum": []string{string(tavily.SearchQueryDepthBasic), string(tavily.SearchQueryDepthAdvanced)},
					},
				},
				"required": []string{OpenAISearchToolParamQuery},
			},
		},
	}
}
