package recommendations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
)

// Service provides access to the Yahoo Finance Recommendations API.
type Service struct {
	fetcher *fetch.Fetcher
}

// NewService creates a new Recommendations service.
func NewService(f *fetch.Fetcher) *Service {
	return &Service{fetcher: f}
}

// Get fetches recommended symbols for the given list of symbols.
// No crumb authentication is required for this endpoint.
func (s *Service) Get(ctx context.Context, symbols []string, opts *Options) ([]Result, error) {
	if len(symbols) == 0 {
		return nil, &errors.InvalidOptionsError{Field: "symbols", Msg: "at least one symbol is required"}
	}

	// Build URL: symbols are comma-joined and placed in the path
	joinedSymbols := url.PathEscape(strings.Join(symbols, ","))
	u := fmt.Sprintf("${YF_QUERY_HOST}/v6/finance/recommendationsbysymbol/%s", joinedSymbols)

	// Fetch without crumb
	body, err := s.fetcher.Fetch(ctx, fetch.FetchConfig{URL: u, NeedsCrumb: false})
	if err != nil {
		return nil, err
	}

	// Parse response: {"finance":{"result":[...],"error":null}}
	var raw struct {
		Finance struct {
			Result []Result     `json:"result"`
			Error  interface{} `json:"error"`
		} `json:"finance"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing recommendations response: %w", err)
	}

	// Check for API error
	if raw.Finance.Error != nil {
		return nil, &errors.HTTPError{StatusCode: 200, Body: fmt.Sprintf("recommendations API error: %v", raw.Finance.Error)}
	}

	return raw.Finance.Result, nil
}
