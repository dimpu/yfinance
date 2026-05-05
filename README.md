# yahoo-finance

Go client for Yahoo Finance API. Port of [yahoo-finance2](https://github.com/gadicc/yahoo-finance2).

## Install

```bash
go get github.com/dimpu/yahoo-finance
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    yf "github.com/dimpu/yahoo-finance"
)

func main() {
    client := yf.NewClient(nil)
    ctx := context.Background()

    // Get real-time quotes
    quotes, err := client.Quote(ctx, []string{"AAPL", "GOOGL"}, nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, q := range quotes {
        fmt.Printf("%s: $%.2f (%s)\n", q.Symbol, *q.RegularMarketPrice, q.MarketState)
    }
}
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
quotes, err := client.Quote(ctx, []string{"AAPL", "MSFT"}, &yf.QuoteOptions{
    Fields: []string{"regularMarketPrice", "fiftyTwoWeekHigh"},
})
```

Needs crumb auth. Returns `[]yf.Quote` with 80+ fields including price, volume, market cap, 52-week range, EPS, P/E ratios, dividends, and more.

### Chart — Historical OHLCV data

```go
result, err := client.Chart(ctx, "AAPL", &yf.ChartOptions{
    Period1:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    Period2:        time.Now(),
    Interval:       "1d",
    IncludePrePost: true,
})
// result.Meta     — currency, exchange, valid ranges
// result.Quotes   — []ChartQuote with date, open, high, low, close, volume, adjclose
// result.Events   — dividends and splits
```

Valid intervals: `1m`, `2m`, `5m`, `15m`, `30m`, `60m`, `90m`, `1h`, `1d`, `5d`, `1wk`, `1mo`, `3mo`

### Historical — Price/dividend/split history

```go
rows, err := client.Historical(ctx, "AAPL", &yf.HistoricalOptions{
    Period1:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    Interval:             "1d",
    IncludeAdjustedClose: true,
})

dividends, err := client.HistoricalDividends(ctx, "AAPL", &yf.HistoricalOptions{
    Period1: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
})

splits, err := client.HistoricalSplits(ctx, "AAPL", &yf.HistoricalOptions{
    Period1: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
})
```

Wraps Chart internally. Valid intervals: `1d`, `1wk`, `1mo`

### Search — Find symbols and news

```go
result, err := client.Search(ctx, "apple", &yf.SearchOptions{
    QuotesCount: 5,
    NewsCount:   3,
})
// result.Quotes — matching symbols with exchange, type, score
// result.News   — related news articles
```

### QuoteSummary — Comprehensive financial data

```go
result, err := client.QuoteSummary(ctx, "AAPL", &yf.QuoteSummaryOptions{
    Modules: []string{"price", "summaryDetail", "financialData", "defaultKeyStatistics"},
})

// Or request all 33 modules:
result, err := client.QuoteSummary(ctx, "AAPL", &yf.QuoteSummaryOptions{
    Modules: []string{"all"},
})
```

Needs crumb auth. Available modules: `assetProfile`, `balanceSheetHistory`, `balanceSheetHistoryQuarterly`, `calendarEvents`, `cashflowStatementHistory`, `cashflowStatementHistoryQuarterly`, `defaultKeyStatistics`, `earnings`, `earningsHistory`, `earningsTrend`, `financialData`, `fundOwnership`, `fundPerformance`, `fundProfile`, `incomeStatementHistory`, `incomeStatementHistoryQuarterly`, `indexTrend`, `industryTrend`, `insiderHolders`, `insiderTransactions`, `institutionOwnership`, `majorDirectHolders`, `majorHoldersBreakdown`, `netSharePurchaseActivity`, `price`, `quoteType`, `recommendationTrend`, `secFilings`, `sectorTrend`, `summaryDetail`, `summaryProfile`, `topHoldings`, `upgradeDowngradeHistory`

### Options — Options chain data

```go
result, err := client.Options(ctx, "AAPL", &yf.OptionsOptions{
    Date: &expirationDate,
})
// result.Strikes          — available strike prices
// result.ExpirationDates  — available expiration dates
// result.Options[0].Calls — call options
// result.Options[0].Puts  — put options
```

Needs crumb auth.

### Screener — Predefined stock screeners

```go
result, err := client.Screener(ctx, yf.ScreenerDayGainers, &yf.ScreenerOptions{
    Count: 10,
})
```

Needs crumb auth. 15 predefined screeners:

| Constant | ID |
|----------|-----|
| `ScreenerAggressiveSmallCaps` | aggressive_small_caps |
| `ScreenerConservativeForeignFunds` | conservative_foreign_funds |
| `ScreenerDayGainers` | day_gainers |
| `ScreenerDayLosers` | day_losers |
| `ScreenerGrowthTechStocks` | growth_technology_stocks |
| `ScreenerHighYieldBond` | high_yield_bond |
| `ScreenerMostActives` | most_actives |
| `ScreenerMostShorted` | most_shorted_stocks |
| `ScreenerPortfolioAnchors` | portfolio_anchors |
| `ScreenerSmallCapGainers` | small_cap_gainers |
| `ScreenerSolidLargeGrowthFunds` | solid_large_growth_funds |
| `ScreenerSolidMidcapGrowthFunds` | solid_midcap_growth_funds |
| `ScreenerTopMutualFunds` | top_mutual_funds |
| `ScreenerUndervaluedGrowth` | undervalued_growth_stocks |
| `ScreenerUndervaluedLargeCaps` | undervalued_large_caps |

### RecommendationsBySymbol — Related stock recommendations

```go
results, err := client.RecommendationsBySymbol(ctx, []string{"AAPL"}, nil)
// results[0].RecommendedSymbols — related symbols with scores
```

### Insights — Analyst insights and research

```go
result, err := client.Insights(ctx, "AAPL", &yf.InsightsOptions{
    ReportsCount: 5,
})
// result.Recommendation   — analyst rating (BUY/SELL/HOLD)
// result.CompanySnapshot  — sector comparison
// result.InstrumentInfo   — technicals and valuation
// result.SigDevs          — significant developments
```

### FundamentalsTimeSeries — Financial statements over time

```go
results, err := client.FundamentalsTimeSeries(ctx, "AAPL", &yf.FundamentalsTimeSeriesOptions{
    Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
    Module:  "financials",
    Type:    "quarterly",
})
// Each result has Date, Type, PeriodType, and Fields map[string]*float64
// with keys like TotalRevenue, NetIncome, EBITDA, etc.
```

Modules: `financials`, `balance-sheet`, `cash-flow`, `all`
Types: `quarterly`, `annual`, `trailing`

### TrendingSymbols — Trending tickers by region

```go
result, err := client.TrendingSymbols(ctx, "US", &yf.TrendingOptions{
    Count: 10,
})
// result.Quotes — trending symbols
```

## Error Handling

```go
quotes, err := client.Quote(ctx, []string{"AAPL"}, nil)
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
        // Other errors (network, context cancellation, etc.)
    }
}
```

## Authentication

The client automatically handles Yahoo Finance's cookie/crumb authentication for endpoints that require it (Quote, Options, QuoteSummary, Screener). On first request to a protected endpoint, the client:

1. Visits `finance.yahoo.com` to collect cookies
2. Handles GDPR consent redirects if needed
3. Fetches a crumb token from Yahoo's API
4. Caches the crumb for subsequent requests
5. Re-fetches automatically on 401 (expired crumb)

## Concurrency

Requests are rate-limited via a semaphore (default: 4 concurrent). Configure via `Config.Concurrency`. Context cancellation is respected — pass a context with timeout for deadlines:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
quotes, err := client.Quote(ctx, []string{"AAPL"}, nil)
```

## Running Tests

```bash
go test -v -race ./...
```

## License

MIT
