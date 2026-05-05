package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Screener fetches predefined screener results from Yahoo Finance.
// The screener ID must be one of the PredefinedScreenerModule constants.
// This endpoint requires crumb authentication.
func (c *Client) Screener(ctx context.Context, scrID PredefinedScreenerModule, opts *ScreenerOptions) (*ScreenerResult, error) {
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
	body, err := c.fetch(ctx, fetchConfig{url: fullURL, needsCrumb: true}, nil)
	if err != nil {
		return nil, err
	}

	// Parse response: {"finance":{"result":[...],"error":null}}
	var raw struct {
		Finance struct {
			Result []ScreenerResult `json:"result"`
			Error  interface{}      `json:"error"`
		} `json:"finance"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing screener response: %w", err)
	}

	// Check for API error
	if raw.Finance.Error != nil {
		return nil, fmt.Errorf("screener API error: %v", raw.Finance.Error)
	}

	// Return first result (there should be exactly one)
	if len(raw.Finance.Result) == 0 {
		return nil, fmt.Errorf("no screener result returned")
	}

	return &raw.Finance.Result[0], nil
}
