package chart

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	internalErrors "github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

const chartTestJSON = `{
  "chart": {
    "result": [{
      "meta": {
        "currency": "USD", "symbol": "AAPL", "exchangeName": "NMS",
        "instrumentType": "EQUITY", "firstTradeDate": 345479400,
        "fullExchangeName": "NasdaqGS", "regularMarketTime": 1674758400,
        "gmtoffset": -18000, "hasPrePostMarketData": true,
        "timezone": "EST", "exchangeTimezoneName": "America/New_York",
        "regularMarketPrice": 142.53, "chartPreviousClose": 129.79,
        "previousClose": 137.34, "regularMarketDayHigh": 143.12,
        "regularMarketDayLow": 140.45, "regularMarketVolume": 54321098,
        "longName": "Apple Inc.", "shortName": "AAPL", "scale": 3,
        "priceHint": 2, "dataGranularity": "1d", "range": "1mo",
        "validRanges": ["1d","5d","1mo","3mo","6mo","1y","2y","5y","10y","ytd","max"],
        "fiftyTwoWeekHigh": 176.15, "fiftyTwoWeekLow": 124.17
      },
      "timestamp": [1672531200, 1672617600],
      "indicators": {
        "quote": [{"open": [130.28, 130.9], "high": [133.47, 132.76], "low": [129.89, 130.22], "close": [133.0, 131.7], "volume": [100000, 95000]}],
        "adjclose": [{"adjclose": [132.5, 131.2]}]
      },
      "events": {
        "dividends": {"1672531200": {"amount": 0.23, "date": 1672531200, "description": "0.230 Dividend"}},
        "splits": {}
      }
    }],
    "error": null
  }
}`

func newTestFetcher(srv *httptest.Server) *fetch.Fetcher {
	return fetch.NewFetcher(fetch.Config{
		QueryHost:  srv.URL,
		HTTPClient: srv.Client(),
		Concurrency: 4,
		Logger:      types.NewDefaultLogger(),
	})
}

func newChartServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestGetMetaParsedCorrectly(t *testing.T) {
	srv := newChartServer(chartTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := result.Meta
	if meta.Currency != "USD" {
		t.Errorf("Currency = %q, want %q", meta.Currency, "USD")
	}
	if meta.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", meta.Symbol, "AAPL")
	}
	if meta.ExchangeName != "NMS" {
		t.Errorf("ExchangeName = %q, want %q", meta.ExchangeName, "NMS")
	}
	if meta.InstrumentType != "EQUITY" {
		t.Errorf("InstrumentType = %q, want %q", meta.InstrumentType, "EQUITY")
	}
	if meta.FullExchangeName != "NasdaqGS" {
		t.Errorf("FullExchangeName = %q, want %q", meta.FullExchangeName, "NasdaqGS")
	}
	if meta.LongName != "Apple Inc." {
		t.Errorf("LongName = %q, want %q", meta.LongName, "Apple Inc.")
	}
	if meta.Timezone != "EST" {
		t.Errorf("Timezone = %q, want %q", meta.Timezone, "EST")
	}
	if meta.ExchangeTimezoneName != "America/New_York" {
		t.Errorf("ExchangeTimezoneName = %q, want %q", meta.ExchangeTimezoneName, "America/New_York")
	}
	if meta.RegularMarketPrice == nil || *meta.RegularMarketPrice != 142.53 {
		t.Errorf("RegularMarketPrice = %v, want 142.53", meta.RegularMarketPrice)
	}
	if meta.FiftyTwoWeekHigh == nil || *meta.FiftyTwoWeekHigh != 176.15 {
		t.Errorf("FiftyTwoWeekHigh = %v, want 176.15", meta.FiftyTwoWeekHigh)
	}
	if meta.FiftyTwoWeekLow == nil || *meta.FiftyTwoWeekLow != 124.17 {
		t.Errorf("FiftyTwoWeekLow = %v, want 124.17", meta.FiftyTwoWeekLow)
	}
}

func TestGetQuotesWithOHLCV(t *testing.T) {
	srv := newChartServer(chartTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Quotes) != 2 {
		t.Fatalf("len(Quotes) = %d, want 2", len(result.Quotes))
	}

	q0 := result.Quotes[0]
	if q0.Open == nil || *q0.Open != 130.28 {
		t.Errorf("Quotes[0].Open = %v, want 130.28", q0.Open)
	}
	if q0.High == nil || *q0.High != 133.47 {
		t.Errorf("Quotes[0].High = %v, want 133.47", q0.High)
	}
	if q0.Low == nil || *q0.Low != 129.89 {
		t.Errorf("Quotes[0].Low = %v, want 129.89", q0.Low)
	}
	if q0.Close == nil || *q0.Close != 133.0 {
		t.Errorf("Quotes[0].Close = %v, want 133.0", q0.Close)
	}
	if q0.Volume == nil || *q0.Volume != 100000 {
		t.Errorf("Quotes[0].Volume = %v, want 100000", q0.Volume)
	}

	q1 := result.Quotes[1]
	if q1.Open == nil || *q1.Open != 130.9 {
		t.Errorf("Quotes[1].Open = %v, want 130.9", q1.Open)
	}
	if q1.High == nil || *q1.High != 132.76 {
		t.Errorf("Quotes[1].High = %v, want 132.76", q1.High)
	}
	if q1.Low == nil || *q1.Low != 130.22 {
		t.Errorf("Quotes[1].Low = %v, want 130.22", q1.Low)
	}
	if q1.Close == nil || *q1.Close != 131.7 {
		t.Errorf("Quotes[1].Close = %v, want 131.7", q1.Close)
	}
	if q1.Volume == nil || *q1.Volume != 95000 {
		t.Errorf("Quotes[1].Volume = %v, want 95000", q1.Volume)
	}
}

func TestGetAdjClosePopulated(t *testing.T) {
	srv := newChartServer(chartTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Quotes) < 2 {
		t.Fatalf("len(Quotes) = %d, want at least 2", len(result.Quotes))
	}
	if result.Quotes[0].AdjClose == nil || *result.Quotes[0].AdjClose != 132.5 {
		t.Errorf("Quotes[0].AdjClose = %v, want 132.5", result.Quotes[0].AdjClose)
	}
	if result.Quotes[1].AdjClose == nil || *result.Quotes[1].AdjClose != 131.2 {
		t.Errorf("Quotes[1].AdjClose = %v, want 131.2", result.Quotes[1].AdjClose)
	}
}

func TestGetEventsParsed(t *testing.T) {
	srv := newChartServer(chartTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Events == nil {
		t.Fatal("Events is nil")
	}
	div, ok := result.Events.Dividends[1672531200]
	if !ok {
		t.Fatal("dividend with date 1672531200 not found")
	}
	if div.Amount != 0.23 {
		t.Errorf("dividend Amount = %v, want 0.23", div.Amount)
	}
	if div.Date != 1672531200 {
		t.Errorf("dividend Date = %d, want 1672531200", div.Date)
	}
	if div.Description != "0.230 Dividend" {
		t.Errorf("dividend Description = %q, want %q", div.Description, "0.230 Dividend")
	}
}

func TestGetInvalidIntervalReturnsValidationError(t *testing.T) {
	srv := newChartServer(chartTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Interval: "invalid",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *errors.InvalidOptionsError", err)
	}
}

func TestGetDefaultsApplied(t *testing.T) {
	var capturedURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(chartTestJSON))
	}))
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedURL, "interval=1d") {
		t.Errorf("default interval not applied: URL = %q", capturedURL)
	}
	if !strings.Contains(capturedURL, "events=div%7Csplit%7Cearn") {
		t.Errorf("default events not applied: URL = %q", capturedURL)
	}
}

// Verify the test JSON is valid.
func init() {
	var v interface{}
	if err := json.Unmarshal([]byte(chartTestJSON), &v); err != nil {
		panic("chartTestJSON is not valid JSON: " + err.Error())
	}
}
