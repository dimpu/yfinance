package historical

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalErrors "github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/chart"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

const historicalTestJSON = `{
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

func newHistoricalServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestGetReturnsRowHistory(t *testing.T) {
	srv := newHistoricalServer(historicalTestJSON)
	defer srv.Close()

	chartSvc := chart.NewService(newTestFetcher(srv))
	svc := NewService(chartSvc)

	rows, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1:              time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		IncludeAdjustedClose: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}

	r0 := rows[0]
	if r0.Open == nil || *r0.Open != 130.28 {
		t.Errorf("rows[0].Open = %v, want 130.28", r0.Open)
	}
	if r0.High == nil || *r0.High != 133.47 {
		t.Errorf("rows[0].High = %v, want 133.47", r0.High)
	}
	if r0.Low == nil || *r0.Low != 129.89 {
		t.Errorf("rows[0].Low = %v, want 129.89", r0.Low)
	}
	if r0.Close == nil || *r0.Close != 133.0 {
		t.Errorf("rows[0].Close = %v, want 133.0", r0.Close)
	}
	if r0.Volume == nil || *r0.Volume != 100000 {
		t.Errorf("rows[0].Volume = %v, want 100000", r0.Volume)
	}
	if r0.AdjClose == nil || *r0.AdjClose != 132.5 {
		t.Errorf("rows[0].AdjClose = %v, want 132.5", r0.AdjClose)
	}
}

func TestGetInvalidOptionsReturnsError(t *testing.T) {
	srv := newHistoricalServer(historicalTestJSON)
	defer srv.Close()

	chartSvc := chart.NewService(newTestFetcher(srv))
	svc := NewService(chartSvc)

	_, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Time{},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *errors.InvalidOptionsError", err)
	}
}
