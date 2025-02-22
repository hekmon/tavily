package tavily

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

const (
	// SearchMaxPossibleResults is the maximum results a search query can request, see https://docs.tavily.com/api-reference/endpoint/search#body-max-results
	SearchMaxPossibleResults = 20
)

// SearchQuery represents the parameters for a search query.
// https://docs.tavily.com/docs/rest-api/api-reference#parameters
type SearchQuery struct {
	Query                    string                   `json:"query"`                                // The search query you want to execute with Tavily.
	Topic                    SearchQueryTopic         `json:"topic,omitempty"`                      // The category of the search. This will determine which of our agents will be used for the search. Currently: only "general" and "news" are supported. Default is "general".
	SearchDepth              SearchQueryDepth         `json:"search_depth,omitempty"`               // The depth of the search. It can be "basic" or "advanced". Default is "basic" unless specified otherwise in a given method.
	MaxResults               int                      `json:"max_results,omitempty"`                // The maximum number of search results to return. Default is 5.
	TimeRange                SearchQueryTimeRange     `json:"time_range,omitempty"`                 // The time range back from the current date to filter results.
	Days                     int                      `json:"days,omitempty"`                       // The number of days back from the current date to include in the search results. This specifies the time frame of data to be retrieved. Please note that this feature is only available when using the "news" search topic. Default is 3.
	IncludeAnswer            SearchQueryIncludeAnswer `json:"include_answer,omitempty"`             // Include a short answer to original query. Default is False.
	IncludeRawContent        bool                     `json:"include_raw_content"`                  // Include the cleaned and parsed HTML content of each search result. Default is False.
	IncludeImages            bool                     `json:"include_images,omitempty"`             // Include a list of query-related images in the response. Default is False.
	IncludeImageDescriptions bool                     `json:"include_image_descriptions,omitempty"` // When include_images is set to True, this option adds descriptive text for each image. Default is False.
	IncludeDomains           []string                 `json:"include_domains,omitempty"`            // A list of domains to specifically include in the search results. Default is [], which includes all domains.
	ExcludeDomains           []string                 `json:"exclude_domains,omitempty"`            // A list of domains to specifically exclude from the search results. Default is [], which doesn't exclude any domains.
}

func (sq SearchQuery) Validate() error {
	// Query
	if sq.Query == "" {
		return errors.New("query is required")
	}
	// Topic
	switch sq.Topic {
	case SearchQueryTopicGeneral, SearchQueryTopicNews, "":
	default:
		return errors.New("invalid topic")
	}
	// Search Depth
	switch sq.SearchDepth {
	case SearchQueryDepthBasic, SearchQueryDepthAdvanced, "":
	default:
		return errors.New("invalid search depth")
	}
	// Max Results
	switch {
	case sq.MaxResults <= 0:
		return errors.New("max_results must be a non-negative integer")
	case sq.MaxResults > SearchMaxPossibleResults:
		return fmt.Errorf("max_results must be less than or equal to %d", SearchMaxPossibleResults)
	}
	// Time Range
	if sq.TimeRange != SearchQueryTimeRangeDisabled {
		if sq.Topic == SearchQueryTopicNews {
			return fmt.Errorf("time_range can only be specified when using the %q topic", SearchQueryTopicGeneral)
		}
		switch sq.TimeRange {
		case SearchQueryTimeRangeDay, SearchQueryTimeRangeWeek, SearchQueryTimeRangeMonth, SearchQueryTimeRangeYear:
		default:
			return errors.New("invalid time range")
		}
	}
	// Days
	switch {
	case sq.Days < 0:
		return errors.New("days must be a non-negative integer")
	case sq.Days > 0 && sq.Topic != SearchQueryTopicNews:
		return fmt.Errorf("days can only be specified when using the %q topic", SearchQueryTopicNews)
	}
	// Include Answer
	switch sq.IncludeAnswer {
	case SearchQueryIncludeAnswerNone, SearchQueryIncludeAnswerBasic, SearchQueryIncludeAnswerAdvanced:
	default:
		return errors.New("invalid include_answer value")
	}
	// Images descriptions
	if !sq.IncludeImages && sq.IncludeImageDescriptions {
		return errors.New("include_image_descriptions can only be true when include_images is true")
	}
	return nil
}

type SearchQueryDepth string

const (
	SearchQueryDepthBasic    SearchQueryDepth = "basic"
	SearchQueryDepthAdvanced SearchQueryDepth = "advanced"
)

type SearchQueryTopic string

const (
	SearchQueryTopicGeneral SearchQueryTopic = "general"
	SearchQueryTopicNews    SearchQueryTopic = "news"
)

type SearchQueryTimeRange string

const (
	SearchQueryTimeRangeDisabled SearchQueryTimeRange = ""
	SearchQueryTimeRangeDay      SearchQueryTimeRange = "day"
	SearchQueryTimeRangeWeek     SearchQueryTimeRange = "week"
	SearchQueryTimeRangeMonth    SearchQueryTimeRange = "month"
	SearchQueryTimeRangeYear     SearchQueryTimeRange = "year"
)

type SearchQueryIncludeAnswer string

const (
	SearchQueryIncludeAnswerNone     SearchQueryIncludeAnswer = ""
	SearchQueryIncludeAnswerBasic    SearchQueryIncludeAnswer = "basic"
	SearchQueryIncludeAnswerAdvanced SearchQueryIncludeAnswer = "advanced"
)

// Execute a search query using Tavily Search.
// See https://docs.tavily.com/api-reference/endpoint/search for more information.
func (c *mainClient) Search(ctx context.Context, query SearchQuery) (answer SearchAnswer, err error) {
	// Prepare query
	if err = query.Validate(); err != nil {
		err = fmt.Errorf("failed to validate search query: %w", err)
		return
	}
	// Execute
	if err = c.request(ctx, "search", query, &answer); err != nil {
		err = fmt.Errorf("failed to execute API query: %w", err)
	}
	return
}

// SearchAnswer represents the response from the search API.
// https://docs.tavily.com/docs/rest-api/api-reference#response
type SearchAnswer struct {
	Query        string               `json:"query"`
	Answer       *string              `json:"answer"`
	Images       []SearchAnswerImage  `json:"images"`
	Results      []SearchAnswerResult `json:"results"`
	ResponseTime time.Duration        `json:"-"`
}

func (sa *SearchAnswer) UnmarshalJSON(data []byte) (err error) {
	type mask SearchAnswer
	tmp := struct {
		*mask
		ResponseTime float64 `json:"response_time"`
	}{
		mask: (*mask)(sa),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into tmp struct: %w", err)
	}
	sa.ResponseTime = time.Duration(tmp.ResponseTime * float64(time.Second))
	return
}

func (sa SearchAnswer) MarshalJSON() ([]byte, error) {
	type mask SearchAnswer
	tmp := struct {
		mask
		ResponseTime float64 `json:"response_time"`
	}{
		mask:         mask(sa),
		ResponseTime: sa.ResponseTime.Seconds(),
	}
	return json.Marshal(tmp)
}

type SearchAnswerImage struct {
	URL         *url.URL `json:"-"`
	Description string   `json:"description"`
}

func (sai *SearchAnswerImage) UnmarshalJSON(data []byte) (err error) {
	type mask SearchAnswerImage
	tmp := struct {
		URL string `json:"url"`
		*mask
	}{
		mask: (*mask)(sai),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into tmp struct: %w", err)
	}
	if sai.URL, err = url.Parse(tmp.URL); err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	return
}

func (sai SearchAnswerImage) MarshalJSON() ([]byte, error) {
	type mask SearchAnswerImage
	tmp := struct {
		URL string `json:"url"`
		mask
	}{
		URL:  sai.URL.String(),
		mask: mask(sai),
	}
	return json.Marshal(tmp)
}

type SearchAnswerResult struct {
	Title      string   `json:"title"`
	URL        *url.URL `json:"-"`
	Content    string   `json:"content"`
	Score      float64  `json:"score"`
	RawContent *string  `json:"raw_content"`
}

func (sar *SearchAnswerResult) UnmarshalJSON(data []byte) (err error) {
	type mask SearchAnswerResult
	tmp := struct {
		*mask
		URL string `json:"url"`
	}{
		mask: (*mask)(sar),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into tmp struct: %w", err)
	}
	if sar.URL, err = url.Parse(tmp.URL); err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	return
}

func (sar SearchAnswerResult) MarshalJSON() ([]byte, error) {
	type mask SearchAnswerResult
	tmp := struct {
		URL string `json:"url"`
		mask
	}{
		URL:  sar.URL.String(),
		mask: mask(sar),
	}
	return json.Marshal(tmp)
}
