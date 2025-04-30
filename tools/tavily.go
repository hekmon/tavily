package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	OpenAISearchToolParamNewsDays = "news_days"

	OpenAIExtractToolName     = "tavily_web_extract"
	OpenAIExtractToolParamURL = "url"

	SystemPrompt = `You are a helpful assistant.
Your primary goal is to answer user queries to the best of your capacities, focusing on providing accurate, relevant, and useful information.

If you don't know or if the user query requires up to date informations, use the provided tools to search the web. You can use the following tools to accomplish this:
* tavily_web_search: This tool allows you to perform a web search. Use to get information you do not know.
* tavily_web_extract: This tool allows you to extract information from a specific URL.

Use these tools wisely and only when necessary. Always provide a clear explanation of why you are using a tool and what you are looking for. Here are the rules if you use any of theses tools:
* tavily_web_search
    * Only choose news category if the user has explicitly specified a recent time frame (eg "recently", "in the past few days", "last week", etc...). Otherwise, default to general.
    * Provide a summary of the search results and ALWAYS include the source URL of the data you used to form your answer. NEVER states something without a backing URL sent to the user alongside the answer. Use markdown to format the URL within your answer. Eg:
	    * A [New York Times article](https://www.nytimes.com/article) states that...
		* The [Wikipedia page](https://en.wikipedia.org/wiki/Topic) provides detailed information on...
	* If the search results are not relevant or do not answer the query, state it clearly to the user and suggest alternative approaches.
	* If you need more information from the user to perform a search or extraction, ask for it politely and clearly.
	* Never include data from web search result that is not directly relevant to the query.
	* If there is no relevant data in the search results, simply state it clearly and concisely.
* tavily_web_extract
    * Ensure the URL is correct and relevant to the query.

Engage in a conversational manner, asking follow-up questions to clarify or deepen the discussion.
Follow ethical guidelines, ensuring that your responses are not harmful, misleading, or biased.
If you are uncertain about a search result or lack sufficient information from the user to perform a web search, clearly state it and suggest ways to find more accurate information.
Always answer in the language of the user's query.
`
)

const (
	defaultSearchMaxResults = 5
	defaultSearchNewsDays   = 7
)

func NewTavilyToolsHelper(client tavily.Client) *OpenAITavilyToolsHelper {
	return &OpenAITavilyToolsHelper{
		client:     client,
		MaxResults: defaultSearchMaxResults,
	}
}

type OpenAITavilyToolsHelper struct {
	client     tavily.Client
	MaxResults int
}

/*
	Search
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
							"The category for the search. Use %q to search for recent news articles. Use the %q parameter to specify the number of days to look back when searching for news articles. The default category is %q, which searches across general topics.",
							tavily.SearchQueryTopicNews, OpenAISearchToolParamNewsDays, tavily.SearchQueryTopicGeneral,
						),
						"enum": []string{string(tavily.SearchQueryTopicGeneral), string(tavily.SearchQueryTopicNews)},
					},
					OpenAISearchToolParamNewsDays: map[string]any{
						"type": "string",
						"description": fmt.Sprintf(
							"The number of days to look back when searching for news articles. This parameter is only used if the category is set to %q. Default is %d.",
							string(tavily.SearchQueryTopicNews), defaultSearchNewsDays,
						),
					},
				},
				"required": []string{OpenAISearchToolParamQuery},
			},
		},
	}
}

func (oaitth OpenAITavilyToolsHelper) Search(ctx context.Context, toolCallID, params string) (toolResultMsg openai.ChatCompletionToolMessageParam, err error) {
	// First parse the parameters
	parsedParams := make(map[string]string, 2)
	if err = json.Unmarshal([]byte(params), &parsedParams); err != nil {
		err = fmt.Errorf("failed to parse parameters: %w", err)
		return
	}
	var newsDays int
	if _, ok := parsedParams[OpenAISearchToolParamCategory]; ok && parsedParams[OpenAISearchToolParamCategory] == string(tavily.SearchQueryTopicNews) {
		if value, ok := parsedParams[OpenAISearchToolParamNewsDays]; ok {
			if newsDays, err = strconv.Atoi(value); err != nil {
				err = fmt.Errorf("failed to convert news_days parameter to integer: %w", err)
				return
			}
		} else {
			newsDays = defaultSearchNewsDays
		}
	}
	// Execute the search
	resp, err := oaitth.client.Search(ctx, tavily.SearchQuery{
		Query:       parsedParams[OpenAISearchToolParamQuery],
		SearchDepth: tavily.SearchQueryDepthAdvanced, // to have a meaningfull content, Advanced is required. Another solution is to query Basic with raw content but this will consome way more tokens.
		Topic:       tavily.SearchQueryTopic(parsedParams[OpenAISearchToolParamCategory]),
		Days:        newsDays,
		MaxResults:  oaitth.MaxResults,
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
				"properties": map[string]any{
					OpenAIExtractToolParamURL: map[string]any{
						"type":        "string",
						"description": "The URL you want to extract content from",
					},
				},
				"required": []string{OpenAIExtractToolParamURL},
			},
		},
	}
}

func (oaitth OpenAITavilyToolsHelper) Extract(ctx context.Context, toolCallID, params string) (toolResultMsg openai.ChatCompletionToolMessageParam, err error) {
	// First parse the parameters
	parsedParams := make(map[string]string, 1)
	if err = json.Unmarshal([]byte(params), &parsedParams); err != nil {
		err = fmt.Errorf("failed to parse parameters: %w", err)
		return
	}
	// Extract
	resp, err := oaitth.client.Extract(ctx, tavily.ExtractRequest{
		URLs:         []string{parsedParams[OpenAIExtractToolParamURL]},
		ExtractDepth: tavily.ExtractRequestDepthAdvanced,
	})
	if err != nil {
		err = fmt.Errorf("failed to perform tavily extract: %w", err)
		return
	}
	// Return the answer
	results := make([]openai.ChatCompletionContentPartTextParam, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = openai.ChatCompletionContentPartTextParam{
			Text: result.RawContent,
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
