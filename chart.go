package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// Chart fetches historical OHLCV data for the given symbol.
func (c *Client) Chart(ctx context.Context, symbol string, opts *ChartOptions) (*ChartResult, error) {
	// Apply defaults
	if opts == nil {
		opts = &ChartOptions{}
	}
	if opts.Period2.IsZero() {
		opts.Period2 = time.Now()
	}
	if opts.Interval == "" {
		opts.Interval = "1d"
	}
	if opts.Events == "" {
		opts.Events = "div|split|earn"
	}

	// Validate options
	if err := validateOptions(opts); err != nil {
		return nil, &InvalidOptionsError{Field: "ChartOptions", Msg: err.Error()}
	}

	// Build URL using ${YF_QUERY_HOST} and {symbol} placeholders for fetch() substitution
	u := fmt.Sprintf("${YF_QUERY_HOST}/v8/finance/chart/{symbol}?period1=%d&period2=%d&interval=%s",
		opts.Period1.Unix(),
		opts.Period2.Unix(),
		url.QueryEscape(opts.Interval),
	)

	// Add optional query params
	params := url.Values{}
	if opts.IncludePrePost {
		params.Set("includePrePost", "true")
	}
	if opts.Events != "" {
		params.Set("events", opts.Events)
	}
	if opts.Lang != "" {
		params.Set("lang", opts.Lang)
	}
	if opts.Return == "object" {
		params.Set("return", "object")
	}
	if len(params) > 0 {
		u += "&" + params.Encode()
	}

	// Fetch (no crumb needed for chart endpoint)
	body, err := c.fetch(ctx, fetchConfig{url: u, symbol: symbol}, nil)
	if err != nil {
		return nil, err
	}

	// Parse the raw Yahoo response
	var raw struct {
		Chart struct {
			Result []rawChartResult `json:"result"`
			Error  interface{}      `json:"error"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing chart response: %w", err)
	}
	if raw.Chart.Error != nil {
		return nil, fmt.Errorf("chart API error: %v", raw.Chart.Error)
	}
	if len(raw.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart result returned")
	}

	return parseChartResult(&raw.Chart.Result[0])
}

// rawChartResult mirrors the Yahoo JSON structure for chart data.
type rawChartResult struct {
	Meta       ChartMeta          `json:"meta"`
	Timestamp  []int64            `json:"timestamp"`
	Indicators rawChartIndicators `json:"indicators"`
	Events     *rawChartEvents    `json:"events,omitempty"`
}

type rawChartIndicators struct {
	Quote    []rawChartQuote `json:"quote"`
	AdjClose []struct {
		AdjClose []*float64 `json:"adjclose"`
	} `json:"adjclose"`
}

type rawChartQuote struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

type rawChartEvents struct {
	Dividends map[string]ChartEventDividend `json:"dividends,omitempty"`
	Splits    map[string]ChartEventSplit    `json:"splits,omitempty"`
}

// parseChartResult converts the raw Yahoo format into ChartResult.
func parseChartResult(raw *rawChartResult) (*ChartResult, error) {
	result := &ChartResult{
		Meta: raw.Meta,
	}

	// Build quotes from timestamp + indicators
	if len(raw.Indicators.Quote) == 0 {
		return result, nil
	}
	q := raw.Indicators.Quote[0]

	// Optional adjclose data
	var adjClose []*float64
	if len(raw.Indicators.AdjClose) > 0 {
		adjClose = raw.Indicators.AdjClose[0].AdjClose
	}

	quotes := make([]ChartQuote, len(raw.Timestamp))
	for i, ts := range raw.Timestamp {
		quotes[i] = ChartQuote{
			Date:   time.Unix(ts, 0),
			Open:   safeFloat64Ptr(q.Open, i),
			High:   safeFloat64Ptr(q.High, i),
			Low:    safeFloat64Ptr(q.Low, i),
			Close:  safeFloat64Ptr(q.Close, i),
			Volume: safeInt64Ptr(q.Volume, i),
		}
		if adjClose != nil && i < len(adjClose) {
			quotes[i].AdjClose = adjClose[i]
		}
	}
	result.Quotes = quotes

	// Parse events
	if raw.Events != nil {
		chartEvents := &ChartEvents{}
		if len(raw.Events.Dividends) > 0 {
			chartEvents.Dividends = make(map[int64]ChartEventDividend, len(raw.Events.Dividends))
			for _, v := range raw.Events.Dividends {
				chartEvents.Dividends[v.Date] = v
			}
		}
		if len(raw.Events.Splits) > 0 {
			chartEvents.Splits = make(map[int64]ChartEventSplit, len(raw.Events.Splits))
			for _, v := range raw.Events.Splits {
				chartEvents.Splits[v.Date] = v
			}
		}
		result.Events = chartEvents
	}

	return result, nil
}

// safeFloat64Ptr returns the pointer at index i, or nil if out of bounds.
func safeFloat64Ptr(slice []*float64, i int) *float64 {
	if i < len(slice) {
		return slice[i]
	}
	return nil
}

// safeInt64Ptr returns the pointer at index i, or nil if out of bounds.
func safeInt64Ptr(slice []*int64, i int) *int64 {
	if i < len(slice) {
		return slice[i]
	}
	return nil
}
