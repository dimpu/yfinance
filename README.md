# yahoo-finance

[![Go Reference](https://pkg.go.dev/badge/github.com/dimpu/yfinance.svg)](https://pkg.go.dev/github.com/dimpu/yfinance)
[![Go Report Card](https://goreportcard.com/badge/github.com/dimpu/yfinance)](https://goreportcard.com/report/github.com/dimpu/yfinance)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go client for the Yahoo Finance API. Port of [yahoo-finance2](https://github.com/gadicc/yahoo-finance2).

**Requires Go 1.25+**

## Install

```bash
go get github.com/dimpu/yfinance
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    yf "github.com/dimpu/yfinance"
    "github.com/dimpu/yfinance/chart"
)

func main() {
    client := yf.NewClient(nil)
    ctx := context.Background()

    // Get chart data
    result, err := client.Chart.Get(ctx, "AAPL", &chart.Options{
        Period1:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        Interval: "1d",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Symbol: %s, Quotes: %d\n", result.Meta.Symbol, len(result.Quotes))
}
```

## Architecture

Each API module is a separate sub-package with its own `Service` type:

```
github.com/dimpu/yfinance/          # Client + Config + re-exported types
├── chart/                                # Chart service (OHLCV data)
├── quote/                                # Quote service (real-time prices)
├── historical/                           # Historical service (delegates to chart)
├── options/                              # Options chain service
├── fundamentals/                         # Fundamentals time series service
├── insights/                             # Analyst insights service
├── screener/                             # Predefined screener service
├── search/                               # Symbol search service
├── trending/                             # Trending symbols service
├── recommendations/                      # Related stock recommendations
├── quotesummary/                         # Comprehensive financial data
└── internal/                             # Shared infrastructure (not importable)
    ├── types/                            # YahooDate, TwoNumberRange, Logger, etc.
    ├── errors/                           # BadRequestError, HTTPError, etc.
    ├── fetch/                            # HTTP client, crumb auth, semaphore
    └── validate/                         # Struct validation
```

## Configuration

```go
client := yf.NewClient(&yf.Config{
    QueryHost:   "query2.finance.yahoo.com", // default
    Concurrency: 4,                          // max concurrent requests, default 4
    Logger:      myLogger,                   // implement yf.Logger interface
    HTTPClient:  customHTTPClient,           // custom *http.Client
    FetchOptions: &yf.FetchOptions{
        Headers: map[string]string{
            "User-Agent": "my-app/1.0",
        },
    },
})
```

## API Methods

### Quote — Real-time price data

```go
import "github.com/dimpu/yfinance/quote"

quotes, err := client.Quote.Get(ctx, []string{"AAPL", "MSFT"}, &quote.Options{
    Fields: []string{"regularMarketPrice", "fiftyTwoWeekHigh"},
})
```

Needs crumb auth. Returns `[]quote.Quote` with 80+ fields.

### Chart — Historical OHLCV data

```go
import "github.com/dimpu/yfinance/chart"

result, err := client.Chart.Get(ctx, "AAPL", &chart.Options{
    Period1:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    Period2:        time.Now(),
    Interval:       "1d",
    IncludePrePost: true,
})
```

Valid intervals: `1m`, `2m`, `5m`, `15m`, `30m`, `60m`, `90m`, `1h`, `1d`, `5d`, `1wk`, `1mo`, `3mo`

### Historical — Price/dividend/split history

```go
import "github.com/dimpu/yfinance/historical"

rows, err := client.Historical.Get(ctx, "AAPL", &historical.Options{
    Period1:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    Interval:             "1d",
    IncludeAdjustedClose: true,
})

dividends, err := client.Historical.Dividends(ctx, "AAPL", &historical.Options{
    Period1: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
})

splits, err := client.Historical.Splits(ctx, "AAPL", &historical.Options{
    Period1: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
})
```

Wraps Chart internally. Valid intervals: `1d`, `1wk`, `1mo`

### Search — Find symbols and news

```go
import "github.com/dimpu/yfinance/search"

result, err := client.Search.Get(ctx, "apple", &search.Options{
    QuotesCount: 5,
    NewsCount:   3,
})
```

### QuoteSummary — Comprehensive financial data

```go
import "github.com/dimpu/yfinance/quotesummary"

result, err := client.QuoteSummary.Get(ctx, "AAPL", &quotesummary.Options{
    Modules: []string{"price", "summaryDetail", "financialData"},
})

// All 33 modules:
result, err := client.QuoteSummary.Get(ctx, "AAPL", &quotesummary.Options{
    Modules: []string{"all"},
})
```

Needs crumb auth.

### Options — Options chain data

```go
import "github.com/dimpu/yfinance/options"

result, err := client.Options.Get(ctx, "AAPL", &options.Options{
    Date: &expirationDate,
})
```

Needs crumb auth.

### Screener — Predefined stock screeners

```go
import "github.com/dimpu/yfinance/screener"

result, err := client.Screener.Get(ctx, screener.ScreenerDayGainers, &screener.Options{
    Count: 10,
})
```

Needs crumb auth. 15 predefined screeners available as constants in the `screener` package.

### Recommendations — Related stock recommendations

```go
import "github.com/dimpu/yfinance/recommendations"

results, err := client.Recommendations.Get(ctx, []string{"AAPL"}, nil)
```

### Insights — Analyst insights and research

```go
import "github.com/dimpu/yfinance/insights"

result, err := client.Insights.Get(ctx, "AAPL", &insights.Options{
    ReportsCount: 5,
})
```

### Fundamentals — Financial statements over time

```go
import "github.com/dimpu/yfinance/fundamentals"

results, err := client.Fundamentals.Get(ctx, "AAPL", &fundamentals.Options{
    Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
    Module:  "financials",
    Type:    "quarterly",
})
```

Modules: `financials`, `balance-sheet`, `cash-flow`, `all`
Types: `quarterly`, `annual`, `trailing`

### Trending — Trending tickers by region

```go
import "github.com/dimpu/yfinance/trending"

result, err := client.Trending.Get(ctx, "US", &trending.Options{
    Count: 10,
})
```

## Error Handling

Error types are re-exported from the root package for convenience:

```go
quotes, err := client.Quote.Get(ctx, []string{"AAPL"}, nil)
if err != nil {
    switch e := err.(type) {
    case *yf.BadRequestError:
        // HTTP 400 from Yahoo
    case *yf.HTTPError:
        // Non-OK HTTP status, e.StatusCode and e.Body available
    case *yf.InvalidOptionsError:
        // Invalid input options, e.Field and e.Msg available
    case *yf.FailedValidationError:
        // Response validation failed, e.Errors has details
    default:
        // Other errors
    }
}
```

## Authentication

The client automatically handles Yahoo Finance's cookie/crumb authentication for endpoints that require it. On first request to a protected endpoint, the client:

1. Visits `finance.yahoo.com` to collect cookies
2. Handles GDPR consent redirects if needed
3. Fetches a crumb token from Yahoo's API
4. Caches the crumb for subsequent requests
5. Re-fetches automatically on 401 (expired crumb)

## Concurrency

Requests are rate-limited via a semaphore (default: 4 concurrent). Configure via `Config.Concurrency`.

## CLI

An optional CLI tool is included for quick lookups from the terminal:

```bash
# Build and install
go install github.com/dimpu/yfinance/cmd/yf@latest

# Usage
yf quote AAPL MSFT
yf chart AAPL --interval 1d --period 1mo
yf options AAPL
yf trending --region US
yf search apple
```

## Running Tests

```bash
go test -v -race ./...
```

## Documentation

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/dimpu/yfinance).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/my-feature`)
3. Commit your changes (`git commit -m 'feat: add my feature'`)
4. Push to the branch (`git push origin feat/my-feature`)
5. Open a Pull Request

Please ensure tests pass (`go test -v -race ./...`) before submitting.

## License

[MIT](LICENSE)
