package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Quote fetches real-time or near real-time quote data for the given symbols.
// It retrieves essential symbol information including current prices, market state,
// volume, and other key metrics for stocks, ETFs, options, cryptocurrencies, and
// other financial instruments.
func (c *Client) Quote(ctx context.Context, symbols []string, opts *QuoteOptions) ([]Quote, error) {
	if len(symbols) == 0 {
		return nil, &InvalidOptionsError{Field: "symbols", Msg: "at least one symbol is required"}
	}

	// Build URL
	joinedSymbols := strings.Join(symbols, ",")
	u := fmt.Sprintf("${YF_QUERY_HOST}/v7/finance/quote?symbols=%s", url.QueryEscape(joinedSymbols))

	// Add optional query params
	if opts != nil {
		params := url.Values{}
		if len(opts.Fields) > 0 {
			params.Set("fields", strings.Join(opts.Fields, ","))
		}
		if opts.Lang != "" {
			params.Set("lang", opts.Lang)
		}
		if opts.Region != "" {
			params.Set("region", opts.Region)
		}
		if len(params) > 0 {
			u += "&" + params.Encode()
		}
	}

	// Fetch with crumb authentication
	body, err := c.fetch(ctx, fetchConfig{url: u, needsCrumb: true}, nil)
	if err != nil {
		return nil, err
	}

	// Parse response
	var raw struct {
		QuoteResponse struct {
			Result []Quote    `json:"result"`
			Error  interface{} `json:"error"`
		} `json:"quoteResponse"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing quote response: %w", err)
	}

	// Check for API error
	if raw.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("quote API error: %v", raw.QuoteResponse.Error)
	}

	// Filter out results where quoteType == "NONE" (delisted symbols)
	results := make([]Quote, 0, len(raw.QuoteResponse.Result))
	for _, q := range raw.QuoteResponse.Result {
		if q.QuoteType != "NONE" {
			results = append(results, q)
		}
	}

	return results, nil
}
