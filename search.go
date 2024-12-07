package tavily

import (
	"context"
	"errors"
	"fmt"
)

// SearchQuery represents the parameters for a search query.
// https://docs.tavily.com/docs/rest-api/api-reference#parameters
type SearchQuery struct {
	Query                    string           `json:"query"`                                // The search query you want to execute with Tavily.
	SearchDepth              SearchQueryDepth `json:"search_depth,omitempty"`               // The depth of the search. It can be "basic" or "advanced". Default is "basic" unless specified otherwise in a given method.
	Topic                    SearchQueryTopic `json:"topic,omitempty"`                      // The category of the search. This will determine which of our agents will be used for the search. Currently: only "general" and "news" are supported. Default is "general".
	Days                     int              `json:"days,omitempty"`                       // The number of days back from the current date to include in the search results. This specifies the time frame of data to be retrieved. Please note that this feature is only available when using the "news" search topic. Default is 3.
	MaxResults               int              `json:"max_results,omitempty"`                // The maximum number of search results to return. Default is 5.
	IncludeImages            bool             `json:"include_images,omitempty"`             // Include a list of query-related images in the response. Default is False.
	IncludeImageDescriptions bool             `json:"include_image_descriptions,omitempty"` // When include_images is set to True, this option adds descriptive text for each image. Default is False.
	IncludeAnswer            bool             `json:"include_answer,omitempty"`             // Include a short answer to original query. Default is False.
	IncludeRawContent        bool             `json:"include_raw_content"`                  // Include the cleaned and parsed HTML content of each search result. Default is False.
	IncludeDomains           []string         `json:"include_domains,omitempty"`            // A list of domains to specifically include in the search results. Default is [], which includes all domains.
	ExcludeDomains           []string         `json:"exclude_domains,omitempty"`            // A list of domains to specifically exclude from the search results. Default is [], which doesn't exclude any domains.
}

func (sq SearchQuery) Validate() error {
	// Query
	if sq.Query == "" {
		return errors.New("query is required")
	}
	// Search Depth
	switch sq.SearchDepth {
	case SearchDepthBasic, SearchDepthAdvanced:
	default:
		return errors.New("invalid search depth")
	}
	// Topic
	switch sq.Topic {
	case SearchQueryTopicGeneral, SearchQueryTopicNews:
	default:
		return errors.New("invalid topic")
	}
	// Days
	switch {
	case sq.Days < 0:
		return errors.New("days must be a non-negative integer")
	case sq.Days > 0 && sq.Topic != SearchQueryTopicNews:
		return fmt.Errorf("days can only be specified when using the %q topic", SearchQueryTopicNews)
	}
	// Max Results
	if sq.MaxResults < 0 {
		return errors.New("max_results must be a non-negative integer")
	}
	// Images descriptions
	if !sq.IncludeImages && sq.IncludeImageDescriptions {
		return errors.New("include_image_descriptions can only be true when include_images is true")
	}
	return nil
}

type SearchQueryDepth string

const (
	SearchDepthBasic    SearchQueryDepth = "basic"
	SearchDepthAdvanced SearchQueryDepth = "advanced"
)

type SearchQueryTopic string

const (
	SearchQueryTopicGeneral SearchQueryTopic = "general"
	SearchQueryTopicNews    SearchQueryTopic = "news"
)

type searchQueryAuth struct {
	APIKey string `json:"api_key"`
	SearchQuery
}

func (c *Client) Search(ctx context.Context, query SearchQuery) (err error) {
	// Prepare query
	if err := query.Validate(); err != nil {
		return fmt.Errorf("failed to validate search query: %w", err)
	}
	authedQuery := searchQueryAuth{
		APIKey:      c.apiKey,
		SearchQuery: query,
	}
	// Execute
	var output any
	if err = c.request(ctx, "search", authedQuery, &output); err != nil {
		return fmt.Errorf("failed to execute API query: %w", err)
	}
	fmt.Printf("%+v\n", output)
	return
}
