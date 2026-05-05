package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Insights fetches Yahoo Finance insights for a symbol, including technical
// analysis, valuation, and analyst recommendations. No crumb authentication
// is required for this endpoint.
func (c *Client) Insights(ctx context.Context, symbol string, opts *InsightsOptions) (*InsightsResult, error) {
	if symbol == "" {
		return nil, &InvalidOptionsError{Field: "symbol", Msg: "symbol is required"}
	}

	// Build URL with query params
	params := url.Values{}
	params.Set("symbol", symbol)

	// Apply defaults and options
	lang := "en-US"
	region := "US"
	if opts != nil {
		if opts.Lang != "" {
			lang = opts.Lang
		}
		if opts.Region != "" {
			region = opts.Region
		}
		if opts.ReportsCount > 0 {
			params.Set("reportsCount", fmt.Sprintf("%d", opts.ReportsCount))
		}
	}
	params.Set("lang", lang)
	params.Set("region", region)

	fullURL := fmt.Sprintf("${YF_QUERY_HOST}/ws/insights/v2/finance/insights?%s", params.Encode())

	// Fetch without crumb
	body, err := c.fetch(ctx, fetchConfig{url: fullURL, needsCrumb: false}, nil)
	if err != nil {
		return nil, err
	}

	// Parse response - wrapped in {"finance":{"result":{...}}}
	var wrapper struct {
		Finance struct {
			Result *InsightsResult `json:"result"`
			Error  *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"finance"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("parsing insights response: %w", err)
	}

	if wrapper.Finance.Error != nil && wrapper.Finance.Error.Code != "" {
		return nil, &HTTPError{StatusCode: 200, Body: wrapper.Finance.Error.Description}
	}

	if wrapper.Finance.Result == nil {
		return nil, &HTTPError{StatusCode: 200, Body: "empty result"}
	}

	return wrapper.Finance.Result, nil
}
