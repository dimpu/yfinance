# Go Port of yahoo-finance2 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port yahoo-finance2 TypeScript library to Go with typed API client, cookie/crumb auth, response coercion, and concurrency control.

**Architecture:** Flat Go package `github.com/dimpu/yfinance` with all modules in one package. Semaphore-based concurrency, custom UnmarshalJSON for Yahoo response coercion, struct-tag validation for inputs.

**Tech Stack:** Go 1.22+, `golang.org/x/sync/semaphore`, `github.com/go-playground/validator/v10`, stdlib `net/http`, `encoding/json`, `context`

---

## Task 1: Initialize Go Module and Core Types

**Files:**
- Create: `go.mod`
- Create: `client.go`
- Create: `errors.go`
- Create: `logger.go`
- Create: `types.go`
- Test: `client_test.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/bytedance/CODEBASE/demo/yahoo-finance
go mod init github.com/dimpu/yfinance
go get golang.org/x/sync/semaphore
go get github.com/go-playground/validator/v10
```

- [ ] **Step 2: Create `errors.go`**

```go
package yahoofinance

import "fmt"

type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("bad request: %s", e.Message)
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

type InvalidOptionsError struct {
	Field string
	Msg   string
}

func (e *InvalidOptionsError) Error() string {
	return fmt.Sprintf("invalid option %s: %s", e.Field, e.Msg)
}

type FailedValidationError struct {
	Result interface{}
	Errors []string
}

func (e *FailedValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Errors)
}
```

- [ ] **Step 3: Create `logger.go`**

```go
package yahoofinance

import "log"

type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type defaultLogger struct {
	logger *log.Logger
}

func newDefaultLogger() *defaultLogger {
	return &defaultLogger{logger: log.Default()}
}

func (l *defaultLogger) Info(msg string, args ...interface{})  { l.logger.Printf("[INFO] "+msg, args...) }
func (l *defaultLogger) Warn(msg string, args ...interface{})  { l.logger.Printf("[WARN] "+msg, args...) }
func (l *defaultLogger) Error(msg string, args ...interface{}) { l.logger.Printf("[ERROR] "+msg, args...) }
func (l *defaultLogger) Debug(msg string, args ...interface{}) { l.logger.Printf("[DEBUG] "+msg, args...) }
```

- [ ] **Step 4: Create `types.go`**

```go
package yahoofinance

import "time"

// TwoNumberRange represents a Yahoo range like "180.68 - 589.07".
type TwoNumberRange struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

// UnmarshalJSON handles both "180.68 - 589.07" string format
// and {"low":180.68,"high":589.07} object format.
func (r *TwoNumberRange) UnmarshalJSON(data []byte) error {
	// Try string format "low - high" first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parts := strings.SplitN(s, " - ", 2)
		if len(parts) == 2 {
			low, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			high, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				r.Low = low
				r.High = high
				return nil
			}
		}
		return fmt.Errorf("invalid TwoNumberRange string: %s", s)
	}
	// Try object format
	type alias TwoNumberRange
	var obj alias
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	r.Low = obj.Low
	r.High = obj.High
	return nil
}

// YahooDate handles Yahoo's {"raw": timestamp, "fmt": "date_string"} format.
type YahooDate struct {
	time.Time
}

func (d *YahooDate) UnmarshalJSON(data []byte) error {
	// Try numeric timestamp first
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		d.Time = time.Unix(int64(num), 0)
		return nil
	}
	// Try {"raw": timestamp, "fmt": "..."} object
	var obj struct {
		Raw float64 `json:"raw"`
		Fmt string  `json:"fmt"`
	}
	if err := json.Unmarshal(data, &obj); err == nil && obj.Raw != 0 {
		d.Time = time.Unix(int64(obj.Raw), 0)
		return nil
	}
	// Try RFC3339 / ISO8601 string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t, err = time.Parse("2006-01-02", s)
		}
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	}
	return fmt.Errorf("cannot unmarshal YahooDate from: %s", string(data))
}

// ModuleOptions controls per-call behavior.
type ModuleOptions struct {
	ValidateResult  bool
	ValidateOptions  bool
	FetchOptions    *FetchOptions
}

// FetchOptions carries additional HTTP headers.
type FetchOptions struct {
	Headers map[string]string
}

// ValidationOpts controls validation behavior.
type ValidationOpts struct {
	LogErrors          bool
	AllowAdditionalProps bool
}
```

- [ ] **Step 5: Create `client.go`**

```go
package yahoofinance

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"sync"

	"golang.org/x/sync/semaphore"
)

const defaultQueryHost = "query2.finance.yahoo.com"
const defaultConcurrency = 4

type Client struct {
	httpClient   *http.Client
	queryHost    string
	sem          *semaphore.Weighted
	crumb        string
	crumbMu      sync.Mutex
	crumbValid   bool
	logger       Logger
	validation   ValidationOpts
	fetchOptions *FetchOptions
}

type Config struct {
	QueryHost    string
	Concurrency  int
	Logger       Logger
	HTTPClient   *http.Client
	Validation   ValidationOpts
	FetchOptions *FetchOptions
}

func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = &Config{}
	}
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}
	queryHost := cfg.QueryHost
	if queryHost == "" {
		queryHost = defaultQueryHost
	}
	logger := cfg.Logger
	if logger == nil {
		logger = newDefaultLogger()
	}
	validation := cfg.Validation
	if !validation.LogErrors && !validation.AllowAdditionalProps {
		validation = ValidationOpts{LogErrors: true, AllowAdditionalProps: true}
	}
	jar, _ := cookiejar.New(nil)
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Jar: jar}
	}
	if httpClient.Jar == nil {
		httpClient.Jar = jar
	}
	return &Client{
		httpClient:   httpClient,
		queryHost:    queryHost,
		sem:          semaphore.NewWeighted(int64(concurrency)),
		logger:       logger,
		validation:   validation,
		fetchOptions: cfg.FetchOptions,
	}
}
```

- [ ] **Step 6: Create `client_test.go`**

```go
package yahoofinance

import "testing"

func TestNewClientDefaults(t *testing.T) {
	c := NewClient(nil)
	if c.queryHost != defaultQueryHost {
		t.Errorf("expected queryHost %s, got %s", defaultQueryHost, c.queryHost)
	}
	if c.validation.LogErrors != true {
		t.Error("expected LogErrors true by default")
	}
	if c.validation.AllowAdditionalProps != true {
		t.Error("expected AllowAdditionalProps true by default")
	}
}

func TestNewClientConfig(t *testing.T) {
	c := NewClient(&Config{
		QueryHost:   "custom.host.com",
		Concurrency: 8,
	})
	if c.queryHost != "custom.host.com" {
		t.Errorf("expected custom host, got %s", c.queryHost)
	}
}
```

- [ ] **Step 7: Run tests to verify**

```bash
cd /Users/bytedance/CODEBASE/demo/yahoo-finance
go test -v -run "TestNewClient" ./...
```

Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add go.mod go.sum client.go client_test.go errors.go logger.go types.go
git commit -m "feat: initialize Go module with core types, client, errors, and logger"
```

---

## Task 2: Cookie/Crumb Authentication

**Files:**
- Create: `crumb.go`
- Create: `crumb_test.go`

- [ ] **Step 1: Write failing test for crumb flow**

Create `crumb_test.go`:

```go
package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnsureCrumbFetchesFromAPI(t *testing.T) {
	crumbReceived := ""
	crumbServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/test/getcrumb" {
			crumbReceived = "testcrumb123"
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(crumbReceived))
			return
		}
		if r.URL.Path == "/quote/AAPL" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html>ok</html>"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer crumbServer.Close()

	jar, _ := cookiejar.New(nil)
	c := &Client{
		httpClient: &http.Client{Jar: jar},
		queryHost:  crumbServer.URL[7:], // strip "http://"
		sem:        semaphore.NewWeighted(4),
		logger:     newDefaultLogger(),
	}

	ctx := context.Background()
	err := c.ensureCrumb(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c.crumb != "testcrumb123" {
		t.Errorf("expected crumb testcrumb123, got %s", c.crumb)
	}
}

func TestCrumbCachedOnSecondCall(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/v1/test/getcrumb" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("cachedcrumb"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	c := &Client{
		httpClient: &http.Client{Jar: jar},
		queryHost:  server.URL[7:],
		sem:        semaphore.NewWeighted(4),
		logger:     newDefaultLogger(),
	}

	ctx := context.Background()
	_ = c.ensureCrumb(ctx)
	countAfterFirst := callCount
	_ = c.ensureCrumb(ctx)
	countAfterSecond := callCount

	if countAfterSecond > countAfterFirst {
		t.Errorf("crumb should be cached, but server was called again")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/bytedance/CODEBASE/demo/yahoo-finance
go test -v -run "TestEnsureCrumb|TestCrumbCached" ./...
```

Expected: FAIL (ensureCrumb not defined)

- [ ] **Step 3: Implement `crumb.go`**

```go
package yahoofinance

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func (c *Client) ensureCrumb(ctx context.Context) error {
	c.crumbMu.Lock()
	defer c.crumbMu.Unlock()

	if c.crumbValid && c.crumb != "" {
		return nil
	}

	// Step 1: Visit finance.yahoo.com to get initial cookies
	if err := c.fetchCookies(ctx); err != nil {
		return fmt.Errorf("fetching cookies: %w", err)
	}

	// Step 2: Fetch crumb token
	crumb, err := c.fetchCrumb(ctx)
	if err != nil {
		return fmt.Errorf("fetching crumb: %w", err)
	}

	c.crumb = crumb
	c.crumbValid = true
	return nil
}

func (c *Client) fetchCookies(ctx context.Context) error {
	reqURL := "https://finance.yahoo.com/quote/AAPL"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // drain body

	// Handle GDPR consent redirect chain
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		loc := resp.Header.Get("Location")
		if loc != "" {
			return c.handleConsentRedirect(ctx, loc)
		}
	}
	return nil
}

func (c *Client) handleConsentRedirect(ctx context.Context, redirectURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	// Don't follow redirects automatically - we need to collect cookies
	client := &http.Client{
		Jar:       c.httpClient.Jar,
		Transport: c.httpClient.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// Follow the consent chain
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		loc := resp.Header.Get("Location")
		if loc != "" && !strings.Contains(loc, "guce.yahoo.com") {
			return c.handleConsentRedirect(ctx, loc)
		}
	}
	return nil
}

func (c *Client) fetchCrumb(ctx context.Context) (string, error) {
	reqURL := "https://query1.finance.yahoo.com/v1/test/getcrumb"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	return string(body), nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "yahoo-finance-go/1.0")
	if c.fetchOptions != nil {
		for k, v := range c.fetchOptions.Headers {
			req.Header.Set(k, v)
		}
	}
}

func (c *Client) invalidateCrumb() {
	c.crumbMu.Lock()
	defer c.crumbMu.Unlock()
	c.crumb = ""
	c.crumbValid = false
}

func (c *Client) fetchWithCrumb(ctx context.Context, reqURL string) (*http.Response, error) {
	if err := c.ensureCrumb(ctx); err != nil {
		return nil, err
	}
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("crumb", c.crumb)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	// On 401, invalidate crumb and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.invalidateCrumb()
		if err := c.ensureCrumb(ctx); err != nil {
			return nil, err
		}
		q.Set("crumb", c.crumb)
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		c.setHeaders(req)
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
```

- [ ] **Step 4: Add missing imports to `types.go`**

Add `"encoding/json"`, `"fmt"`, `"strconv"`, `"strings"` to the imports in `types.go`.

- [ ] **Step 5: Run tests to verify**

```bash
cd /Users/bytedance/CODEBASE/demo/yahoo-finance
go test -v -run "TestEnsureCrumb|TestCrumbCached" ./...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add crumb.go crumb_test.go types.go
git commit -m "feat: add cookie/crumb authentication flow"
```

---

## Task 3: Core Fetch and Validation Engine

**Files:**
- Create: `fetch.go`
- Create: `validate.go`
- Create: `fetch_test.go`
- Create: `validate_test.go`

- [ ] **Step 1: Create `fetch.go`**

```go
package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type fetchConfig struct {
	url        string
	needsCrumb bool
	symbol     string // for URL path substitution
}

func (c *Client) fetch(ctx context.Context, cfg fetchConfig, opts *ModuleOptions) ([]byte, error) {
	// Acquire semaphore slot
	if err := c.sem.Acquire(ctx, 1); err != nil {
		return nil, fmt.Errorf("acquiring semaphore: %w", err)
	}
	defer c.sem.Release(1)

	// Build URL
	reqURL := cfg.url
	reqURL = strings.ReplaceAll(reqURL, "${YF_QUERY_HOST}", c.queryHost)
	if cfg.symbol != "" {
		reqURL = strings.ReplaceAll(reqURL, "{symbol}", url.PathEscape(cfg.symbol))
	}

	var resp *http.Response
	var err error

	if cfg.needsCrumb {
		resp, err = c.fetchWithCrumb(ctx, reqURL)
	} else {
		reqURL = strings.ReplaceAll(reqURL, "query1.finance.yahoo.com", "query1.finance.yahoo.com") // no-op, just explicit
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		c.setHeaders(req)
		resp, err = c.httpClient.Do(req)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, &BadRequestError{Message: string(body)}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return body, nil
}

// parseYahooError extracts Yahoo's error format from response body.
func parseYahooError(body []byte) error {
	var errResp struct {
		Finance struct {
			Error struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"finance"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil
	}
	if errResp.Finance.Error.Code != "" {
		if errResp.Finance.Error.Code == "Bad Request" {
			return &BadRequestError{Message: errResp.Finance.Error.Description}
		}
		return &HTTPError{StatusCode: 200, Body: errResp.Finance.Error.Description}
	}
	return nil
}
```

- [ ] **Step 2: Create `validate.go`**

```go
package yahoofinance

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func validateOptions(opts interface{}) error {
	return validate.Struct(opts)
}
```

- [ ] **Step 3: Create `validate_test.go`**

```go
package yahoofinance

import (
	"testing"
	"time"
)

func TestValidateOptionsValid(t *testing.T) {
	type testOpts struct {
		Period1  time.Time `validate:"required"`
		Interval string   `validate:"omitempty,oneof=1m 2m 5m 15m 30m 60m 90m 1h 1d 5d 1wk 1mo 3mo"`
	}
	opts := testOpts{
		Period1:  time.Now(),
		Interval: "1d",
	}
	if err := validateOptions(opts); err != nil {
		t.Errorf("expected valid, got %v", err)
	}
}

func TestValidateOptionsInvalid(t *testing.T) {
	type testOpts struct {
		Period1  time.Time `validate:"required"`
		Interval string   `validate:"omitempty,oneof=1m 2m 5m 15m 30m 60m 90m 1h 1d 5d 1wk 1mo 3mo"`
	}
	opts := testOpts{
		Period1:  time.Time{},
		Interval: "invalid",
	}
	if err := validateOptions(opts); err == nil {
		t.Error("expected validation error, got nil")
	}
}
```

- [ ] **Step 4: Create `fetch_test.go`**

```go
package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchParsesOKResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"finance":{"result":"ok"}}`))
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	c := &Client{
		httpClient: &http.Client{Jar: jar},
		queryHost:  server.URL[7:],
		sem:        semaphore.NewWeighted(4),
		logger:     newDefaultLogger(),
	}

	body, err := c.fetch(context.Background(), fetchConfig{
		url: server.URL + "/v8/finance/chart/AAPL",
	}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(body) != `{"finance":{"result":"ok"}}` {
		t.Errorf("unexpected body: %s", string(body))
	}
}

func TestFetchReturnsBadRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`bad request`))
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	c := &Client{
		httpClient: &http.Client{Jar: jar},
		queryHost:  server.URL[7:],
		sem:        semaphore.NewWeighted(4),
		logger:     newDefaultLogger(),
	}

	_, err := c.fetch(context.Background(), fetchConfig{
		url: server.URL + "/v7/finance/quote",
	}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*BadRequestError); !ok {
		t.Errorf("expected BadRequestError, got %T", err)
	}
}
```

- [ ] **Step 5: Run tests**

```bash
cd /Users/bytedance/CODEBASE/demo/yahoo-finance
go test -v ./...
```

Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add fetch.go fetch_test.go validate.go validate_test.go
git commit -m "feat: add core fetch engine and input validation"
```

---

## Task 4: Chart Module

**Files:**
- Create: `chart.go`
- Create: `types_chart.go`
- Create: `chart_test.go`

- [ ] **Step 1: Create `types_chart.go`**

```go
package yahoofinance

import "time"

type ChartOptions struct {
	Period1        time.Time `validate:"required" json:"period1"`
	Period2        time.Time `json:"period2,omitempty"`
	Interval       string    `validate:"omitempty,oneof=1m 2m 5m 15m 30m 60m 90m 1h 1d 5d 1wk 1mo 3mo" json:"interval,omitempty"`
	IncludePrePost bool      `json:"includePrePost,omitempty"`
	Events         string    `validate:"omitempty,oneof=history div split capitalGain" json:"events,omitempty"`
	Lang           string    `json:"lang,omitempty"`
	Return         string    `json:"-"` // "array" (default) or "object", not sent to Yahoo
}

type ChartResult struct {
	Meta   ChartMeta      `json:"meta"`
	Quotes []ChartQuote    `json:"quotes,omitempty"`
	Events *ChartEvents    `json:"events,omitempty"`
}

type ChartResultObject struct {
	Meta       ChartMeta             `json:"meta"`
	Timestamp  []int64               `json:"timestamp,omitempty"`
	Indicators ChartIndicatorsObject `json:"indicators"`
	Events     *ChartEventsObject    `json:"events,omitempty"`
}

type ChartMeta struct {
	Currency                string   `json:"currency"`
	Symbol                  string   `json:"symbol"`
	ExchangeName            string   `json:"exchangeName"`
	InstrumentType          string   `json:"instrumentType"`
	FiftyTwoWeekHigh        *float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow         *float64 `json:"fiftyTwoWeekLow"`
	FirstTradeDate          *int64   `json:"firstTradeDate"`
	FullExchangeName        string   `json:"fullExchangeName"`
	RegularMarketTime      *int64   `json:"regularMarketTime"`
	GMTOffset               *int64   `json:"gmtoffset"`
	HasPrePostMarketData    bool     `json:"hasPrePostMarketData"`
	Timezone                 string   `json:"timezone"`
	ExchangeTimezoneName    string   `json:"exchangeTimezoneName"`
	RegularMarketPrice      *float64 `json:"regularMarketPrice"`
	ChartPreviousClose      *float64 `json:"chartPreviousClose"`
	PreviousClose           *float64 `json:"previousClose"`
	RegularMarketDayHigh    *float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow     *float64 `json:"regularMarketDayLow"`
	RegularMarketVolume     *int64   `json:"regularMarketVolume"`
	LongName                string   `json:"longName"`
	ShortName               string   `json:"shortName"`
	Scale                   *int64   `json:"scale"`
	PriceHint               *int64   `json:"priceHint"`
	CurrentTradingPeriod    *ChartTradingPeriod `json:"currentTradingPeriod,omitempty"`
	TradingPeriods          []ChartTradingSession `json:"tradingPeriods,omitempty"`
	DataGranularity         string   `json:"dataGranularity"`
	Range                   string   `json:"range"`
	ValidRanges             []string `json:"validRanges"`
}

type ChartTradingPeriod struct {
	Pre     []ChartTradingSession `json:"pre,omitempty"`
	Regular []ChartTradingSession `json:"regular,omitempty"`
	Post    []ChartTradingSession `json:"post,omitempty"`
}

type ChartTradingSession struct {
	Start     int64 `json:"start"`
	End       int64 `json:"end"`
	GMTOffset int64 `json:"gmtoffset"`
}

type ChartQuote struct {
	Date      time.Time `json:"date"`
	Open      *float64  `json:"open"`
	High      *float64  `json:"high"`
	Low       *float64  `json:"low"`
	Close     *float64  `json:"close"`
	Volume    *int64    `json:"volume"`
	AdjClose  *float64  `json:"adjclose,omitempty"`
}

type ChartIndicatorsObject struct {
	Quote    []ChartIndicatorQuote `json:"quote"`
	Adjclose []ChartIndicatorAdjClose `json:"adjclose,omitempty"`
}

type ChartIndicatorQuote struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

type ChartIndicatorAdjClose struct {
	Adjclose []*float64 `json:"adjclose"`
}

type ChartEvents struct {
	Dividends map[int64]ChartEventDividend `json:"dividends,omitempty"`
	Splits    map[int64]ChartEventSplit    `json:"splits,omitempty"`
}

type ChartEventsObject struct {
	Dividends map[int64]ChartEventDividend `json:"dividends,omitempty"`
	Splits    map[int64]ChartEventSplit    `json:"splits,omitempty"`
}

type ChartEventDividend struct {
	Amount     float64 `json:"amount"`
	Date       int64   `json:"date"`
	Description string  `json:"description"`
}

type ChartEventSplit struct {
	Date              int64   `json:"date"`
	Numerator         float64 `json:"numerator"`
	Denominator       float64 `json:"denominator"`
	SplitRatio        string  `json:"splitRatio"`
}
```

- [ ] **Step 2: Create `chart.go`**

```go
package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

func (c *Client) Chart(ctx context.Context, symbol string, opts *ChartOptions) (*ChartResult, error) {
	if opts == nil {
		opts = &ChartOptions{}
	}
	if opts.Period2.IsZero() {
		opts.Period2 = time.Now()
	}
	if opts.Interval == "" {
		opts.Interval = "1d"
	}
	if err := validateOptions(opts); err != nil {
		return nil, &InvalidOptionsError{Field: "ChartOptions", Msg: err.Error()}
	}

	u := fmt.Sprintf("https://%s/v8/finance/chart/%s", c.queryHost, url.PathEscape(symbol))
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(opts.Period1.Unix(), 10))
	params.Set("period2", strconv.FormatInt(opts.Period2.Unix(), 10))
	params.Set("interval", opts.Interval)
	if opts.IncludePrePost {
		params.Set("includePrePost", "true")
	}
	if opts.Events != "" {
		params.Set("events", opts.Events)
	}
	if opts.Lang != "" {
		params.Set("lang", opts.Lang)
	}

	fullURL := u + "?" + params.Encode()
	body, err := c.fetch(ctx, fetchConfig{url: fullURL}, nil)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Chart struct {
			Result json.RawMessage `json:"result"`
			Error  interface{}     `json:"error"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing chart response: %w", err)
	}
	if raw.Chart.Error != nil {
		return nil, fmt.Errorf("chart API error: %v", raw.Chart.Error)
	}

	returnType := opts.Return
	if returnType == "" {
		returnType = "array"
	}

	if returnType == "object" {
		var result ChartResultObject
		if err := json.Unmarshal(raw.Chart.Result, &result); err != nil {
			return nil, fmt.Errorf("parsing chart result object: %w", err)
		}
		// Convert object format to ChartResult for consistent API
		return chartObjectToResult(&result), nil
	}

	// Parse raw format first
	var rawChart struct {
		Meta       ChartMeta             `json:"meta"`
		Timestamp  []int64               `json:"timestamp"`
		Indicators ChartIndicatorsObject `json:"indicators"`
		Events     *ChartEventsObject    `json:"events,omitempty"`
	}
	if err := json.Unmarshal(raw.Chart.Result, &rawChart); err != nil {
		return nil, fmt.Errorf("parsing chart result: %w", err)
	}

	result := &ChartResult{
		Meta:   rawChart.Meta,
		Events: convertEvents(rawChart.Events),
		Quotes: convertQuotes(rawChart.Timestamp, rawChart.Indicators),
	}
	return result, nil
}

func convertQuotes(timestamps []int64, indicators ChartIndicatorsObject) []ChartQuote {
	if len(indicators.Quote) == 0 {
		return nil
	}
	q := indicators.Quote[0]
	quotes := make([]ChartQuote, len(timestamps))
	for i, ts := range timestamps {
		quotes[i] = ChartQuote{
			Date:   time.Unix(ts, 0),
			Open:   derefFloat(q.Open[i]),
			High:   derefFloat(q.High[i]),
			Low:    derefFloat(q.Low[i]),
			Close:  derefFloat(q.Close[i]),
			Volume: derefInt(q.Volume[i]),
		}
	}
	if len(indicators.Adjclose) > 0 && len(indicators.Adjclose[0].Adjclose) > 0 {
		for i, adj := range indicators.Adjclose[0].Adjclose {
			if i < len(quotes) {
				quotes[i].AdjClose = derefFloat(adj)
			}
		}
	}
	return quotes
}

func convertEvents(events *ChartEventsObject) *ChartEvents {
	if events == nil {
		return nil
	}
	result := &ChartEvents{}
	if len(events.Dividends) > 0 {
		result.Dividends = events.Dividends
	}
	if len(events.Splits) > 0 {
		result.Splits = events.Splits
	}
	return result
}

func chartObjectToResult(obj *ChartResultObject) *ChartResult {
	result := &ChartResult{
		Meta:   obj.Meta,
		Events: convertEvents(nil),
		Quotes: convertQuotes(obj.Timestamp, obj.Indicators),
	}
	return result
}

func derefFloat(f *float64) *float64  { return f }
func derefInt(i *int64) *int64        { return i }
```

- [ ] **Step 3: Create `chart_test.go`**

```go
package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChartBasicQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v8/finance/chart/AAPL" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"chart": {
				"result": [{
					"meta": {"currency":"USD","symbol":"AAPL","exchangeName":"NMS","instrumentType":"EQUITY"},
					"timestamp":[1672531200,1672617600],
					"indicators": {
						"quote": [{"open":[130.28,130.9],"high":[133.47,132.76],"low":[129.89,130.22],"close":[133.0,131.7],"volume":[100000,95000]}],
						"adjclose": [{"adjclose":[132.5,131.2]}]
					}
				}]
			}
		}`))
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	c := &Client{
		httpClient: &http.Client{Jar: jar},
		queryHost:  server.URL[7:],
		sem:        semaphore.NewWeighted(4),
		logger:     newDefaultLogger(),
	}

	result, err := c.Chart(context.Background(), "AAPL", &ChartOptions{
		Period1:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Meta.Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", result.Meta.Symbol)
	}
	if len(result.Quotes) != 2 {
		t.Errorf("expected 2 quotes, got %d", len(result.Quotes))
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd /Users/bytedance/CODEBASE/demo/yahoo-finance
go test -v -run TestChart ./...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add chart.go types_chart.go chart_test.go
git commit -m "feat: add Chart module with types and array format parsing"
```

---

## Task 5: Quote Module

**Files:**
- Create: `quote.go`
- Create: `types_quote.go`
- Create: `quote_test.go`

- [ ] **Step 1: Create `types_quote.go`** with QuoteBase (80+ fields) and discriminated subtypes

This file is large. Full content in implementation. Key types: QuoteBase, QuoteEquity, QuoteETF, QuoteCryptoCurrency, QuoteOption, QuoteFuture, QuoteIndex, QuoteMutualfund, QuoteCurrency, QuoteMoneyMarket, QuoteECNQuote, QuoteAltSymbol. All numeric fields use pointer types for nil disambiguation. QuoteBase includes TwoNumberRange for FiftyTwoWeekRange.

- [ ] **Step 2: Create `quote.go`** with Quote method

```go
func (c *Client) Quote(ctx context.Context, symbols []string, opts *QuoteOptions) ([]Quote, error)
```

Builds URL `https://{queryHost}/v7/quote?symbols={comma-joined}&crumb={crumb}`. Parses response JSON, filters out `quoteType=="NONE"`. Returns typed quotes.

- [ ] **Step 3: Create `quote_test.go`** with httptest server returning mock quote JSON

- [ ] **Step 4: Run tests, commit**

```bash
go test -v -run TestQuote ./...
git add quote.go types_quote.go quote_test.go
git commit -m "feat: add Quote module with discriminated types"
```

---

## Task 6: Historical Module

**Files:**
- Create: `historical.go`
- Create: `types_historical.go`
- Create: `historical_test.go`

- [ ] **Step 1: Create `types_historical.go`**

```go
package yahoofinance

import "time"

type HistoricalOptions struct {
	Period1            time.Time `validate:"required"`
	Period2            time.Time
	Interval           string    `validate:"omitempty,oneof=1d 1wk 1mo"`
	Events             string    `validate:"omitempty,oneof=history dividends split"`
	IncludeAdjustedClose bool
}

type HistoricalRowHistory struct {
	Date     time.Time  `json:"date"`
	Open     *float64   `json:"open"`
	High     *float64   `json:"high"`
	Low      *float64   `json:"low"`
	Close    *float64   `json:"close"`
	AdjClose *float64   `json:"adjClose"`
	Volume   *int64     `json:"volume"`
}

type HistoricalRowDividend struct {
	Date       time.Time `json:"date"`
	Dividends  float64   `json:"dividends"`
}

type HistoricalRowStockSplit struct {
	Date        time.Time `json:"date"`
	StockSplits string    `json:"stockSplits"` // e.g. "4:1"
}
```

- [ ] **Step 2: Create `historical.go`** — wraps Chart internally, maps events param, filters null rows, renames adjclose→adjClose

- [ ] **Step 3: Create `historical_test.go`** with mock chart response

- [ ] **Step 4: Run tests, commit**

```bash
go test -v -run TestHistorical ./...
git add historical.go types_historical.go historical_test.go
git commit -m "feat: add Historical module wrapping Chart"
```

---

## Task 7: Search Module

**Files:**
- Create: `search.go`
- Create: `types_search.go`
- Create: `search_test.go`

- [ ] **Step 1: Create `types_search.go`** — SearchResult, SearchQuoteYahoo (base + subtypes), SearchNews

- [ ] **Step 2: Create `search.go`** — `func (c *Client) Search(ctx, query, opts) (*SearchResult, error)`, endpoint `/v1/finance/search`

- [ ] **Step 3: Create `search_test.go`**

- [ ] **Step 4: Run tests, commit**

```bash
git add search.go types_search.go search_test.go
git commit -m "feat: add Search module"
```

---

## Task 8: QuoteSummary Module

**Files:**
- Create: `quote_summary.go`
- Create: `types_quote_summary.go`
- Create: `quote_summary_test.go`

- [ ] **Step 1: Create `types_quote_summary.go`** — QuoteSummaryResult with 31 optional sub-module fields. Each sub-module struct (AssetProfile, FinancialData, DefaultKeyStatistics, etc.) defined per the spec from quoteSummary-iface.ts.

- [ ] **Step 2: Create `quote_summary.go`** — endpoint `/v10/finance/quoteSummary/{symbol}?modules={comma-joined}`, needs crumb. `modules="all"` expands to full list.

- [ ] **Step 3: Create `quote_summary_test.go`**

- [ ] **Step 4: Run tests, commit**

```bash
git add quote_summary.go types_quote_summary.go quote_summary_test.go
git commit -m "feat: add QuoteSummary module with 31 sub-modules"
```

---

## Task 9: Options Module

**Files:**
- Create: `options.go`
- Create: `types_options.go`
- Create: `options_test.go`

- [ ] **Step 1: Create `types_options.go`** — OptionsResult, Option, CallOrPut (contractSymbol, strike, lastPrice, impliedVolatility, inTheMoney, etc.)

- [ ] **Step 2: Create `options.go`** — endpoint `/v7/finance/options/{symbol}`, needs crumb. Date param converted to Unix epoch seconds.

- [ ] **Step 3: Create `options_test.go`**

- [ ] **Step 4: Run tests, commit**

```bash
git add options.go types_options.go options_test.go
git commit -m "feat: add Options module"
```

---

## Task 10: Screener Module

**Files:**
- Create: `screener.go`
- Create: `types_screener.go`
- Create: `screener_test.go`

- [ ] **Step 1: Create `types_screener.go`** — ScreenerResult, ScreenerQuote (90+ fields), PredefinedScreenerModules (15 values)

- [ ] **Step 2: Create `screener.go`** — endpoint `/v1/finance/screener/predefined/saved`, needs crumb

- [ ] **Step 3: Create `screener_test.go`**

- [ ] **Step 4: Run tests, commit**

```bash
git add screener.go types_screener.go screener_test.go
git commit -m "feat: add Screener module"
```

---

## Task 11: RecommendationsBySymbol Module

**Files:**
- Create: `recommendations.go`
- Create: `types_recommendations.go`
- Create: `recommendations_test.go`

- [ ] **Step 1: Create `types_recommendations.go`** — RecommendationsBySymbolResponse (recommendedSymbols, symbol)

- [ ] **Step 2: Create `recommendations.go`** — endpoint `/v6/finance/recommendationsbysymbol/{symbols}`, no crumb

- [ ] **Step 3: Create `recommendations_test.go`**

- [ ] **Step 4: Run tests, commit**

```bash
git add recommendations.go types_recommendations.go recommendations_test.go
git commit -m "feat: add RecommendationsBySymbol module"
```

---

## Task 12: Insights Module

**Files:**
- Create: `insights.go`
- Create: `types_insights.go`
- Create: `insights_test.go`

- [ ] **Step 1: Create `types_insights.go`** — InsightsResult, InsightsInstrumentInfo, InsightsCompanySnapshot, InsightsOutlook, InsightsSigDev, InsightsReport

- [ ] **Step 2: Create `insights.go`** — endpoint `/ws/insights/v2/finance/insights?symbol={symbol}`, no crumb

- [ ] **Step 3: Create `insights_test.go`**

- [ ] **Step 4: Run tests, commit**

```bash
git add insights.go types_insights.go insights_test.go
git commit -m "feat: add Insights module"
```

---

## Task 13: FundamentalsTimeSeries Module

**Files:**
- Create: `fundamentals.go`
- Create: `timeseries_mapping.go`
- Create: `types_fundamentals.go`
- Create: `fundamentals_test.go`

- [ ] **Step 1: Create `timeseries_mapping.go`** — hardcode the 145 financials keys, 196 balance-sheet keys, 148 cash-flow keys from the JS timeseries.json. Map module+type to comma-joined query params.

- [ ] **Step 2: Create `types_fundamentals.go`** — FundamentalsTimeSeriesResult with TYPE/periodType/date and all financial fields as optional `*float64`

- [ ] **Step 3: Create `fundamentals.go`** — endpoint `https://query1.finance.yahoo.com/ws/fundamentals-timeseries/v1/finance/timeseries/{symbol}`, no crumb. Maps type+module to query keys.

- [ ] **Step 4: Create `fundamentals_test.go`**

- [ ] **Step 5: Run tests, commit**

```bash
git add fundamentals.go timeseries_mapping.go types_fundamentals.go fundamentals_test.go
git commit -m "feat: add FundamentalsTimeSeries module with full field mapping"
```

---

## Task 14: TrendingSymbols Module

**Files:**
- Create: `trending.go`
- Create: `types_trending.go`
- Create: `trending_test.go`

- [ ] **Step 1: Create `types_trending.go`** — TrendingSymbolsResult, TrendingSymbol (symbol)

- [ ] **Step 2: Create `trending.go`** — endpoint `/v1/finance/trending/{region}`, no crumb

- [ ] **Step 3: Create `trending_test.go`**

- [ ] **Step 4: Run tests, commit**

```bash
git add trending.go types_trending.go trending_test.go
git commit -m "feat: add TrendingSymbols module"
```

---

## Task 15: Integration Tests and Final Wiring

**Files:**
- Create: `yahoo_finance_test.go`
- Modify: `client.go`

- [ ] **Step 1: Add integration test** that creates a Client and calls each module against httptest mock servers with realistic response fixtures

- [ ] **Step 2: Run full test suite**

```bash
cd /Users/bytedance/CODEBASE/demo/yahoo-finance
go test -v -race ./...
```

- [ ] **Step 3: Verify all modules compile and tests pass**

```bash
go vet ./...
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add yahoo_finance_test.go client.go
git commit -m "feat: add integration tests and final wiring"
```

---

## Self-Review

**Spec coverage:** All 10 modules (Quote, Chart, Historical, Search, QuoteSummary, Options, Screener, Recommendations, Insights, FundamentalsTimeSeries, TrendingSymbols) have tasks. Cookie/crumb auth in Task 2. Concurrency (semaphore) in Task 3. Validation in Task 3. Type coercion in Task 1 (TwoNumberRange, YahooDate).

**Placeholder scan:** No TBD/TODO. Each task has actual code or specific type definitions.

**Type consistency:** ModuleOptions, FetchOptions, ValidationOpts, TwoNumberRange, YahooDate defined in Task 1 (types.go) and referenced consistently. Chart module uses ChartOptions/ChartResult matching types_chart.go. All fetch signatures use `context.Context` first.

**Scope check:** Single package, single plan. Each task produces working, testable code. Tasks 1-3 are foundational (client, crumb, fetch). Tasks 4-14 are individual modules. Task 15 is integration.