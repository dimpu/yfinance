package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testHistoricalJSON = `{
  "chart": {
    "result": [{
      "meta": {
        "currency": "USD",
        "symbol": "AAPL",
        "exchangeName": "NMS",
        "instrumentType": "EQUITY",
        "fullExchangeName": "NasdaqGS",
        "regularMarketTime": 1674758400,
        "gmtoffset": -18000,
        "timezone": "EST",
        "exchangeTimezoneName": "America/New_York",
        "regularMarketPrice": 142.53,
        "chartPreviousClose": 129.79,
        "dataGranularity": "1d",
        "range": "1mo",
        "validRanges": ["1d","5d","1mo"]
      },
      "timestamp": [1672531200, 1672617600, 1672704000],
      "indicators": {
        "quote": [{
          "open": [130.28, 130.9, null],
          "high": [133.47, 132.76, null],
          "low": [129.89, 130.22, null],
          "close": [133.0, 131.7, null],
          "volume": [100000, 95000, null]
        }],
        "adjclose": [{"adjclose": [132.5, 131.2, null]}]
      }
    }],
    "error": null
  }
}`

// testHistoricalDividendsJSON returns chart data with dividend events.
const testHistoricalDividendsJSON = `{
  "chart": {
    "result": [{
      "meta": {
        "currency": "USD",
        "symbol": "AAPL",
        "exchangeName": "NMS",
        "instrumentType": "EQUITY",
        "fullExchangeName": "NasdaqGS",
        "regularMarketTime": 1674758400,
        "gmtoffset": -18000,
        "timezone": "EST",
        "exchangeTimezoneName": "America/New_York",
        "regularMarketPrice": 142.53,
        "dataGranularity": "1d",
        "range": "1mo",
        "validRanges": ["1d","5d","1mo"]
      },
      "timestamp": [1672531200, 1672617600],
      "indicators": {
        "quote": [{
          "open": [130.28, 130.9],
          "high": [133.47, 132.76],
          "low": [129.89, 130.22],
          "close": [133.0, 131.7],
          "volume": [100000, 95000]
        }],
        "adjclose": [{"adjclose": [132.5, 131.2]}]
      },
      "events": {
        "dividends": {
          "1672531200": {
            "amount": 0.23,
            "date": 1672531200,
            "description": "0.230 Dividend"
          },
          "1672617600": {
            "amount": 0.24,
            "date": 1672617600,
            "description": "0.240 Dividend"
          }
        },
        "splits": {}
      }
    }],
    "error": null
  }
}`

// testHistoricalSplitsJSON returns chart data with split events.
const testHistoricalSplitsJSON = `{
  "chart": {
    "result": [{
      "meta": {
        "currency": "USD",
        "symbol": "TSLA",
        "exchangeName": "NMS",
        "instrumentType": "EQUITY",
        "fullExchangeName": "NasdaqGS",
        "regularMarketTime": 1674758400,
        "gmtoffset": -18000,
        "timezone": "EST",
        "exchangeTimezoneName": "America/New_York",
        "regularMarketPrice": 200.0,
        "dataGranularity": "1d",
        "range": "1mo",
        "validRanges": ["1d","5d","1mo"]
      },
      "timestamp": [1672531200],
      "indicators": {
        "quote": [{
          "open": [100.0],
          "high": [110.0],
          "low": [95.0],
          "close": [105.0],
          "volume": [500000]
        }],
        "adjclose": [{"adjclose": [105.0]}]
      },
      "events": {
        "dividends": {},
        "splits": {
          "1672531200": {
            "date": 1672531200,
            "numerator": 3,
            "denominator": 1,
            "splitRatio": "3:1"
          }
        }
      }
    }],
    "error": null
  }
}`

func newHistoricalTestClient(srv *httptest.Server) *Client {
	client := NewClient(&Config{QueryHost: srv.URL})
	srvClient := srv.Client()
	client.httpClient.Transport = srvClient.Transport
	return client
}

func TestHistoricalReturnsPriceData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHistoricalJSON))
	}))
	defer srv.Close()

	client := newHistoricalTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := client.Historical(context.Background(), "AAPL", &HistoricalOptions{
		Period1:  period1,
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Row 3 is all-null (OHLCV all nil), so should be filtered out
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows after filtering nulls, got %d", len(rows))
	}

	// First row
	r0 := rows[0]
	wantDate0 := time.Unix(1672531200, 0)
	if !r0.Date.Equal(wantDate0) {
		t.Errorf("row[0] date: expected %v, got %v", wantDate0, r0.Date)
	}
	if r0.Open == nil || *r0.Open != 130.28 {
		t.Errorf("row[0] open: expected 130.28, got %v", r0.Open)
	}
	if r0.High == nil || *r0.High != 133.47 {
		t.Errorf("row[0] high: expected 133.47, got %v", r0.High)
	}
	if r0.Low == nil || *r0.Low != 129.89 {
		t.Errorf("row[0] low: expected 129.89, got %v", r0.Low)
	}
	if r0.Close == nil || *r0.Close != 133.0 {
		t.Errorf("row[0] close: expected 133.0, got %v", r0.Close)
	}
	if r0.Volume == nil || *r0.Volume != 100000 {
		t.Errorf("row[0] volume: expected 100000, got %v", r0.Volume)
	}

	// Second row
	r1 := rows[1]
	wantDate1 := time.Unix(1672617600, 0)
	if !r1.Date.Equal(wantDate1) {
		t.Errorf("row[1] date: expected %v, got %v", wantDate1, r1.Date)
	}
	if r1.Open == nil || *r1.Open != 130.9 {
		t.Errorf("row[1] open: expected 130.9, got %v", r1.Open)
	}
}

func TestHistoricalNullRowsFiltered(t *testing.T) {
	// testHistoricalJSON has a 3rd row with all OHLCV as null
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHistoricalJSON))
	}))
	defer srv.Close()

	client := newHistoricalTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := client.Historical(context.Background(), "AAPL", &HistoricalOptions{
		Period1:  period1,
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for i, row := range rows {
		if row.Open == nil && row.High == nil && row.Low == nil && row.Close == nil && row.Volume == nil {
			t.Errorf("row[%d] should have been filtered as all-null but was not", i)
		}
	}
}

func TestHistoricalAdjCloseRenamed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHistoricalJSON))
	}))
	defer srv.Close()

	client := newHistoricalTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := client.Historical(context.Background(), "AAPL", &HistoricalOptions{
		Period1:              period1,
		Interval:             "1d",
		IncludeAdjustedClose: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rows) < 2 {
		t.Fatalf("expected at least 2 rows, got %d", len(rows))
	}

	// Verify AdjClose is populated (renamed from adjclose)
	if rows[0].AdjClose == nil || *rows[0].AdjClose != 132.5 {
		t.Errorf("row[0] adjClose: expected 132.5, got %v", rows[0].AdjClose)
	}
	if rows[1].AdjClose == nil || *rows[1].AdjClose != 131.2 {
		t.Errorf("row[1] adjClose: expected 131.2, got %v", rows[1].AdjClose)
	}
}

func TestHistoricalWithoutAdjClose(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHistoricalJSON))
	}))
	defer srv.Close()

	client := newHistoricalTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := client.Historical(context.Background(), "AAPL", &HistoricalOptions{
		Period1:              period1,
		Interval:             "1d",
		IncludeAdjustedClose: false,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rows) < 1 {
		t.Fatalf("expected at least 1 row, got %d", len(rows))
	}

	// AdjClose should not be populated when IncludeAdjustedClose is false
	if rows[0].AdjClose != nil {
		t.Errorf("row[0] adjClose: expected nil when IncludeAdjustedClose=false, got %v", rows[0].AdjClose)
	}
}

func TestHistoricalDividendsReturnsData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHistoricalDividendsJSON))
	}))
	defer srv.Close()

	client := newHistoricalTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := client.HistoricalDividends(context.Background(), "AAPL", &HistoricalOptions{
		Period1:  period1,
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 dividend rows, got %d", len(rows))
	}

	wantDate0 := time.Unix(1672531200, 0)
	if !rows[0].Date.Equal(wantDate0) {
		t.Errorf("dividend[0] date: expected %v, got %v", wantDate0, rows[0].Date)
	}
	if rows[0].Dividends != 0.23 {
		t.Errorf("dividend[0] amount: expected 0.23, got %f", rows[0].Dividends)
	}

	wantDate1 := time.Unix(1672617600, 0)
	if !rows[1].Date.Equal(wantDate1) {
		t.Errorf("dividend[1] date: expected %v, got %v", wantDate1, rows[1].Date)
	}
	if rows[1].Dividends != 0.24 {
		t.Errorf("dividend[1] amount: expected 0.24, got %f", rows[1].Dividends)
	}
}

func TestHistoricalSplitsReturnsData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHistoricalSplitsJSON))
	}))
	defer srv.Close()

	client := newHistoricalTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := client.HistoricalSplits(context.Background(), "TSLA", &HistoricalOptions{
		Period1:  period1,
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 split row, got %d", len(rows))
	}

	wantDate := time.Unix(1672531200, 0)
	if !rows[0].Date.Equal(wantDate) {
		t.Errorf("split[0] date: expected %v, got %v", wantDate, rows[0].Date)
	}
	if rows[0].StockSplits != "3:1" {
		t.Errorf("split[0] stockSplits: expected 3:1, got %s", rows[0].StockSplits)
	}
}

func TestHistoricalValidationInvalidInterval(t *testing.T) {
	client := NewClient(&Config{})

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.Historical(context.Background(), "AAPL", &HistoricalOptions{
		Period1:  period1,
		Interval: "1h",
	})
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}

	var invOptsErr *InvalidOptionsError
	if !isInvalidOptionsError(err) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
	_ = invOptsErr
}

func TestHistoricalValidationInvalidEvents(t *testing.T) {
	client := NewClient(&Config{})

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.Historical(context.Background(), "AAPL", &HistoricalOptions{
		Period1:  period1,
		Interval: "1d",
		Events:   "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid events, got nil")
	}

	if !isInvalidOptionsError(err) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestHistoricalEventsMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"history", ""},
		{"", ""},
		{"dividends", "div"},
		{"split", "split"},
	}
	for _, tt := range tests {
		got := mapHistoricalEvents(tt.input)
		if got != tt.expected {
			t.Errorf("mapHistoricalEvents(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestHistoricalDefaultsApplied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify defaults are applied to the chart call
		interval := r.URL.Query().Get("interval")
		if interval != "1d" {
			t.Errorf("expected default interval=1d, got %s", interval)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHistoricalJSON))
	}))
	defer srv.Close()

	client := newHistoricalTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.Historical(context.Background(), "AAPL", &HistoricalOptions{
		Period1: period1,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// isInvalidOptionsError checks if the error is an InvalidOptionsError.
func isInvalidOptionsError(err error) bool {
	var invOptsErr *InvalidOptionsError
	for e := err; e != nil; e = unwrapError(e) {
		if _, ok := e.(*InvalidOptionsError); ok {
			return true
		}
	}
	_ = invOptsErr
	return false
}

// unwrapError unwraps an error one level.
func unwrapError(err error) error {
	type unwrapper interface{ Unwrap() error }
	if u, ok := err.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}
