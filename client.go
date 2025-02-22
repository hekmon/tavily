package tavily

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/time/rate"
)

type Client interface {
	Search(context.Context, SearchQuery) (SearchAnswer, error)
	Extract(context.Context, ExtractRequest) (ExtractAnswer, error)
	// Stats return the number of searchs (basic and advanced) as well as extract (basic and advanced) this client has performed.
	Stats() Stats
	// Create a child client for a specific session. This is useful for tracking stats per session. Parent stats will include child stats.
	NewSession() Client
}

const (
	// ReqPerMinuteDev represents the number of requests a Development API key can made per minute. For more information see https://docs.tavily.com/guides/rate-limits
	ReqPerMinuteDev = 100
	// ReqPerMinuteProd represents the number of requests a Production API key can made per minute. For more information see https://docs.tavily.com/guides/rate-limits
	ReqPerMinuteProd = 1000
)

// APIKeyType represents the type of API key used.
type APIKeyType string

const (
	// APIKeyTypeDev represents a development API key.
	APIKeyTypeDev APIKeyType = "dev"
	// APIKeyTypeProd represents a production API key.
	APIKeyTypeProd APIKeyType = "prod"
)

func NewClient(APIKey string, keyType APIKeyType, customHTTPClient *http.Client) Client {
	var reqPerMinute int
	switch keyType {
	case APIKeyTypeDev:
		reqPerMinute = ReqPerMinuteDev
	case APIKeyTypeProd:
		reqPerMinute = ReqPerMinuteProd
	default:
		reqPerMinute = ReqPerMinuteDev
	}
	if customHTTPClient == nil {
		customHTTPClient = cleanhttp.DefaultPooledClient()
	}
	mc := &mainClient{
		apiKey:     APIKey,
		throughput: rate.NewLimiter(rate.Limit(reqPerMinute)/rate.Limit(time.Minute/time.Second), reqPerMinute),
		httpClient: customHTTPClient,
	}
	return mc.NewSession()
}

type mainClient struct {
	apiKey string
	// Controllers
	throughput *rate.Limiter
	httpClient *http.Client
}

// main client does not hold stats as it is never returned directly to the client (a session is), just implementing interface here
func (c *mainClient) Stats() Stats {
	return Stats{}
}

// creates a root session for this client
func (c *mainClient) NewSession() Client {
	return &session{
		client: c,
	}
}
