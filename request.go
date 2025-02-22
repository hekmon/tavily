package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

const (
	apiURL    = "https://api.tavily.com/"
	userAgent = "github.com/hekmon/tavily"
)

var (
	baseURL *url.URL
)

func init() {
	var err error
	if baseURL, err = url.Parse(apiURL); err != nil {
		panic(err)
	}
}

func (c *mainClient) request(ctx context.Context, endpoint string, payload, response any) (err error) {
	// Prepare payload
	var body bytes.Buffer
	if payload != nil {
		if err = json.NewEncoder(&body).Encode(payload); err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}
	// Create request
	reqURL := *baseURL
	reqURL.Path = path.Join(reqURL.Path, endpoint)
	req, err := http.NewRequest("POST", reqURL.String(), &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	if response != nil {
		req.Header.Set("Accept", "application/json")
	}
	// Respect Tavily rate limits
	if err = c.throughput.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for rate limiting: %w", err)
	}
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	// Handle status code
	switch resp.StatusCode {
	case http.StatusOK:
		if response == nil {
			// no need to continue to unmarshalling
			return nil
		}
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
		http.StatusMethodNotAllowed, http.StatusUnprocessableEntity, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		// Handle known errors
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte(fmt.Sprintf("failed to read response body: %s", err))
		}
		return APIError{
			Code: resp.StatusCode,
			Body: body,
		}
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	// Unmarshal response
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return
}

// APIError represents a known error from the Tavily API.
// https://docs.tavily.com/docs/rest-api/api-reference#error-codes
type APIError struct {
	Code int
	Body []byte
}

func (e APIError) Error() string {
	return fmt.Sprintf("Tavily API error: %d %s", e.Code, http.StatusText(e.Code))
}
