package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// QuoteSummary fetches detailed summary data for a symbol using the
// /v10/finance/quoteSummary endpoint. The modules parameter controls which
// sub-modules are returned; it defaults to ["price", "summaryDetail"] and
// supports "all" to request every available module.
func (c *Client) QuoteSummary(ctx context.Context, symbol string, opts *QuoteSummaryOptions) (*QuoteSummaryResult, error) {
	if symbol == "" {
		return nil, &InvalidOptionsError{Field: "symbol", Msg: "symbol is required"}
	}

	modules := resolveModules(opts)
	modulesParam := strings.Join(modules, ",")

	u := fmt.Sprintf("${YF_QUERY_HOST}/v10/finance/quoteSummary/{symbol}?modules=%s", modulesParam)

	body, err := c.fetch(ctx, fetchConfig{url: u, needsCrumb: true, symbol: symbol}, nil)
	if err != nil {
		return nil, err
	}

	var raw quoteSummaryResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing quoteSummary response: %w", err)
	}

	if raw.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("quoteSummary API error: %v", raw.QuoteSummary.Error)
	}

	if len(raw.QuoteSummary.Result) == 0 {
		return nil, &quoteSummaryParseError{Msg: "empty result array"}
	}

	return &raw.QuoteSummary.Result[0], nil
}
