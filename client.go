package tavily

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/time/rate"
)

const (
	reqPerMinute = 100 // https://docs.tavily.com/docs/rest-api/api-reference#rate-limiting
)

type Client struct {
	apiKey     string
	throughput *rate.Limiter
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		throughput: rate.NewLimiter(rate.Limit(reqPerMinute)/rate.Limit(time.Minute/time.Second), reqPerMinute),
		httpClient: cleanhttp.DefaultPooledClient(),
	}
}
