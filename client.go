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
	basicSearches    atomic.Int64
	advancedSearches atomic.Int64
	basicExtracts    atomic.Int64
	advancedExtracts atomic.Int64
}

func (c *Client) SessionStats() (s Stats) {
	s.BasicSearches = int(c.basicSearches.Load())
	s.AdvancedSearches = int(c.advancedSearches.Load())
	s.BasicExtracts = int(c.basicExtracts.Load())
	s.AdvancedExtracts = int(c.advancedExtracts.Load())
	return
}

type Stats struct {
	BasicSearches    int
	AdvancedSearches int
	BasicExtracts    int
	AdvancedExtracts int
}

// APICredits will return the API credits costs of the stats.
// https://docs.tavily.com/guides/api-credits
func (s Stats) APICreditsCost() int {
	return s.BasicSearches + s.AdvancedSearches*2 + s.BasicExtracts/5 + (s.AdvancedExtracts/5)*2
}
