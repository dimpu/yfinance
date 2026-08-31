package historical

import (
	"context"
	"time"

	"github.com/dimpu/yfinance/chart"
	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/validate"
)

// Service provides access to the Yahoo Finance historical data API.
type Service struct {
	chart *chart.Service
}

// NewService creates a new historical Service.
func NewService(chartService *chart.Service) *Service {
	return &Service{chart: chartService}
}

// Get returns price history for the given symbol.
func (s *Service) Get(ctx context.Context, symbol string, opts *Options) ([]RowHistory, error) {
	if opts == nil {
		opts = &Options{}
	}

	if err := validate.Struct(opts); err != nil {
		return nil, &errors.InvalidOptionsError{Field: "HistoricalOptions", Msg: err.Error()}
	}

	chartOpts := historicalToChartOpts(opts)

	result, err := s.chart.Get(ctx, symbol, chartOpts)
	if err != nil {
		return nil, err
	}

	return chartToHistoricalRows(result, opts.IncludeAdjustedClose), nil
}

// Dividends returns dividend history for the given symbol.
func (s *Service) Dividends(ctx context.Context, symbol string, opts *Options) ([]RowDividend, error) {
	if opts == nil {
		opts = &Options{}
	}

	optsForChart := *opts
	optsForChart.Events = "dividends"

	if err := validate.Struct(&optsForChart); err != nil {
		return nil, &errors.InvalidOptionsError{Field: "HistoricalOptions", Msg: err.Error()}
	}

	chartOpts := historicalToChartOpts(&optsForChart)

	result, err := s.chart.Get(ctx, symbol, chartOpts)
	if err != nil {
		return nil, err
	}

	if result.Events == nil || len(result.Events.Dividends) == 0 {
		return nil, nil
	}

	rows := make([]RowDividend, 0, len(result.Events.Dividends))
	for _, div := range result.Events.Dividends {
		rows = append(rows, RowDividend{
			Date:      time.Unix(div.Date, 0),
			Dividends: div.Amount,
		})
	}
	return rows, nil
}

// Splits returns stock split history for the given symbol.
func (s *Service) Splits(ctx context.Context, symbol string, opts *Options) ([]RowStockSplit, error) {
	if opts == nil {
		opts = &Options{}
	}

	optsForChart := *opts
	optsForChart.Events = "split"

	if err := validate.Struct(&optsForChart); err != nil {
		return nil, &errors.InvalidOptionsError{Field: "HistoricalOptions", Msg: err.Error()}
	}

	chartOpts := historicalToChartOpts(&optsForChart)

	result, err := s.chart.Get(ctx, symbol, chartOpts)
	if err != nil {
		return nil, err
	}

	if result.Events == nil || len(result.Events.Splits) == 0 {
		return nil, nil
	}

	rows := make([]RowStockSplit, 0, len(result.Events.Splits))
	for _, split := range result.Events.Splits {
		rows = append(rows, RowStockSplit{
			Date:        time.Unix(split.Date, 0),
			StockSplits: split.SplitRatio,
		})
	}
	return rows, nil
}

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

func historicalToChartOpts(opts *Options) *chart.Options {
	chartOpts := &chart.Options{
		Period1:        opts.Period1,
		Period2:        opts.Period2,
		Interval:       opts.Interval,
		Events:         mapHistoricalEvents(opts.Events),
		IncludePrePost: false,
	}

	if opts.IncludeAdjustedClose {
		chartOpts.Events = "div|split|earn"
	}

	return chartOpts
}

func chartToHistoricalRows(result *chart.Result, includeAdjClose bool) []RowHistory {
	if len(result.Quotes) == 0 {
		return nil
	}

	rows := make([]RowHistory, 0, len(result.Quotes))
	for _, q := range result.Quotes {
		if q.Open == nil && q.High == nil && q.Low == nil && q.Close == nil && q.Volume == nil {
			continue
		}

		row := RowHistory{
			Date:   q.Date,
			Open:   q.Open,
			High:   q.High,
			Low:    q.Low,
			Close:  q.Close,
			Volume: q.Volume,
		}

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
