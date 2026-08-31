// Package yahoofinance provides a Go client for the Yahoo Finance API.
//
// The client uses a service-based architecture where each API module
// (chart, quote, historical, etc.) is a separate sub-package with its
// own Service type. Create a Client with NewClient, then access modules
// via the service pointers:
//
//	client := yahoofinance.NewClient(&yahoofinance.Config{})
//	result, err := client.Chart.Get(ctx, "AAPL", &chart.Options{...})
package yahoofinance

import (
	"github.com/dimpu/yfinance/chart"
	"github.com/dimpu/yfinance/fundamentals"
	"github.com/dimpu/yfinance/historical"
	"github.com/dimpu/yfinance/insights"
	"github.com/dimpu/yfinance/options"
	"github.com/dimpu/yfinance/quote"
	"github.com/dimpu/yfinance/quotesummary"
	"github.com/dimpu/yfinance/recommendations"
	"github.com/dimpu/yfinance/screener"
	"github.com/dimpu/yfinance/search"
	"github.com/dimpu/yfinance/trending"

	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

// Re-export shared types for convenience.
type (
	YahooDate      = types.YahooDate
	TwoNumberRange = types.TwoNumberRange
	FetchOptions   = types.FetchOptions
	ValidationOpts = types.ValidationOpts
	Logger         = types.Logger
)

// Re-export error types.
type (
	BadRequestError       = errors.BadRequestError
	HTTPError             = errors.HTTPError
	InvalidOptionsError   = errors.InvalidOptionsError
	FailedValidationError = errors.FailedValidationError
)

// Client is the main entry point for the Yahoo Finance API.
type Client struct {
	Chart           *chart.Service
	Quote           *quote.Service
	Historical      *historical.Service
	Options         *options.Service
	Fundamentals    *fundamentals.Service
	Insights        *insights.Service
	Screener        *screener.Service
	Search          *search.Service
	Trending        *trending.Service
	Recommendations *recommendations.Service
	QuoteSummary    *quotesummary.Service
}

// NewClient creates a new Client with the given configuration.
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg = applyDefaults(cfg)

	fetcher := fetch.NewFetcher(fetch.Config{
		QueryHost:    cfg.QueryHost,
		HTTPClient:   cfg.HTTPClient,
		Concurrency:  cfg.Concurrency,
		Logger:       cfg.Logger,
		FetchOptions: cfg.FetchOptions,
	})

	chartSvc := chart.NewService(fetcher)
	quoteSvc := quote.NewService(fetcher)
	historicalSvc := historical.NewService(chartSvc)
	optionsSvc := options.NewService(fetcher)
	fundamentalsSvc := fundamentals.NewService(fetcher)
	insightsSvc := insights.NewService(fetcher)
	screenerSvc := screener.NewService(fetcher)
	searchSvc := search.NewService(fetcher)
	trendingSvc := trending.NewService(fetcher)
	recommendationsSvc := recommendations.NewService(fetcher)
	quoteSummarySvc := quotesummary.NewService(fetcher)

	return &Client{
		Chart:           chartSvc,
		Quote:           quoteSvc,
		Historical:      historicalSvc,
		Options:         optionsSvc,
		Fundamentals:    fundamentalsSvc,
		Insights:        insightsSvc,
		Screener:        screenerSvc,
		Search:          searchSvc,
		Trending:        trendingSvc,
		Recommendations: recommendationsSvc,
		QuoteSummary:    quoteSummarySvc,
	}
}
