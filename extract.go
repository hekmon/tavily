package tavily

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

type ExtractRequestDepth string

const (
	ExtractRequestDepthBasic    ExtractRequestDepth = "basic"
	ExtractRequestDepthAdvanced ExtractRequestDepth = "advanced"
)

type ExtractRequestFormat string

const (
	ExtractRequestFormatMarkdown ExtractRequestFormat = "markdown"
	ExtractRequestFormatText     ExtractRequestFormat = "text"
)

type ExtractRequest struct {
	URLs            []string             `json:"urls"`
	Query           string               `json:"query,omitempty"`
	ChunksPerSource int                  `json:"chunks_per_source,omitempty"`
	ExtractDepth    ExtractRequestDepth  `json:"extract_depth,omitempty"`
	IncludeImages   bool                 `json:"include_images,omitempty"`
	IncludeFavicon  bool                 `json:"include_favicon,omitempty"`
	Format          ExtractRequestFormat `json:"format,omitempty"`
	Timeout         *float64             `json:"timeout,omitempty"`
	IncludeUsage    bool                 `json:"include_usage,omitempty"`
}

func (er ExtractRequest) Validate() error {
	// URLs
	if len(er.URLs) == 0 {
		return errors.New("urls is required")
	}
	if len(er.URLs) > 20 {
		return errors.New("urls must not exceed 20 URLs")
	}
	for _, u := range er.URLs {
		if _, err := url.ParseRequestURI(u); err != nil {
			return fmt.Errorf("invalid URL %q: %w", u, err)
		}
	}
	// Extract depth
	switch er.ExtractDepth {
	case ExtractRequestDepthBasic, ExtractRequestDepthAdvanced, "":
	default:
		return errors.New("invalid extract depth")
	}
	// Chunks per source
	if er.ChunksPerSource > 0 {
		if er.ChunksPerSource < 1 || er.ChunksPerSource > 5 {
			return errors.New("chunks_per_source must be between 1 and 5")
		}
	}
	// Format
	switch er.Format {
	case ExtractRequestFormatMarkdown, ExtractRequestFormatText, "":
	default:
		return errors.New("invalid format")
	}
	// Timeout
	if er.Timeout != nil {
		if *er.Timeout < 1.0 || *er.Timeout > 60.0 {
			return errors.New("timeout must be between 1.0 and 60.0")
		}
	}
	return nil
}

// Extract web page content from one or more specified URLs using Tavily Extract.
// See https://docs.tavily.com/api-reference/endpoint/extract for more infos.
func (c *mainClient) Extract(ctx context.Context, request ExtractRequest) (answer ExtractAnswer, err error) {
	// Prepare query
	if err = request.Validate(); err != nil {
		err = fmt.Errorf("failed to validate extract request: %w", err)
		return
	}
	// Execute
	if err = c.request(ctx, "extract", request, &answer); err != nil {
		err = fmt.Errorf("failed to execute API query: %w", err)
	}
	return
}

type ExtractAnswer struct {
	Results       []ExtractAnswerResult       `json:"results"`
	FailedResults []ExtractAnswerFailedResult `json:"failed_results"`
	ResponseTime  time.Duration               `json:"-"`
	Usage         *ExtractAnswerUsage         `json:"usage,omitempty"`
	RequestID     string                      `json:"request_id"`
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

type ExtractAnswerUsage struct {
	Credits int `json:"credits"`
}

type ExtractAnswerResult struct {
	URL        *url.URL             `json:"-"`
	RawContent string               `json:"raw_content"`
	Images     []ExtractAnswerImage `json:"images"`
}

func (ear *ExtractAnswerResult) UnmarshalJSON(data []byte) (err error) {
	type mask ExtractAnswerResult
	tmp := struct {
		*mask
		URL string `json:"url"`
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

type ExtractAnswerImage struct {
	URL *url.URL `json:"-"`
}

func (eai *ExtractAnswerImage) UnmarshalJSON(data []byte) (err error) {
	type mask ExtractAnswerImage
	tmp := struct {
		URL string `json:"url"`
		*mask
	}{
		mask: (*mask)(eai),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into tmp struct: %w", err)
	}
	if eai.URL, err = url.Parse(tmp.URL); err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	return
}

func (eai ExtractAnswerImage) MarshalJSON() ([]byte, error) {
	type mask ExtractAnswerImage
	tmp := struct {
		URL string `json:"url"`
		mask
	}{
		mask: mask(eai),
	}
	if eai.URL != nil {
		tmp.URL = eai.URL.String()
	}
	return json.Marshal(tmp)
}

type ExtractAnswerFailedResult struct {
	URL   string `json:"url"` // can be invalid, can not use url.URL here
	Error string `json:"error"`
}
