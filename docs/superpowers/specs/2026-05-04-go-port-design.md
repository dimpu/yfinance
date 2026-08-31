# Go Port of yahoo-finance2 — Design Spec

**Date:** 2026-05-04
**Module path:** `github.com/dimpu/yfinance`
**Source:** https://github.com/gadicc/yahoo-finance2 (TypeScript/Deno)

## Overview

Port yahoo-finance2 Node/Deno library to Go. Provides typed access to Yahoo Finance API endpoints with cookie/crumb authentication, response validation/coercion, and concurrency control.

## Architecture: Flat Package

Single Go package `yahoo-finance`. All types, client, and modules in one package. Matches standard Go API client pattern (go-github, google-api-go-client).

### File Structure

```
yahoo-finance/
  go.mod
  client.go          — Client struct, NewClient, Config
  crumb.go           — Cookie/crumb auth flow
  fetch.go           — Core HTTP fetch logic (acquire semaphore, build request, parse response)
  queue.go           — Semaphore-based concurrency control
  validate.go        — Input validation engine
  coerce.go          — Response coercion (date objects, TwoNumberRange strings)
  errors.go          — Custom error types
  logger.go          — Logger interface
  types.go           — Shared types (TwoNumberRange, DateInMs, ModuleOptions)

  quote.go           — Quote module
  chart.go           — Chart module
  historical.go       — Historical module (wraps Chart)
  search.go          — Search module
  quote_summary.go   — QuoteSummary module
  options.go         — Options module
  screener.go        — Screener module
  recommendations.go — RecommendationsBySymbol module
  insights.go        — Insights module
  fundamentals.go    — FundamentalsTimeSeries module
  trending.go        — TrendingSymbols module

  types_quote.go     — Quote types (Quote, QuoteEquity, QuoteETF, etc.)
  types_chart.go     — Chart types (ChartResult, ChartMeta, etc.)
  types_historical.go — Historical types
  types_search.go    — Search types
  types_quote_summary.go — QuoteSummary types (31 sub-modules)
  types_options.go   — Options types
  types_screener.go  — Screener types
  types_recommendations.go — Recommendations types
  types_insights.go  — Insights types
  types_fundamentals.go — FundamentalsTimeSeries types
  types_trending.go  — TrendingSymbols types

  yahoo_finance_test.go  — Integration tests
  crumb_test.go
  coerce_test.go
  validate_test.go
  quote_test.go
  chart_test.go
  ...
```

## Client & Configuration

```go
type Client struct {
    httpClient    *http.Client
    queryHost     string           // default: query2.finance.yahoo.com
    sem           semaphore.Weighted
    cookieJar     http.CookieJar
    crumb         string
    crumbMu       sync.Mutex
    crumbValid    bool
    logger        Logger
    validation    ValidationOpts
    fetchOptions  *FetchOptions
}

type Config struct {
    QueryHost     string           // optional, default query2.finance.yahoo.com
    Concurrency   int              // default: 4
    Logger        Logger           // optional, defaults to stdlib log
    HTTPClient    *http.Client     // optional
    Validation    ValidationOpts
    FetchOptions  *FetchOptions    // base headers, user-agent
}

type ValidationOpts struct {
    LogErrors          bool  // default: true
    AllowAdditionalProps bool  // default: true
}

type FetchOptions struct {
    Headers map[string]string
}

type Logger interface {
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
    Debug(msg string, args ...interface{})
}

func NewClient(cfg *Config) *Client
```

### Per-call ModuleOptions

```go
type ModuleOptions struct {
    ValidateResult bool
    ValidateOptions bool
    FetchOptions    *FetchOptions
}
```

## Cookie/Crumb Authentication

Yahoo Finance requires crumb tokens for: quote, options, quoteSummary, screener.

### Flow

1. `GET https://finance.yahoo.com/quote/AAPL` — collect initial cookies
2. Handle GDPR consent redirects via `guce.yahoo.com` (collectConsent → copyConsent)
3. `GET https://query1.finance.yahoo.com/v1/test/getcrumb` — fetch crumb token
4. Cache crumb + cookies; re-fetch on 401 or expiry

### Implementation

```go
func (c *Client) ensureCrumb(ctx context.Context) error
// - Thread-safe via sync.Mutex
// - Skips if crumb already valid
// - Handles full consent redirect chain
// - Returns crumb or error

func (c *Client) fetchWithCrumb(ctx context.Context, url string) (*http.Response, error)
// - Calls ensureCrumb
// - Appends crumb as query param
// - On 401: invalidates crumb, retries once
```

Key differences from JS:
- `net/http.CookieJar` replaces `tough-cookie`
- `context.Context` for cancellation/timeout
- `sync.Mutex` + bool flag replaces JS singleton promise

## Module Methods — All 10 Active Modules

| Method | Signature | Yahoo Endpoint | Crumb |
|--------|-----------|---------------|-------|
| Quote | `(ctx, symbols []string, opts *QuoteOptions) ([]Quote, error)` | `/v7/finance/quote` | Yes |
| Chart | `(ctx, symbol string, opts *ChartOptions) (*ChartResult, error)` | `/v8/finance/chart/{symbol}` | No |
| Historical | `(ctx, symbol string, opts *HistoricalOptions) ([]HistoricalRow, error)` | wraps Chart | No |
| Search | `(ctx, query string, opts *SearchOptions) (*SearchResult, error)` | `/v1/finance/search` | No |
| QuoteSummary | `(ctx, symbol string, opts *QuoteSummaryOptions) (*QuoteSummaryResult, error)` | `/v10/finance/quoteSummary/{symbol}` | Yes |
| Options | `(ctx, symbol string, opts *OptionsOptions) (*OptionsResult, error)` | `/v7/finance/options/{symbol}` | Yes |
| Screener | `(ctx, scrID string, opts *ScreenerOptions) (*ScreenerResult, error)` | `/v1/finance/screener/predefined/saved` | Yes |
| RecommendationsBySymbol | `(ctx, symbols []string, opts *RecommendationsOptions) ([]RecommendationsResult, error)` | `/v6/finance/recommendationsbysymbol/{symbols}` | No |
| Insights | `(ctx, symbol string, opts *InsightsOptions) (*InsightsResult, error)` | `/ws/insights/v2/finance/insights` | No |
| FundamentalsTimeSeries | `(ctx, symbol string, opts *FundamentalsOptions) ([]FundamentalsResult, error)` | `query1/ws/fundamentals-timeseries/v1/...` | No |
| TrendingSymbols | `(ctx, region string, opts *TrendingOptions) (*TrendingResult, error)` | `/v1/finance/trending/{region}` | No |

### Internal Module Execution Pipeline

Each module follows the same internal flow (mirrors JS `moduleExec`):

1. Validate input options (if `ValidateOptions` enabled)
2. Merge defaults with user options
3. Determine if crumb needed, ensure crumb loaded
4. Acquire semaphore slot
5. Build and execute HTTP request
6. Release semaphore slot
7. Parse Yahoo response JSON
8. Coerce response types (date objects, range strings)
9. Validate coerced result (if `ValidateResult` enabled)
10. Return typed result

## Validation & Type Coercion

### Response Coercion (Yahoo Quirks)

Yahoo returns inconsistent formats requiring coercion:

1. **Date objects:** `{"raw": 1700000000, "fmt": "2023-11-14"}` → `time.Time`
2. **TwoNumberRange strings:** `"180.68 - 589.07"` → `TwoNumberRange{Low: 180.68, High: 589.07}`
3. **Nil vs zero:** Optional numeric fields use pointer types (`*float64`, `*int64`)

Implementation: Custom `UnmarshalJSON` methods on response types.

```go
type TwoNumberRange struct {
    Low  float64
    High float64
}

// UnmarshalJSON handles both string "180.68 - 589.07" and object {low, high} formats
func (r *TwoNumberRange) UnmarshalJSON(data []byte) error { ... }

// YahooDate wraps {raw, fmt} → time.Time
type YahooDate struct { time.Time }
func (d *YahooDate) UnmarshalJSON(data []byte) error { ... }
```

### Input Validation

Struct tags + `go-playground/validator` for module options:

```go
type ChartOptions struct {
    Period1        time.Time `validate:"required"`
    Period2        time.Time
    Interval       string    `validate:"omitempty,oneof=1m 2m 5m 15m 30m 60m 90m 1h 1d 5d 1wk 1mo 3mo"`
    IncludePrePost bool
    Events         string    `validate:"omitempty,oneof=history dividents split capitalGain"`
}
```

## Concurrency & Error Handling

### Concurrency

`golang.org/x/sync/semaphore` with weighted acquire/release:

```go
func (c *Client) acquire(ctx context.Context) error
func (c *Client) release()
```

Fetch flow: acquire → build request → HTTP do → release → parse response.

### Custom Errors

```go
type BadRequestError struct { Message string }
type HTTPError struct { StatusCode int; Body string }
type InvalidOptionsError struct { Field string; Msg string }
type FailedValidationError struct { Result interface{}; Errors []string }
```

Yahoo error format parsing:
```json
{"finance": {"error": {"code": "Bad Request", "description": "..."}}}
```

Parsed into appropriate error type from HTTP response.

## Dependencies

- `golang.org/x/sync/semaphore` — concurrency control
- `github.com/go-playground/validator/v10` — input validation
- Standard library: `net/http`, `encoding/json`, `sync`, `context`, `time`, `fmt`, `log`

## Key Types (Summary)

### Quote (11 subtypes discriminated by QuoteType)

QuoteEquity, QuoteETF, QuoteCryptoCurrency, QuoteOption, QuoteFuture, QuoteIndex, QuoteMutualfund, QuoteCurrency, QuoteMoneyMarket, QuoteECNQuote, QuoteAltSymbol

Common fields: Symbol, Currency, MarketState, RegularMarketPrice, FiftyTwoWeekRange, EPS, MarketCap, etc.

### Chart

ChartResultArray (date, OHLCV, adjClose), ChartMeta (currency, symbol, exchange, valid ranges), ChartEvents (dividends, splits)

### QuoteSummary (31 sub-modules)

AssetProfile, FinancialData, DefaultKeyStatistics, Earnings, EarningsHistory, EarningsTrend, FinancialsTemplate, IncomeStatementHistory, BalanceSheetHistory, CashflowStatementHistory, etc.

### Options

OptionsResult with expirationDates, strikes, calls[], puts[]. CallOrPut: contractSymbol, strike, lastPrice, impliedVolatility, inTheMoney, Greeks.

## Not Ported (by design)

- `autoc()` — Yahoo decommissioned this endpoint
- `dailyGainers()` / `dailyLosers()` — deprecated, use Screener
- `quoteCombine()` — Go has native concurrency; batch via goroutines + channels
- Runtime detection (Deno/Bun/Cloudflare) — not relevant in Go
- Version checking — not relevant for compiled Go binary
- Notice/survey system — not applicable
