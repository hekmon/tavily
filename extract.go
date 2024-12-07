package tavily

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type extractRequest struct {
	APIKey string   `json:"api_key"`
	URLs   []string `json:"urls"`
}

// Extract retrieve raw web content from specified URLs.
// https://docs.tavily.com/docs/rest-api/api-reference#endpoint-post-extract
func (c *Client) Extract(ctx context.Context, urls []string) (answer ExtractAnswer, err error) {
	// Prepare query
	authedQuery := extractRequest{
		APIKey: c.apiKey,
		URLs:   urls,
	}
	// Execute
	if err = c.request(ctx, "extract", authedQuery, &answer); err != nil {
		err = fmt.Errorf("failed to execute API query: %w", err)
	}
	return
}

type ExtractAnswer struct {
	Results       []ExtractAnswerResult       `json:"results"`
	FailedResults []ExtractAnswerFailedResult `json:"failed_results"`
	ResponseTime  time.Duration               `json:"-"`
}

func (ea *ExtractAnswer) UnmarshalJSON(data []byte) (err error) {
	type mask ExtractAnswer
	tmp := struct {
		*mask
		ResponseTime float64 `json:"response_time"`
	}{
		mask: (*mask)(ea),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into tmp struct: %w", err)
	}
	ea.ResponseTime = time.Duration(tmp.ResponseTime * float64(time.Second))
	return
}

func (ea ExtractAnswer) MarshalJSON() ([]byte, error) {
	type mask ExtractAnswer
	tmp := struct {
		mask
		ResponseTime float64 `json:"response_time"`
	}{
		mask:         mask(ea),
		ResponseTime: ea.ResponseTime.Seconds(),
	}
	return json.Marshal(tmp)
}

type ExtractAnswerResult struct {
	URL        *url.URL `json:"-"`
	RawContent string   `json:"raw_content"`
}

func (ear *ExtractAnswerResult) UnmarshalJSON(data []byte) (err error) {
	type mask ExtractAnswerResult
	tmp := struct {
		URL string `json:"url"`
		*mask
	}{
		mask: (*mask)(ear),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into tmp struct: %w", err)
	}
	if ear.URL, err = url.Parse(tmp.URL); err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	return
}

func (ear ExtractAnswerResult) MarshalJSON() ([]byte, error) {
	type mask ExtractAnswerResult
	tmp := struct {
		URL string `json:"url"`
		mask
	}{
		URL:  ear.URL.String(),
		mask: mask(ear),
	}
	return json.Marshal(tmp)
}

type ExtractAnswerFailedResult struct {
	URL    string `json:"url"` // can be invalid, can not use url.URL here
	Reason string `json:"reason"`
}
