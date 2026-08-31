package quotesummary

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
)

// Service provides access to the Yahoo Finance quoteSummary API.
type Service struct {
	fetcher *fetch.Fetcher
}

// NewService creates a new quotesummary Service.
func NewService(f *fetch.Fetcher) *Service {
	return &Service{fetcher: f}
}

// Get fetches detailed summary data for a symbol using the
// /v10/finance/quoteSummary endpoint.
func (s *Service) Get(ctx context.Context, symbol string, opts *Options) (*Result, error) {
	if symbol == "" {
		return nil, &errors.InvalidOptionsError{Field: "symbol", Msg: "symbol is required"}
	}

	modules := resolveModules(opts)
	modulesParam := strings.Join(modules, ",")

	u := fmt.Sprintf("${YF_QUERY_HOST}/v10/finance/quoteSummary/{symbol}?modules=%s", modulesParam)

	body, err := s.fetcher.Fetch(ctx, fetch.FetchConfig{URL: u, NeedsCrumb: true, Symbol: symbol})
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
