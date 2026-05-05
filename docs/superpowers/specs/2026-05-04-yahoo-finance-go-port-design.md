# yahoo-finance2 Go Port Design

Port of [gadicc/yahoo-finance2](https://github.com/gadicc/yahoo-finance2) to idiomatic Go.

## Scope

All 12 active modules. Skip deprecated `autoc`, `dailyGainers`, `dailyLosers`.

| Method | Yahoo Endpoint | Crumb Required |
|--------|---------------|----------------|
| `Chart` | `/v8/finance/chart/{symbol}` | No |
| `Quote` | `/v7/finance/quote` | No |
| `QuoteSummary` | `/v10/finance/quoteSummary/{symbol}` | Yes |
| `Search` | `/v1/finance/search` | No |
| `Historical` | wraps `Chart()` | No |
| `Options` | `/v7/finance/options/{symbol}` | Yes |
| `Insights` | `/ws/insights/v2/finance/insights` | No |
| `TrendingSymbols` | `/v1/finance/trending/{region}` | No |
| `RecommendationsBySymbol` | `/v6/finance/recommendationsbysymbol/{symbols}` | No |
| `Screener` | `/v1/finance/screener/predefined/saved` | Yes |
| `FundamentalsTimeSeries` | `query1.finance.yahoo.com/ws/fundamentals-timeseries/v1/finance/timeseries/{symbol}` | No |
| `QuoteCombine` | debounced batch → `Quote()` | No |

## Architecture: Mirror JS Structure

Struct-based client with 3-stage execution pipeline per module (validate → fetch → validate). Single Go package.

### Core Client

```go
type Client struct {
    queryHost   string
    cookieJar   http.CookieJar
    crumb       string
    crumbMu     sync.Mutex
    httpClient  *http.Client
    queue       chan struct{}     // concurrency semaphore, default 4
    validation  ValidationConfig
    logger      Logger
    userAgent   string
    fetchOpts   *FetchOptions
}

type Option func(*Client)

client := yahoofinance.NewClient(
    yahoofinance.WithQueryHost("query2.finance.yahoo.com"),
    yahoofinance.WithConcurrency(4),
    yahoofinance.WithValidation(true, true),
)
```

All public methods take `context.Context` as first argument.

### Cookie/Crumb Authentication

Yahoo requires cookies + crumb token for some endpoints.

1. First request to Yahoo → obtain session cookies
2. If redirected to consent page → parse HTML, extract CSRF token + sessionId, POST consent form, follow redirects
3. GET `/v1/test/getcrumb` → store crumb
4. Append `?crumb=<value>` on endpoints that require it
5. Cache crumb in Client, re-fetch on 401 or expired sessions
6. Concurrent crumb fetches deduplicated via `sync.Once` pattern

Crumb-required: `QuoteSummary`, `Options`, `Screener`.

### Module Execution Pipeline

Every public method follows:

1. **Validate options** — struct tag validation via `go-playground/validator`
2. **Fetch** — `moduleExec()` handles URL building, crumb, cookies, semaphore, error detection, unmarshaling
3. **Validate result** — check required fields, log warnings for missing/unexpected fields, never block returning data

```go
func (c *Client) Chart(ctx context.Context, symbol string, opts *ChartOptions) (*ChartResult, error)
```

### Error Types

| Type | When |
|------|------|
| `BadRequestError` | HTTP 400 |
| `HTTPError` | Non-OK HTTP status (carries StatusCode) |
| `InvalidOptionsError` | Input validation failure |
| `FailedValidationError` | Result validation failure (carries partial Result + Errors) |

### Validation Strategy

- **Input validation:** `go-playground/validator` struct tags. Fatal — returns `InvalidOptionsError`.
- **Result validation:** Custom validators checking required fields, type coercion (epoch → `time.Time`, string numbers → `float64`). Non-fatal — log warnings, return data anyway.
- Reason: Yahoo returns inconsistent data. Blocking on result validation would make the library unusable.

### Type Design

- Quote discriminated union → struct with `QuoteType` field + type-specific fields (not interfaces, keeps JSON unmarshaling simple)
- Dates → `time.Time` throughout (epoch/ISO conversion during unmarshaling)
- Nullable fields → pointer types (`*float64`, `*time.Time`) where Yahoo omits data

## File Structure

```
yahoo-finance/
  go.mod
  go.sum
  client.go              // Client struct, NewClient, options
  module_exec.go         // Shared fetch pipeline
  crumb.go               // Cookie/crumb auth
  errors.go              // Custom error types
  validate.go            // Input validation helpers
  validate_result.go     // Result validation helpers
  logger.go              // Logger interface

  chart.go               // Chart() + ChartOptions + ChartResult
  quote.go               // Quote() + quote type discriminated union
  quote_summary.go       // QuoteSummary() + 34 submodule types
  search.go              // Search() + SearchResult
  historical.go          // Historical() (wraps Chart)
  options.go             // Options() + OptionsResult
  insights.go            // Insights() + InsightsResult
  trending.go            // TrendingSymbols()
  recommendations.go     // RecommendationsBySymbol()
  screener.go            // Screener() + ScreenerResult
  fundamentals.go        // FundamentalsTimeSeries()
  quote_combine.go       // QuoteCombine() (debounced batch)

  types.go               // Shared types (DateInMs, TwoNumberRange, etc.)
  queue.go               // Concurrency semaphore

  yahoo_finance_test.go  // Integration tests
  chart_test.go          // Per-module tests
  ...
  testdata/              // JSON fixtures from real Yahoo API
```

## Dependencies

| Dependency | Purpose |
|-----------|---------|
| `go-playground/validator` | Input struct tag validation |
| Standard `net/http/cookiejar` | Cookie management |

No other external dependencies. Queue, CSV parser, deep merge, schema validator all implemented in-package.

## Testing Strategy

- **Unit tests:** `httptest.Server` mocks with fixtures from `testdata/`. Tests input validation, URL building, result unmarshaling per module.
- **Integration tests:** Behind `//go:build integration` tag. Hit real Yahoo API. Verify end-to-end including cookie/crumb.
- **Golden files:** JSON response snapshots in `testdata/`. Catch Yahoo API changes via comparison.
- No external dependencies for unit tests.

## Implementation Order

1. Core infrastructure: client, errors, logger, queue, module_exec
2. Cookie/crumb auth
3. Chart module (most complex, other modules depend on patterns)
4. Quote module (discriminated union pattern)
5. Search, Historical (simpler modules)
6. QuoteSummary (largest type surface — 34 submodules)
7. Options, Insights, TrendingSymbols, RecommendationsBySymbol, Screener, FundamentalsTimeSeries
8. QuoteCombine (debouncing pattern)
9. Result validation across all modules
10. Integration tests + golden files
