package yahoofinance

import (
	"context"
	"fmt"
	"time"
)

// mapHistoricalEvents maps the Historical events param to the Chart events param.
// "history" (default) maps to "" (no events in chart response, just OHLCV).
// "dividends" maps to "div", "split" maps to "split".
func mapHistoricalEvents(events string) string {
	switch events {
	case "history", "":
		return ""
	case "dividends":
		return "div"
	case "split":
		return "split"
	default:
		return events
	}
}

// Historical returns price history for the given symbol.
// For dividends or splits, use HistoricalDividends or HistoricalSplits instead.
func (c *Client) Historical(ctx context.Context, symbol string, opts *HistoricalOptions) ([]HistoricalRowHistory, error) {
	if opts == nil {
		opts = &HistoricalOptions{}
	}

	// Validate options
	if err := validateOptions(opts); err != nil {
		return nil, &InvalidOptionsError{Field: "HistoricalOptions", Msg: err.Error()}
	}

	chartOpts := historicalToChartOpts(opts)

	result, err := c.Chart(ctx, symbol, chartOpts)
	if err != nil {
		return nil, err
	}

	return chartToHistoricalRows(result, opts.IncludeAdjustedClose), nil
}

// HistoricalDividends returns dividend history for the given symbol.
func (c *Client) HistoricalDividends(ctx context.Context, symbol string, opts *HistoricalOptions) ([]HistoricalRowDividend, error) {
	if opts == nil {
		opts = &HistoricalOptions{}
	}

	// Force events to "dividends" for the chart call
	optsForChart := *opts
	optsForChart.Events = "dividends"

	if err := validateOptions(&optsForChart); err != nil {
		return nil, &InvalidOptionsError{Field: "HistoricalOptions", Msg: err.Error()}
	}

	chartOpts := historicalToChartOpts(&optsForChart)

	result, err := c.Chart(ctx, symbol, chartOpts)
	if err != nil {
		return nil, err
	}

	if result.Events == nil || len(result.Events.Dividends) == 0 {
		return nil, nil
	}

	rows := make([]HistoricalRowDividend, 0, len(result.Events.Dividends))
	for _, div := range result.Events.Dividends {
		rows = append(rows, HistoricalRowDividend{
			Date:      time.Unix(div.Date, 0),
			Dividends: div.Amount,
		})
	}
	return rows, nil
}

// HistoricalSplits returns stock split history for the given symbol.
func (c *Client) HistoricalSplits(ctx context.Context, symbol string, opts *HistoricalOptions) ([]HistoricalRowStockSplit, error) {
	if opts == nil {
		opts = &HistoricalOptions{}
	}

	// Force events to "split" for the chart call
	optsForChart := *opts
	optsForChart.Events = "split"

	if err := validateOptions(&optsForChart); err != nil {
		return nil, &InvalidOptionsError{Field: "HistoricalOptions", Msg: err.Error()}
	}

	chartOpts := historicalToChartOpts(&optsForChart)

	result, err := c.Chart(ctx, symbol, chartOpts)
	if err != nil {
		return nil, err
	}

	if result.Events == nil || len(result.Events.Splits) == 0 {
		return nil, nil
	}

	rows := make([]HistoricalRowStockSplit, 0, len(result.Events.Splits))
	for _, split := range result.Events.Splits {
		rows = append(rows, HistoricalRowStockSplit{
			Date:        time.Unix(split.Date, 0),
			StockSplits: split.SplitRatio,
		})
	}
	return rows, nil
}

// historicalToChartOpts converts HistoricalOptions to ChartOptions.
func historicalToChartOpts(opts *HistoricalOptions) *ChartOptions {
	chartOpts := &ChartOptions{
		Period1:        opts.Period1,
		Period2:        opts.Period2,
		Interval:       opts.Interval,
		Events:         mapHistoricalEvents(opts.Events),
		IncludePrePost: false,
	}

	// When IncludeAdjustedClose is true, request events so adjclose is included
	if opts.IncludeAdjustedClose {
		chartOpts.Events = "div|split|earn"
	}

	return chartOpts
}

// chartToHistoricalRows converts ChartResult quotes to HistoricalRowHistory slice,
// filtering out all-null rows (where OHLCV are all nil except date).
func chartToHistoricalRows(result *ChartResult, includeAdjClose bool) []HistoricalRowHistory {
	if len(result.Quotes) == 0 {
		return nil
	}

	rows := make([]HistoricalRowHistory, 0, len(result.Quotes))
	for _, q := range result.Quotes {
		// Filter out all-null rows: if all OHLCV are nil, skip
		if q.Open == nil && q.High == nil && q.Low == nil && q.Close == nil && q.Volume == nil {
			continue
		}

		row := HistoricalRowHistory{
			Date:   q.Date,
			Open:   q.Open,
			High:   q.High,
			Low:    q.Low,
			Close:  q.Close,
			Volume: q.Volume,
		}

		// Rename adjclose -> adjClose (the field name change from ChartQuote.AdjClose)
		if includeAdjClose && q.AdjClose != nil {
			row.AdjClose = q.AdjClose
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil
	}

	return rows
}

// Ensure HistoricalOptions provides a helpful error for required Period1.
func validateHistoricalOpts(opts *HistoricalOptions) error {
	if opts.Period1.IsZero() {
		return fmt.Errorf("Period1 is required")
	}
	return nil
}
