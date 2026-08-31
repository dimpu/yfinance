package screener

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
)

// Service provides access to the Yahoo Finance Screener API.
type Service struct {
	fetcher *fetch.Fetcher
}

// NewService creates a new Screener service.
func NewService(f *fetch.Fetcher) *Service {
	return &Service{fetcher: f}
}

// Get fetches predefined screener results from Yahoo Finance.
// The screener ID must be one of the PredefinedScreenerModule constants.
// This endpoint requires crumb authentication.
func (s *Service) Get(ctx context.Context, scrID PredefinedScreenerModule, opts *Options) (*Result, error) {
	// Build URL with query params
	params := url.Values{}
	params.Set("scrIds", string(scrID))

	// Apply defaults and options
	count := 25 // default count
	if opts != nil {
		if opts.Count > 0 {
			count = opts.Count
		}
	}
	params.Set("count", fmt.Sprintf("%d", count))

	fullURL := fmt.Sprintf("${YF_QUERY_HOST}/v1/finance/screener/predefined/saved?%s", params.Encode())

	// Fetch with crumb authentication
	body, err := s.fetcher.Fetch(ctx, fetch.FetchConfig{URL: fullURL, NeedsCrumb: true})
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
		return nil, fmt.Errorf("parsing screener response: %w", err)
	}

	// Check for API error
	if raw.Finance.Error != nil {
		return nil, &errors.HTTPError{StatusCode: 200, Body: fmt.Sprintf("screener API error: %v", raw.Finance.Error)}
	}

	// Return first result (there should be exactly one)
	if len(raw.Finance.Result) == 0 {
		return nil, &errors.HTTPError{StatusCode: 200, Body: "no screener result returned"}
	}

	return &raw.Finance.Result[0], nil
}
