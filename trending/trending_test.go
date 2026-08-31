package trending

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	internalErrors "github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

const trendingTestJSON = `{
  "finance": {
    "result": [{
      "count": 5,
      "quotes": [
        {"symbol": "AAPL"},
        {"symbol": "TSLA"},
        {"symbol": "NVDA"},
        {"symbol": "AMZN"},
        {"symbol": "MSFT"}
      ],
      "jobTimestamp": 1698796800,
      "startInterval": 1698790000
    }],
    "error": null
  }
}`

func newTestFetcher(srv *httptest.Server) *fetch.Fetcher {
	return fetch.NewFetcher(fetch.Config{
		QueryHost:   srv.URL,
		HTTPClient:  srv.Client(),
		Concurrency: 4,
		Logger:      types.NewDefaultLogger(),
	})
}

func newTrendingServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestGetReturnsTrendingResult(t *testing.T) {
	srv := newTrendingServer(trendingTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "US", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 5 {
		t.Errorf("Count = %d, want 5", result.Count)
	}
	if len(result.Quotes) != 5 {
		t.Fatalf("len(Quotes) = %d, want 5", len(result.Quotes))
	}
	if result.Quotes[0].Symbol != "AAPL" {
		t.Errorf("Quotes[0].Symbol = %q, want %q", result.Quotes[0].Symbol, "AAPL")
	}
	if result.Quotes[2].Symbol != "NVDA" {
		t.Errorf("Quotes[2].Symbol = %q, want %q", result.Quotes[2].Symbol, "NVDA")
	}
	if result.JobTimestamp != 1698796800 {
		t.Errorf("JobTimestamp = %d, want 1698796800", result.JobTimestamp)
	}
}

func TestGetEmptyRegionReturnsError(t *testing.T) {
	srv := newTrendingServer(trendingTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *InvalidOptionsError", err)
	}
}

func TestGetAPIErrorReturnsHTTPError(t *testing.T) {
	errorJSON := `{"finance":{"result":null,"error":{"code":"Not Found","description":"Region not found"}}}`
	srv := newTrendingServer(errorJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "US", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *internalErrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("error type = %T, want *HTTPError", err)
	}
}

func TestGetEmptyResultReturnsHTTPError(t *testing.T) {
	emptyJSON := `{"finance":{"result":[],"error":null}}`
	srv := newTrendingServer(emptyJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "US", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *internalErrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("error type = %T, want *HTTPError", err)
	}
}

func TestGetWithOptions(t *testing.T) {
	var capturedPath string
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(trendingTestJSON))
	}))
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "GB", &Options{Lang: "en-GB", Count: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath == "" {
		t.Fatal("expected path, got empty")
	}
	if capturedQuery == "" {
		t.Fatal("expected query params, got empty")
	}
}
