package recommendations

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

const recommendationsTestJSON = `{
  "finance": {
    "result": [
      {
        "recommendedSymbols": [
          {"score": 0.95, "symbol": "MSFT"},
          {"score": 0.88, "symbol": "GOOG"},
          {"score": 0.82, "symbol": "AMZN"}
        ],
        "symbol": "AAPL"
      }
    ],
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

func newRecommendationsServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestGetReturnsRecommendations(t *testing.T) {
	srv := newRecommendationsServer(recommendationsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	results, err := svc.Get(context.Background(), []string{"AAPL"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	if results[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", results[0].Symbol, "AAPL")
	}

	if len(results[0].RecommendedSymbols) != 3 {
		t.Fatalf("len(RecommendedSymbols) = %d, want 3", len(results[0].RecommendedSymbols))
	}

	rs0 := results[0].RecommendedSymbols[0]
	if rs0.Symbol != "MSFT" {
		t.Errorf("RecommendedSymbols[0].Symbol = %q, want %q", rs0.Symbol, "MSFT")
	}
	if rs0.Score != 0.95 {
		t.Errorf("RecommendedSymbols[0].Score = %v, want 0.95", rs0.Score)
	}
}

func TestGetEmptySymbolsReturnsError(t *testing.T) {
	srv := newRecommendationsServer(recommendationsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), []string{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *InvalidOptionsError", err)
	}
}

func TestGetAPIErrorReturnsHTTPError(t *testing.T) {
	errorJSON := `{"finance":{"result":null,"error":{"code":"Not Found","description":"No data found"}}}`
	srv := newRecommendationsServer(errorJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), []string{"AAPL"}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *internalErrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("error type = %T, want *HTTPError", err)
	}
}
