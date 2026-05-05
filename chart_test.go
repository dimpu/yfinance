package yahoofinance

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testChartJSON = `{
  "chart": {
    "result": [{
      "meta": {
        "currency": "USD",
        "symbol": "AAPL",
        "exchangeName": "NMS",
        "instrumentType": "EQUITY",
        "firstTradeDate": 345479400,
        "fullExchangeName": "NasdaqGS",
        "regularMarketTime": 1674758400,
        "gmtoffset": -18000,
        "hasPrePostMarketData": true,
        "timezone": "EST",
        "exchangeTimezoneName": "America/New_York",
        "regularMarketPrice": 142.53,
        "chartPreviousClose": 129.79,
        "previousClose": 137.34,
        "regularMarketDayHigh": 143.12,
        "regularMarketDayLow": 140.45,
        "regularMarketVolume": 54321098,
        "longName": "Apple Inc.",
        "shortName": "AAPL",
        "scale": 3,
        "priceHint": 2,
        "dataGranularity": "1d",
        "range": "1mo",
        "validRanges": ["1d","5d","1mo","3mo","6mo","1y","2y","5y","10y","ytd","max"],
        "fiftyTwoWeekHigh": 176.15,
        "fiftyTwoWeekLow": 124.17
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
          }
        },
        "splits": {}
      }
    }],
    "error": null
  }
}`

func newChartTestClient(srv *httptest.Server) *Client {
	client := NewClient(&Config{QueryHost: srv.URL})
	srvClient := srv.Client()
	client.httpClient.Transport = srvClient.Transport
	return client
}

func TestChartMetaParsedCorrectly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testChartJSON))
	}))
	defer srv.Close()

	client := newChartTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := client.Chart(context.Background(), "AAPL", &ChartOptions{
		Period1:  period1,
		Interval: "1d",
		Events:   "div|split|earn",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	meta := result.Meta
	if meta.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", meta.Currency)
	}
	if meta.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", meta.Symbol)
	}
	if meta.ExchangeName != "NMS" {
		t.Errorf("expected exchangeName NMS, got %s", meta.ExchangeName)
	}
	if meta.InstrumentType != "EQUITY" {
		t.Errorf("expected instrumentType EQUITY, got %s", meta.InstrumentType)
	}
	if meta.LongName != "Apple Inc." {
		t.Errorf("expected longName Apple Inc., got %s", meta.LongName)
	}
	if meta.ShortName != "AAPL" {
		t.Errorf("expected shortName AAPL, got %s", meta.ShortName)
	}
	if meta.Timezone != "EST" {
		t.Errorf("expected timezone EST, got %s", meta.Timezone)
	}
	if meta.ExchangeTimezoneName != "America/New_York" {
		t.Errorf("expected exchangeTimezoneName America/New_York, got %s", meta.ExchangeTimezoneName)
	}
	if meta.DataGranularity != "1d" {
		t.Errorf("expected dataGranularity 1d, got %s", meta.DataGranularity)
	}
	if meta.Range != "1mo" {
		t.Errorf("expected range 1mo, got %s", meta.Range)
	}
	if meta.RegularMarketPrice == nil || *meta.RegularMarketPrice != 142.53 {
		t.Errorf("expected regularMarketPrice 142.53, got %v", meta.RegularMarketPrice)
	}
	if meta.ChartPreviousClose == nil || *meta.ChartPreviousClose != 129.79 {
		t.Errorf("expected chartPreviousClose 129.79, got %v", meta.ChartPreviousClose)
	}
	if meta.FiftyTwoWeekHigh == nil || *meta.FiftyTwoWeekHigh != 176.15 {
		t.Errorf("expected fiftyTwoWeekHigh 176.15, got %v", meta.FiftyTwoWeekHigh)
	}
	if meta.FiftyTwoWeekLow == nil || *meta.FiftyTwoWeekLow != 124.17 {
		t.Errorf("expected fiftyTwoWeekLow 124.17, got %v", meta.FiftyTwoWeekLow)
	}
	if !meta.HasPrePostMarketData {
		t.Error("expected hasPrePostMarketData true")
	}
	if len(meta.ValidRanges) != 11 {
		t.Errorf("expected 11 validRanges, got %d", len(meta.ValidRanges))
	}
}

func TestChartQuotesWithOHLCV(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testChartJSON))
	}))
	defer srv.Close()

	client := newChartTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := client.Chart(context.Background(), "AAPL", &ChartOptions{
		Period1:  period1,
		Interval: "1d",
		Events:   "div|split|earn",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(result.Quotes))
	}

	// First quote
	q0 := result.Quotes[0]
	wantDate0 := time.Unix(1672531200, 0)
	if !q0.Date.Equal(wantDate0) {
		t.Errorf("quote[0] date: expected %v, got %v", wantDate0, q0.Date)
	}
	if q0.Open == nil || *q0.Open != 130.28 {
		t.Errorf("quote[0] open: expected 130.28, got %v", q0.Open)
	}
	if q0.High == nil || *q0.High != 133.47 {
		t.Errorf("quote[0] high: expected 133.47, got %v", q0.High)
	}
	if q0.Low == nil || *q0.Low != 129.89 {
		t.Errorf("quote[0] low: expected 129.89, got %v", q0.Low)
	}
	if q0.Close == nil || *q0.Close != 133.0 {
		t.Errorf("quote[0] close: expected 133.0, got %v", q0.Close)
	}
	if q0.Volume == nil || *q0.Volume != 100000 {
		t.Errorf("quote[0] volume: expected 100000, got %v", q0.Volume)
	}

	// Second quote
	q1 := result.Quotes[1]
	wantDate1 := time.Unix(1672617600, 0)
	if !q1.Date.Equal(wantDate1) {
		t.Errorf("quote[1] date: expected %v, got %v", wantDate1, q1.Date)
	}
	if q1.Open == nil || *q1.Open != 130.9 {
		t.Errorf("quote[1] open: expected 130.9, got %v", q1.Open)
	}
	if q1.High == nil || *q1.High != 132.76 {
		t.Errorf("quote[1] high: expected 132.76, got %v", q1.High)
	}
	if q1.Low == nil || *q1.Low != 130.22 {
		t.Errorf("quote[1] low: expected 130.22, got %v", q1.Low)
	}
	if q1.Close == nil || *q1.Close != 131.7 {
		t.Errorf("quote[1] close: expected 131.7, got %v", q1.Close)
	}
	if q1.Volume == nil || *q1.Volume != 95000 {
		t.Errorf("quote[1] volume: expected 95000, got %v", q1.Volume)
	}
}

func TestChartAdjClosePopulated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testChartJSON))
	}))
	defer srv.Close()

	client := newChartTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := client.Chart(context.Background(), "AAPL", &ChartOptions{
		Period1:  period1,
		Interval: "1d",
		Events:   "div|split|earn",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Quotes) < 2 {
		t.Fatalf("expected at least 2 quotes, got %d", len(result.Quotes))
	}

	if result.Quotes[0].AdjClose == nil || *result.Quotes[0].AdjClose != 132.5 {
		t.Errorf("quote[0] adjclose: expected 132.5, got %v", result.Quotes[0].AdjClose)
	}
	if result.Quotes[1].AdjClose == nil || *result.Quotes[1].AdjClose != 131.2 {
		t.Errorf("quote[1] adjclose: expected 131.2, got %v", result.Quotes[1].AdjClose)
	}
}

func TestChartEventsParsed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testChartJSON))
	}))
	defer srv.Close()

	client := newChartTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := client.Chart(context.Background(), "AAPL", &ChartOptions{
		Period1:  period1,
		Interval: "1d",
		Events:   "div|split|earn",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Events == nil {
		t.Fatal("expected events to be non-nil")
	}
	if len(result.Events.Dividends) != 1 {
		t.Fatalf("expected 1 dividend, got %d", len(result.Events.Dividends))
	}
	div, ok := result.Events.Dividends[1672531200]
	if !ok {
		t.Fatal("expected dividend at timestamp 1672531200")
	}
	if div.Amount != 0.23 {
		t.Errorf("expected dividend amount 0.23, got %f", div.Amount)
	}
	if div.Description != "0.230 Dividend" {
		t.Errorf("expected dividend description '0.230 Dividend', got %s", div.Description)
	}
}

func TestChartInvalidIntervalReturnsValidationError(t *testing.T) {
	client := NewClient(&Config{})

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.Chart(context.Background(), "AAPL", &ChartOptions{
		Period1:  period1,
		Interval: "invalid",
		Events:   "div|split|earn",
	})
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}

	var invOptsErr *InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestChartDefaultsApplied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify default query params
		interval := r.URL.Query().Get("interval")
		if interval != "1d" {
			t.Errorf("expected default interval=1d, got %s", interval)
		}
		events := r.URL.Query().Get("events")
		if events != "div|split|earn" {
			t.Errorf("expected default events=div|split|earn, got %s", events)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testChartJSON))
	}))
	defer srv.Close()

	client := newChartTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.Chart(context.Background(), "AAPL", &ChartOptions{
		Period1: period1,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
