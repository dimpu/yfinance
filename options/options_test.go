package options

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

const optionsTestJSON = `{
  "optionChain": {
    "result": [{
      "underlyingSymbol": "AAPL",
      "expirationDates": [1672531200],
      "strikes": [140.0, 145.0],
      "hasMiniOptions": false,
      "quote": {"symbol": "AAPL", "quoteType": "EQUITY"},
      "options": [{"expirationDate": 1672531200, "calls": [], "puts": []}]
    }],
    "error": null
  }
}`

const optionsEmptyResultJSON = `{
  "optionChain": {
    "result": [],
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

func newOptionsServer(body string) *httptest.Server {
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

func TestGetReturnsOptionsResult(t *testing.T) {
	srv := newOptionsServer(optionsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UnderlyingSymbol != "AAPL" {
		t.Errorf("UnderlyingSymbol = %q, want %q", result.UnderlyingSymbol, "AAPL")
	}
	if len(result.Strikes) != 2 {
		t.Errorf("len(Strikes) = %d, want 2", len(result.Strikes))
	}
}

func TestGetNoResultReturnsError(t *testing.T) {
	srv := newOptionsServer(optionsEmptyResultJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", nil)
	if err == nil {
		t.Fatal("expected error for empty result, got nil")
	}
}
