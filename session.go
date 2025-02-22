package tavily

import (
	"context"
	"sync/atomic"
)

// session is a sub client that allows to track API usage for a specific session. Instanciate it from the original client.
type session struct {
	parent Client
	statsCounter
}

// Execute a search query using Tavily Search.
// See https://docs.tavily.com/api-reference/endpoint/search for more information.
func (s *session) Search(ctx context.Context, query SearchQuery) (answer SearchAnswer, err error) {
	if answer, err = s.parent.Search(ctx, query); err != nil {
		return
	}
	switch query.SearchDepth {
	case "", SearchQueryDepthBasic:
		s.statsCounter.basicSearches.Add(1)
	case SearchQueryDepthAdvanced:
		s.statsCounter.advancedSearches.Add(1)
	}
	return
}

// Extract web page content from one or more specified URLs using Tavily Extract.
// See https://docs.tavily.com/api-reference/endpoint/extract for more infos.
func (s *session) Extract(ctx context.Context, request ExtractRequest) (answer ExtractAnswer, err error) {
	if answer, err = s.parent.Extract(ctx, request); err != nil {
		return
	}
	switch request.ExtractDepth {
	case "", ExtractRequestDepthBasic:
		s.statsCounter.basicExtracts.Add(int64(len(answer.Results)))
	case ExtractRequestDepthAdvanced:
		s.statsCounter.advancedExtracts.Add(int64(len(answer.Results)))
	}
	return
}

// Create a child client for a new specific session. This is useful for tracking stats per session.
func (s *session) NewSession() Client {
	return &session{
		parent: s,
	}
}

// Stats return the number of searchs (basic and advanced) as well as extract (basic and advanced) performed during this session.
func (s *session) Stats() Stats {
	return s.statsCounter.stats()
}

type statsCounter struct {
	basicSearches    atomic.Int64
	advancedSearches atomic.Int64
	basicExtracts    atomic.Int64
	advancedExtracts atomic.Int64
}

func (sc *statsCounter) stats() (s Stats) {
	s.BasicSearches = int(sc.basicSearches.Load())
	s.AdvancedSearches = int(sc.advancedSearches.Load())
	s.BasicExtracts = int(sc.basicExtracts.Load())
	s.AdvancedExtracts = int(sc.advancedExtracts.Load())
	return
}

// Stats represents an API usage statistics.
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
