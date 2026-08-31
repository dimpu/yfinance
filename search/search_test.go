package search

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

const searchTestJSON = `{"count": 1, "quotes": [{"symbol": "AAPL", "isYahooFinance": true, "exchange": "NMS", "quoteType": "EQUITY", "score": 1.0}], "news": []}`

func newTestFetcher(srv *httptest.Server) *fetch.Fetcher {
	return fetch.NewFetcher(fetch.Config{
		QueryHost:  srv.URL,
		HTTPClient: srv.Client(),
		Concurrency: 4,
		Logger:      types.NewDefaultLogger(),
	})
}

func newSearchServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestGetReturnsSearchResult(t *testing.T) {
	srv := newSearchServer(searchTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "Apple", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Count = %d, want 1", result.Count)
	}
	if len(result.Quotes) != 1 {
		t.Fatalf("len(Quotes) = %d, want 1", len(result.Quotes))
	}
	if result.Quotes[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", result.Quotes[0].Symbol, "AAPL")
	}
}

func TestGetEmptyQueryReturnsError(t *testing.T) {
	srv := newSearchServer(searchTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *errors.InvalidOptionsError", err)
	}
}
