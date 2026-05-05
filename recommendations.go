package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// RecommendationsBySymbol fetches recommended symbols for the given list of symbols.
// No crumb authentication is required for this endpoint.
func (c *Client) RecommendationsBySymbol(ctx context.Context, symbols []string, opts *RecommendationsOptions) ([]RecommendationsResult, error) {
	if len(symbols) == 0 {
		return nil, &InvalidOptionsError{Field: "symbols", Msg: "at least one symbol is required"}
	}

	// Build URL: symbols are comma-joined and placed in the path
	joinedSymbols := url.PathEscape(strings.Join(symbols, ","))
	u := fmt.Sprintf("${YF_QUERY_HOST}/v6/finance/recommendationsbysymbol/%s", joinedSymbols)

	// Fetch without crumb
	body, err := c.fetch(ctx, fetchConfig{url: u, needsCrumb: false}, nil)
	if err != nil {
		return nil, err
	}

	// Parse response: {"finance":{"result":[...],"error":null}}
	var raw struct {
		Finance struct {
			Result []RecommendationsResult `json:"result"`
			Error  interface{}             `json:"error"`
		} `json:"finance"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing recommendations response: %w", err)
	}

	// Check for API error
	if raw.Finance.Error != nil {
		return nil, fmt.Errorf("recommendations API error: %v", raw.Finance.Error)
	}

	return raw.Finance.Result, nil
}
