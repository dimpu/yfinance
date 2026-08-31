package quote

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	internalErrors "github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

const quoteTestJSON = `{
  "quoteResponse": {
    "result": [
      {"symbol": "AAPL", "quoteType": "EQUITY", "regularMarketPrice": 150.0, "currency": "USD", "marketState": "REGULAR"},
      {"symbol": "DELISTED", "quoteType": "NONE"}
    ],
    "error": null
  }
}`

// redirectTransport routes all requests to the test server.
type redirectTransport struct {
	server *httptest.Server
	base   http.RoundTripper
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := *req.URL
	newURL.Scheme = "http"
	newURL.Host = t.server.Listener.Addr().String()
	req.URL = &newURL
	return t.base.RoundTrip(req)
}

func newTestFetcher(srv *httptest.Server) *fetch.Fetcher {
	jar, _ := cookiejar.New(nil)
	return fetch.NewFetcher(fetch.Config{
		QueryHost:  srv.URL,
		HTTPClient: &http.Client{Jar: jar, Transport: &redirectTransport{server: srv, base: http.DefaultTransport}},
		Concurrency: 4,
		Logger:      types.NewDefaultLogger(),
	})
}

func newQuoteServer(body string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	})
	return httptest.NewServer(mux)
}

func TestGetReturnsQuotes(t *testing.T) {
	srv := newQuoteServer(quoteTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	quotes, err := svc.Get(context.Background(), []string{"AAPL"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(quotes) != 1 {
		t.Fatalf("len(quotes) = %d, want 1 (NONE filtered out)", len(quotes))
	}
	if quotes[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", quotes[0].Symbol, "AAPL")
	}
	if quotes[0].RegularMarketPrice == nil || *quotes[0].RegularMarketPrice != 150.0 {
		t.Errorf("RegularMarketPrice = %v, want 150.0", quotes[0].RegularMarketPrice)
	}
}

func TestGetEmptySymbolsReturnsError(t *testing.T) {
	srv := newQuoteServer(quoteTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), []string{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *errors.InvalidOptionsError", err)
	}
}
