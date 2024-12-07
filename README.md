# Tavily Go bindings

[![Go Reference](https://pkg.go.dev/badge/github.com/hekmon/tavily.svg)](https://pkg.go.dev/github.com/hekmon/tavily) [![Go report card](https://goreportcard.com/badge/github.com/hekmon/tavily)](https://goreportcard.com/report/github.com/hekmon/tavily)

These Go bindings implements the [Tavily REST API](https://docs.tavily.com/docs/rest-api/api-reference) for the [Tavily](https://tavily.com/) SaaS service. Tavily offers APIs to search the web and retreive results in a simpe and clean way. It is first intended for LLM Agents but can be used for other purposes as well.

## Features

### Endpoints

All current endpoints are supported:

- [x] [Search](https://docs.tavily.com/docs/rest-api/api-reference#endpoint-post-search)
- [x] [Extract](https://docs.tavily.com/docs/rest-api/api-reference#endpoint-post-extract)

### Rate Limiting

The client will automatically handle Tavily [rate limiting](https://docs.tavily.com/docs/rest-api/api-reference#rate-limiting) for you.

### API Credits

The client will track current session API credits usage thru its stats method/object.

## Installation

```bash
go get -v github.com/hekmon/tavily
```
