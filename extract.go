package tavily

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type ExtractRequestDepth string

const (
	ExtractRequestDepthBasic    ExtractRequestDepth = "basic"
	ExtractRequestDepthAdvanced ExtractRequestDepth = "advanced"
)

type ExtractRequest struct {
	URLs          []string            `json:"urls"`
	IncludeImages bool                `json:"include_images,omitempty"`
	ExtractDepth  ExtractRequestDepth `json:"extract_depth,omitempty"`
}

type extractRequestAuth struct {
	APIKey string `json:"api_key"`
	ExtractRequest
}

// Extract web page content from one or more specified URLs using Tavily Extract.
// See https://docs.tavily.com/api-reference/endpoint/extract for more infos.
func (c *mainClient) Extract(ctx context.Context, request ExtractRequest) (answer ExtractAnswer, err error) {
	// Validate URLs
	for _, u := range request.URLs {
		if _, err := url.ParseRequestURI(u); err != nil {
			return answer, fmt.Errorf("invalid URL %q: %w", u, err)
		}
	}
	// Prepare query
	authedRequest := extractRequestAuth{
		APIKey:         c.apiKey,
		ExtractRequest: request,
	}
	// Execute
	if err = c.request(ctx, "extract", authedRequest, &answer); err != nil {
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
