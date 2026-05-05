package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// TrendingSymbols fetches trending symbols for the given region.
// No crumb authentication is required for this endpoint.
func (c *Client) TrendingSymbols(ctx context.Context, region string, opts *TrendingOptions) (*TrendingResult, error) {
	if region == "" {
		return nil, &InvalidOptionsError{Field: "region", Msg: "region is required"}
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
	body, err := c.fetch(ctx, fetchConfig{url: u, needsCrumb: false}, nil)
	if err != nil {
		return nil, err
	}

	// Parse response: {"finance":{"result":[...],"error":null}}
	var raw struct {
		Finance struct {
			Result []TrendingResult `json:"result"`
			Error  interface{}      `json:"error"`
		} `json:"finance"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing trending response: %w", err)
	}

	// Check for API error
	if raw.Finance.Error != nil {
		return nil, fmt.Errorf("trending API error: %v", raw.Finance.Error)
	}

	if len(raw.Finance.Result) == 0 {
		return nil, fmt.Errorf("no trending results returned")
	}

	return &raw.Finance.Result[0], nil
}
