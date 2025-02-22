package tavily

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/time/rate"
)

const (
	// Rate Limiting https://docs.tavily.com/guides/rate-limits
	reqPerMinuteDev  = 100
	reqPerMinuteProd = 1000
)

type APIKeyType string

const (
	APIKeyTypeDev  APIKeyType = "dev"
	APIKeyTypeProd APIKeyType = "prod"
)

func NewClient(APIKey string, keyType APIKeyType, customHTTPClient *http.Client) *Client {
	var reqPerMinute int
	switch keyType {
	case APIKeyTypeDev:
		reqPerMinute = reqPerMinuteDev
	case APIKeyTypeProd:
		reqPerMinute = reqPerMinuteProd
	default:
		reqPerMinute = reqPerMinuteDev
	}
	if customHTTPClient == nil {
		customHTTPClient = cleanhttp.DefaultPooledClient()
	}
	return &Client{
		apiKey:     APIKey,
		throughput: rate.NewLimiter(rate.Limit(reqPerMinute)/rate.Limit(time.Minute/time.Second), reqPerMinute),
		httpClient: customHTTPClient,
	}
}

type Client struct {
	apiKey string
	// Controllers
	throughput *rate.Limiter
	httpClient *http.Client
	// Stats
	statsCounter
}

func (c *Client) Stats() (s Stats) {
	return c.statsCounter.Stats()
}

type statsCounter struct {
	basicSearches    atomic.Int64
	advancedSearches atomic.Int64
	basicExtracts    atomic.Int64
	advancedExtracts atomic.Int64
}

func (sc *statsCounter) Stats() (s Stats) {
	s.BasicSearches = int(sc.basicSearches.Load())
	s.AdvancedSearches = int(sc.advancedSearches.Load())
	s.BasicExtracts = int(sc.basicExtracts.Load())
	s.AdvancedExtracts = int(sc.advancedExtracts.Load())
	return
}

type Stats struct {
	BasicSearches    int
	AdvancedSearches int
	BasicExtracts    int
	AdvancedExtracts int
}

// BasicSearchesCost will return the API credits cost of the basic searches.
// See https://docs.tavily.com/guides/api-credits for more infos.
func (s Stats) BasicSearchesCost() float64 {
	return float64(s.BasicSearches)
}

// AdvancedSearchesCost will return the API credits cost of the advanced searches.
// See https://docs.tavily.com/guides/api-credits for more infos.
func (s Stats) AdvancedSearchesCost() float64 {
	return float64(s.AdvancedSearches) * 2
}

// BasicExtractsCost will return the API credits cost of the basic extracts.
// See https://docs.tavily.com/guides/api-credits for more infos.
func (s Stats) BasicExtractsCost() float64 {
	return float64(s.BasicExtracts) / 5
}

// AdvancedExtractsCost will return the API credits cost of the advanced extracts.
// See https://docs.tavily.com/guides/api-credits for more infos.
func (s Stats) AdvancedExtractsCost() float64 {
	return (float64(s.AdvancedExtracts) / 5) * 2
}

// TotalCost will return the total API credits cost of all the searches and extracts.
// See https://docs.tavily.com/guides/api-credits for more infos.
func (s Stats) TotalCost() float64 {
	return s.BasicSearchesCost() + s.AdvancedSearchesCost() + s.BasicExtractsCost() + s.AdvancedExtractsCost()
}
