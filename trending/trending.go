package trending

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
)

// Service provides access to the Yahoo Finance Trending API.
type Service struct {
	fetcher *fetch.Fetcher
}

// NewService creates a new Trending service.
func NewService(f *fetch.Fetcher) *Service {
	return &Service{fetcher: f}
}

// Get fetches trending symbols for the given region.
// No crumb authentication is required for this endpoint.
func (s *Service) Get(ctx context.Context, region string, opts *Options) (*Result, error) {
	if region == "" {
		return nil, &errors.InvalidOptionsError{Field: "region", Msg: "region is required"}
	}

	// Build URL with region in path and optional query params
	params := url.Values{}
	if opts != nil {
		if opts.Lang != "" {
			params.Set("lang", opts.Lang)
		}
		if opts.Region != "" {
			params.Set("region", opts.Region)
		}
		if opts.Count > 0 {
			params.Set("count", fmt.Sprintf("%d", opts.Count))
		}
	}

	encodedRegion := url.PathEscape(region)
	u := fmt.Sprintf("${YF_QUERY_HOST}/v1/finance/trending/%s", encodedRegion)
	if len(params) > 0 {
		u = fmt.Sprintf("%s?%s", u, params.Encode())
	}

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
		return nil, fmt.Errorf("parsing trending response: %w", err)
	}

	// Check for API error
	if raw.Finance.Error != nil {
		return nil, &errors.HTTPError{StatusCode: 200, Body: fmt.Sprintf("trending API error: %v", raw.Finance.Error)}
	}

	if len(raw.Finance.Result) == 0 {
		return nil, &errors.HTTPError{StatusCode: 200, Body: "no trending results returned"}
	}

	return &raw.Finance.Result[0], nil
}
