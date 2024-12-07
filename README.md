# Tavily Go bindings

[![Go Reference](https://pkg.go.dev/badge/github.com/hekmon/tavily.svg)](https://pkg.go.dev/github.com/hekmon/tavily) [![Go report card](https://goreportcard.com/badge/github.com/hekmon/tavily)](https://goreportcard.com/report/github.com/hekmon/tavily)

These Go bindings implements the [Tavily REST API](https://docs.tavily.com/docs/rest-api/api-reference) for the [Tavily](https://tavily.com/) SaaS service. Tavily offers APIs to search the web and retreive results in a simpe and clean way. It is first intended for LLM Agents but can be used for other purposes as well.

## Features

### Endpoints

All current endpoints are supported:

- [x] Search
- [x] Extract

### Rate Limiting

The client will automatically handle Tavily [rate limiting](https://docs.tavily.com/docs/rest-api/api-reference#rate-limiting) for you.

## Golang types

Every fields of tavily API responses that can be convert to high level Golang types will be converted for ease of use within your code base.

For example: `time.Duration`, `*url.URL`

But they will be reverted to their original type and value if they are marshal again to JSON.

### Error Handling

The client will return an error if the API returns an error status code.

### API Credits

The client will track current session API credits usage thru its stats method/object.

## Usage

### Installation

```bash
go get -v github.com/hekmon/tavily
```

### Example

```go
package main

import (
	"context"

	"github.com/hekmon/tavily"
)

func main() {
	client := tavily.NewClient("<your-tavily-API-key>", nil)
	answer, err := client.Search(context.TODO(), tavily.SearchQuery{
		Query:                    "What is Tavily ?",
		SearchDepth:              tavily.SearchDepthAdvanced,     // optional
		Topic:                    tavily.SearchQueryTopicGeneral, // optional
		MaxResults:               3,                              // optional
		IncludeImages:            true,                           // optional
		IncludeImageDescriptions: true,                           // optional
		IncludeAnswer:            true,                           // optional but recommended for LLMs agents
        // ... others optional params exist
	})
	if err != nil {
		panic(err)
	}
    // Do something with the answer
    // ...
}
```
