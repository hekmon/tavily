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
type SearchQuery struct {
	Query                    string                        `json:"query"`                                // The search query you want to execute with Tavily.
	Topic                    SearchQueryTopic              `json:"topic,omitempty"`                      // The category of the search. This will determine which of our agents will be used for the search. Currently: "general", "news" and "finance" are supported. Default is "general".
	SearchDepth              SearchQueryDepth              `json:"search_depth,omitempty"`               // The depth of the search. It can be "basic", "advanced", "fast" or "ultra-fast". Default is "basic".
	ChunksPerSource          int                           `json:"chunks_per_source,omitempty"`          // The maximum number of chunks to return per source. Available when search_depth is advanced, basic or fast. Must be between 1 and 3.
	MaxResults               int                           `json:"max_results,omitempty"`                // The maximum number of search results to return. Default is 5.
	TimeRange                SearchQueryTimeRange          `json:"time_range,omitempty"`                 // The time range back from the current date to filter results.
	StartDate                string                        `json:"start_date,omitempty"`                 // Will return all results after the specified start date based on publish date (YYYY-MM-DD).
	EndDate                  string                        `json:"end_date,omitempty"`                   // Will return all results before the specified end date based on publish date (YYYY-MM-DD).
	IncludePublishedDate     bool                          `json:"include_published_date,omitempty"`     // Include a published_date field in each result.
	FilterByPublishedDate    bool                          `json:"filter_by_published_date,omitempty"`   // Remove results whose published date falls outside the time window.
	IncludeAnswer            SearchQueryIncludeAnswer      `json:"include_answer,omitempty"`             // Include a short answer to original query. Default is false.
	IncludeRawContent        SearchQueryIncludeRawContent  `json:"include_raw_content"`                  // Include the cleaned and parsed HTML content of each search result.
	IncludeImages            bool                          `json:"include_images,omitempty"`             // Include a list of query-related images in the response. Default is false.
	IncludeImageDescriptions bool                          `json:"include_image_descriptions,omitempty"` // When include_images is set to True, this option adds descriptive text for each image. Default is false.
	IncludeFavicon           bool                          `json:"include_favicon,omitempty"`            // Whether to include the favicon URL for each result.
	IncludeDomains           []string                      `json:"include_domains,omitempty"`            // A list of domains to specifically include in the search results. Default is [], which includes all domains.
	ExcludeDomains           []string                      `json:"exclude_domains,omitempty"`            // A list of domains to specifically exclude from the search results. Default is [], which doesn't exclude any domains.
	IncludeDomainsMode       SearchQueryIncludeDomainsMode `json:"include_domains_mode,omitempty"`       // Controls how include_domains is applied: "restrict" or "prefer".
	Country                  SearchQueryCountry            `json:"country,omitempty"`                    // Boost search results from a specific country.
	Language                 SearchQueryLanguage           `json:"language,omitempty"`                   // Boost search results in a specific language (ISO 639-1 code or English name).
	FilterByLanguage         bool                          `json:"filter_by_language,omitempty"`         // Strictly filter out search results that don't match the language parameter.
	AutoParameters           bool                          `json:"auto_parameters,omitempty"`            // Automatically configure search parameters based on query content.
	ExactMatch               bool                          `json:"exact_match,omitempty"`                // Ensure that only search results containing the exact quoted phrase(s) in the query are returned.
	IncludeUsage             bool                          `json:"include_usage,omitempty"`              // Whether to include credit usage information in the response.
	SafeSearch               bool                          `json:"safe_search,omitempty"`                // Whether to filter out adult or unsafe content from results.
}

func (sq SearchQuery) Validate() error {
	// Query
	if sq.Query == "" {
		return errors.New("query is required")
	}
	// Topic
	switch sq.Topic {
	case SearchQueryTopicGeneral, SearchQueryTopicNews, SearchQueryTopicFinance, "":
	default:
		return errors.New("invalid topic")
	}
	// Search Depth
	switch sq.SearchDepth {
	case SearchQueryDepthBasic, SearchQueryDepthAdvanced, SearchQueryDepthFast, SearchQueryDepthUltraFast, "":
	default:
		return errors.New("invalid search depth")
	}
	// Max Results
	switch {
	case sq.MaxResults < 0:
		return errors.New("max_results must be a non-negative integer")
	case sq.MaxResults > SearchMaxPossibleResults:
		return fmt.Errorf("max_results must be less than or equal to %d", SearchMaxPossibleResults)
	}
	// Domains
	if len(sq.IncludeDomains) > 300 {
		return errors.New("include_domains must not exceed 300 domains")
	}
	if len(sq.ExcludeDomains) > 150 {
		return errors.New("exclude_domains must not exceed 150 domains")
	}
	// Chunks per source
	if sq.ChunksPerSource > 0 {
		if sq.ChunksPerSource < 1 || sq.ChunksPerSource > 3 {
			return errors.New("chunks_per_source must be between 1 and 3")
		}
		if sq.SearchDepth == SearchQueryDepthUltraFast {
			return errors.New("chunks_per_source is not available with ultra-fast search depth")
		}
	}
	// Time Range
	if sq.TimeRange != "" {
		switch sq.TimeRange {
		case SearchQueryTimeRangeDay, SearchQueryTimeRangeWeek, SearchQueryTimeRangeMonth, SearchQueryTimeRangeYear,
			SearchQueryTimeRangeDayShort, SearchQueryTimeRangeWeekShort, SearchQueryTimeRangeMonthShort, SearchQueryTimeRangeYearShort:
		default:
			return errors.New("invalid time range")
		}
	}
	// Dates
	if sq.StartDate != "" {
		if _, err := time.Parse("2006-01-02", sq.StartDate); err != nil {
			return fmt.Errorf("invalid start_date format: %w", err)
		}
	}
	if sq.EndDate != "" {
		if _, err := time.Parse("2006-01-02", sq.EndDate); err != nil {
			return fmt.Errorf("invalid end_date format: %w", err)
		}
	}
	// Include Answer
	switch sq.IncludeAnswer {
	case "", SearchQueryIncludeAnswerFalse, SearchQueryIncludeAnswerTrue,
		SearchQueryIncludeAnswerBasic, SearchQueryIncludeAnswerAdvanced:
	default:
		return errors.New("invalid include_answer value")
	}
	// Include Raw Content
	switch sq.IncludeRawContent {
	case "", SearchQueryIncludeRawContentFalse, SearchQueryIncludeRawContentTrue,
		SearchQueryIncludeRawContentMarkdown, SearchQueryIncludeRawContentText:
	default:
		return errors.New("invalid include_raw_content value")
	}
	// Images descriptions
	if !sq.IncludeImages && sq.IncludeImageDescriptions {
		return errors.New("include_image_descriptions can only be true when include_images is true")
	}
	// Include domains mode
	if sq.IncludeDomainsMode != "" && len(sq.IncludeDomains) == 0 {
		return errors.New("include_domains_mode requires include_domains to be set")
	}
	switch sq.IncludeDomainsMode {
	case SearchQueryIncludeDomainsModeRestrict, SearchQueryIncludeDomainsModePrefer, "":
	default:
		return errors.New("invalid include_domains_mode")
	}
	// Filter by language
	if sq.FilterByLanguage && sq.Language == "" {
		return errors.New("filter_by_language requires language to be set")
	}
	// Safe search
	if sq.SafeSearch && (sq.SearchDepth == SearchQueryDepthFast || sq.SearchDepth == SearchQueryDepthUltraFast) {
		return errors.New("safe_search is not supported for fast or ultra-fast search depths")
	}
	return nil
}

type SearchQueryDepth string

const (
	SearchQueryDepthBasic     SearchQueryDepth = "basic"
	SearchQueryDepthAdvanced  SearchQueryDepth = "advanced"
	SearchQueryDepthFast      SearchQueryDepth = "fast"
	SearchQueryDepthUltraFast SearchQueryDepth = "ultra-fast"
)

type SearchQueryTopic string

const (
	SearchQueryTopicGeneral SearchQueryTopic = "general"
	SearchQueryTopicNews    SearchQueryTopic = "news"
	SearchQueryTopicFinance SearchQueryTopic = "finance"
)

type SearchQueryTimeRange string

const (
	SearchQueryTimeRangeDay        SearchQueryTimeRange = "day"
	SearchQueryTimeRangeWeek       SearchQueryTimeRange = "week"
	SearchQueryTimeRangeMonth      SearchQueryTimeRange = "month"
	SearchQueryTimeRangeYear       SearchQueryTimeRange = "year"
	SearchQueryTimeRangeDayShort   SearchQueryTimeRange = "d"
	SearchQueryTimeRangeWeekShort  SearchQueryTimeRange = "w"
	SearchQueryTimeRangeMonthShort SearchQueryTimeRange = "m"
	SearchQueryTimeRangeYearShort  SearchQueryTimeRange = "y"
)

type SearchQueryIncludeAnswer string

const (
	SearchQueryIncludeAnswerFalse    SearchQueryIncludeAnswer = "false"
	SearchQueryIncludeAnswerTrue     SearchQueryIncludeAnswer = "true"
	SearchQueryIncludeAnswerBasic    SearchQueryIncludeAnswer = "basic"
	SearchQueryIncludeAnswerAdvanced SearchQueryIncludeAnswer = "advanced"
)

func (ia SearchQueryIncludeAnswer) MarshalJSON() ([]byte, error) {
	switch ia {
	case "", SearchQueryIncludeAnswerFalse:
		return []byte("false"), nil
	case SearchQueryIncludeAnswerTrue, SearchQueryIncludeAnswerBasic:
		return []byte("true"), nil
	case SearchQueryIncludeAnswerAdvanced:
		return []byte(`"advanced"`), nil
	default:
		return json.Marshal(string(ia))
	}
}

func (ia *SearchQueryIncludeAnswer) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*ia = SearchQueryIncludeAnswer(s)
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	if b {
		*ia = SearchQueryIncludeAnswerTrue
	} else {
		*ia = SearchQueryIncludeAnswerFalse
	}
	return nil
}

type SearchQueryIncludeRawContent string

const (
	SearchQueryIncludeRawContentFalse    SearchQueryIncludeRawContent = "false"
	SearchQueryIncludeRawContentTrue     SearchQueryIncludeRawContent = "true"
	SearchQueryIncludeRawContentMarkdown SearchQueryIncludeRawContent = "markdown"
	SearchQueryIncludeRawContentText     SearchQueryIncludeRawContent = "text"
)

func (irc SearchQueryIncludeRawContent) MarshalJSON() ([]byte, error) {
	switch irc {
	case "", SearchQueryIncludeRawContentFalse:
		return []byte("false"), nil
	case SearchQueryIncludeRawContentTrue, SearchQueryIncludeRawContentMarkdown:
		return []byte("true"), nil
	case SearchQueryIncludeRawContentText:
		return []byte(`"text"`), nil
	default:
		return json.Marshal(string(irc))
	}
}

func (irc *SearchQueryIncludeRawContent) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*irc = SearchQueryIncludeRawContent(s)
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	if b {
		*irc = SearchQueryIncludeRawContentTrue
	} else {
		*irc = SearchQueryIncludeRawContentFalse
	}
	return nil
}

type SearchQueryIncludeDomainsMode string

const (
	SearchQueryIncludeDomainsModeRestrict SearchQueryIncludeDomainsMode = "restrict"
	SearchQueryIncludeDomainsModePrefer   SearchQueryIncludeDomainsMode = "prefer"
)

type SearchQueryCountry string
type SearchQueryLanguage string

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
// https://docs.tavily.com/documentation/api-reference/endpoint/search
type SearchAnswer struct {
	Query          string                      `json:"query"`
	Answer         *string                     `json:"answer"`
	Images         []SearchAnswerImage         `json:"images"`
	Results        []SearchAnswerResult        `json:"results"`
	ResponseTime   time.Duration               `json:"-"`
	AutoParameters *SearchAnswerAutoParameters `json:"auto_parameters,omitempty"`
	Usage          *SearchAnswerUsage          `json:"usage,omitempty"`
	RequestID      string                      `json:"request_id"`
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

type SearchAnswerAutoParameters struct {
	Topic       SearchQueryTopic `json:"topic"`
	SearchDepth SearchQueryDepth `json:"search_depth"`
}

type SearchAnswerUsage struct {
	Credits int `json:"credits"`
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
		mask: mask(sai),
	}
	if sai.URL != nil {
		tmp.URL = sai.URL.String()
	}
	return json.Marshal(tmp)
}

type SearchAnswerResult struct {
	Title         string              `json:"title"`
	URL           *url.URL            `json:"-"`
	Content       string              `json:"content"`
	Score         float64             `json:"score"`
	RawContent    *string             `json:"raw_content"`
	PublishedDate *time.Time          `json:"-"`
	Favicon       *url.URL            `json:"-"`
	Images        []SearchAnswerImage `json:"images"`
	ID            string              `json:"id"`
}

func (sar *SearchAnswerResult) UnmarshalJSON(data []byte) (err error) {
	type mask SearchAnswerResult
	tmp := struct {
		*mask
		URL           string `json:"url"`
		PublishedDate string `json:"published_date"`
		Favicon       string `json:"favicon"`
	}{
		mask: (*mask)(sar),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into tmp struct: %w", err)
	}
	if sar.URL, err = url.Parse(tmp.URL); err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	if tmp.PublishedDate != "" {
		pd, err := time.Parse(time.RFC1123, tmp.PublishedDate)
		if err != nil {
			return fmt.Errorf("failed to parse published_date: %w", err)
		}
		sar.PublishedDate = &pd
	}
	if tmp.Favicon != "" {
		if sar.Favicon, err = url.Parse(tmp.Favicon); err != nil {
			return fmt.Errorf("failed to parse favicon URL: %w", err)
		}
	}
	return
}

func (sar SearchAnswerResult) MarshalJSON() ([]byte, error) {
	type mask SearchAnswerResult
	tmp := struct {
		mask
		URL           string  `json:"url"`
		PublishedDate *string `json:"published_date,omitempty"`
		Favicon       string  `json:"favicon,omitempty"`
	}{
		mask: mask(sar),
		URL:  sar.URL.String(),
	}
	if sar.PublishedDate != nil {
		s := sar.PublishedDate.Format(time.RFC1123)
		tmp.PublishedDate = &s
	}
	if sar.Favicon != nil {
		tmp.Favicon = sar.Favicon.String()
	}
	return json.Marshal(tmp)
}
