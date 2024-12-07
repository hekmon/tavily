package tavily

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/time/rate"
)

const (
	reqPerMinute = 100 // https://docs.tavily.com/docs/rest-api/api-reference#rate-limiting
)

func NewClient(apiKey string, customHTTPClient *http.Client) *Client {
	if customHTTPClient == nil {
		customHTTPClient = cleanhttp.DefaultPooledClient()
	}
	return &Client{
		apiKey:     apiKey,
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
	extracts         atomic.Int64
}

func (c *Client) SessionStats() (s Stats) {
	s.BasicSearches = int(c.basicSearches.Load())
	s.AdvancedSearches = int(c.advancedSearches.Load())
	s.Extracts = int(c.extracts.Load())
	return
}

type Stats struct {
	BasicSearches    int
	AdvancedSearches int
	Extracts         int
}

// APICredits will return the API credits costs of the stats.
// https://docs.tavily.com/docs/rest-api/api-reference#tavily-api-credit-deduction-overview
func (s Stats) APICreditsCost() int {
	return s.BasicSearches + s.AdvancedSearches*2 + s.Extracts/5
}
