package screener

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

const screenerTestJSON = `{
  "finance": {
    "result": [{
      "id": "most_actives",
      "title": "Most Actives",
      "description": "Most active stocks",
      "canonicalName": "most_actives",
      "count": 5,
      "total": 5,
      "quotes": [
        {
          "symbol": "TSLA",
          "shortName": "Tesla, Inc.",
          "quoteType": "EQUITY",
          "exchange": "NMS",
          "currency": "USD",
          "marketState": "REGULAR",
          "regularMarketPrice": 245.67,
          "regularMarketChange": 5.43,
          "regularMarketChangePercent": 2.26,
          "regularMarketVolume": 98765432,
          "fiftyTwoWeekLow": 101.81,
          "fiftyTwoWeekHigh": 299.29
        }
      ]
    }],
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

func newScreenerServer(body string) *httptest.Server {
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

func TestGetReturnsScreenerResult(t *testing.T) {
	srv := newScreenerServer(screenerTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), MostActives, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "most_actives" {
		t.Errorf("ID = %q, want %q", result.ID, "most_actives")
	}
	if result.Title != "Most Actives" {
		t.Errorf("Title = %q, want %q", result.Title, "Most Actives")
	}
	if result.Count != 5 {
		t.Errorf("Count = %d, want 5", result.Count)
	}

	if len(result.Quotes) != 1 {
		t.Fatalf("len(Quotes) = %d, want 1", len(result.Quotes))
	}
	q := result.Quotes[0]
	if q.Symbol != "TSLA" {
		t.Errorf("Quotes[0].Symbol = %q, want %q", q.Symbol, "TSLA")
	}
	if q.RegularMarketPrice == nil || *q.RegularMarketPrice != 245.67 {
		t.Errorf("Quotes[0].RegularMarketPrice = %v, want 245.67", q.RegularMarketPrice)
	}
	if q.FiftyTwoWeekLow == nil || *q.FiftyTwoWeekLow != 101.81 {
		t.Errorf("Quotes[0].FiftyTwoWeekLow = %v, want 101.81", q.FiftyTwoWeekLow)
	}
}

func TestGetAPIErrorReturnsHTTPError(t *testing.T) {
	errorJSON := `{"finance":{"result":null,"error":{"code":"Not Found","description":"Screener not found"}}}`
	srv := newScreenerServer(errorJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), MostActives, nil)
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
	srv := newScreenerServer(emptyJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), MostActives, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *internalErrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("error type = %T, want *HTTPError", err)
	}
}

func TestGetWithCountOption(t *testing.T) {
	var capturedQuery string
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(screenerTestJSON))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), DayGainers, &Options{Count: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedQuery == "" {
		t.Fatal("expected query params, got empty")
	}
}

func TestPredefinedScreenerModuleConstants(t *testing.T) {
	modules := map[PredefinedScreenerModule]string{
		AggressiveSmallCaps:      "aggressive_small_caps",
		ConservativeForeignFunds: "conservative_foreign_funds",
		DayGainers:               "day_gainers",
		DayLosers:                "day_losers",
		GrowthTechStocks:         "growth_technology_stocks",
		HighYieldBond:            "high_yield_bond",
		MostActives:              "most_actives",
		MostShorted:              "most_shorted_stocks",
		PortfolioAnchors:         "portfolio_anchors",
		SmallCapGainers:          "small_cap_gainers",
		SolidLargeGrowthFunds:    "solid_large_growth_funds",
		SolidMidcapGrowthFunds:   "solid_midcap_growth_funds",
		TopMutualFunds:           "top_mutual_funds",
		UndervaluedGrowth:        "undervalued_growth_stocks",
		UndervaluedLargeCaps:     "undervalued_large_caps",
	}
	for k, v := range modules {
		if string(k) != v {
			t.Errorf("PredefinedScreenerModule constant mismatch: got %q, want %q", string(k), v)
		}
	}
}
