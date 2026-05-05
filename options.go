package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Options fetches the options chain for the given symbol.
// If opts.Date is set, it retrieves the options chain for that specific expiration date;
// otherwise, the nearest expiration is returned.
func (c *Client) Options(ctx context.Context, symbol string, opts *OptionsOptions) (*OptionsResult, error) {
	// Build URL with symbol in path
	u := fmt.Sprintf("${YF_QUERY_HOST}/v7/finance/options/{symbol}")

	// Add optional query params
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

	// Fetch with crumb authentication
	body, err := c.fetch(ctx, fetchConfig{url: u, needsCrumb: true, symbol: symbol}, nil)
	if err != nil {
		return nil, err
	}

	// Parse response: {"optionChain":{"result":[...],"error":null}}
	var raw struct {
		OptionChain struct {
			Result []OptionsResult `json:"result"`
			Error  interface{}     `json:"error"`
		} `json:"optionChain"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing options response: %w", err)
	}

	// Check for API error
	if raw.OptionChain.Error != nil {
		return nil, fmt.Errorf("options API error: %v", raw.OptionChain.Error)
	}

	if len(raw.OptionChain.Result) == 0 {
		return nil, fmt.Errorf("no options result returned")
	}

	return &raw.OptionChain.Result[0], nil
}
