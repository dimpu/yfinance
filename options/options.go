package options

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/dimpu/yfinance/internal/fetch"
)

// Service provides access to the Yahoo Finance options API.
type Service struct {
	fetcher *fetch.Fetcher
}

// NewService creates a new options Service.
func NewService(f *fetch.Fetcher) *Service {
	return &Service{fetcher: f}
}

// Get fetches the options chain for the given symbol.
// If opts.Date is set, it retrieves the options chain for that specific expiration date;
// otherwise, the nearest expiration is returned.
func (s *Service) Get(ctx context.Context, symbol string, opts *Options) (*Result, error) {
	u := fmt.Sprintf("${YF_QUERY_HOST}/v7/finance/options/{symbol}")

	params := url.Values{}
	if opts != nil {
		if opts.Date != nil {
			params.Set("date", fmt.Sprintf("%d", opts.Date.Unix()))
		}
		if opts.Formatted {
			params.Set("formatted", "true")
		}
		if opts.Lang != "" {
			params.Set("lang", opts.Lang)
		}
		if opts.Region != "" {
			params.Set("region", opts.Region)
		}
	}
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	body, err := s.fetcher.Fetch(ctx, fetch.FetchConfig{URL: u, NeedsCrumb: true, Symbol: symbol})
	if err != nil {
		return nil, err
	}

	var raw struct {
		OptionChain struct {
			Result []Result      `json:"result"`
			Error  interface{}   `json:"error"`
		} `json:"optionChain"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing options response: %w", err)
	}

	if raw.OptionChain.Error != nil {
		return nil, fmt.Errorf("options API error: %v", raw.OptionChain.Error)
	}

	if len(raw.OptionChain.Result) == 0 {
		return nil, fmt.Errorf("no options result returned")
	}

	return &raw.OptionChain.Result[0], nil
}
