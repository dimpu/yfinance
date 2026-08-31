package insights

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

const insightsTestJSON = `{
  "finance": {
    "result": {
      "symbol": "AAPL",
      "instrumentInfo": {
        "keyTechnicals": {
          "provider": "TipRanks",
          "support": 170.5,
          "resistance": 185.0,
          "stopLoss": 165.0
        },
        "technicalEvents": {
          "provider": "Trading Central",
          "sector": "Technology",
          "shortTermOutlook": {
            "stateDescription": "Bullish",
            "direction": "Up",
            "score": 8.0,
            "scoreDescription": "Strong Buy"
          }
        },
        "valuation": {
          "color": "green",
          "description": "Undervalued",
          "discount": 15.5,
          "provider": "TipRanks",
          "relativeValue": 1.2
        }
      },
      "companySnapshot": {
        "sectorInfo": "Technology",
        "company": {
          "innovativeness": 85.0,
          "hiring": 70.0,
          "sustainability": 60.0,
          "insiderSentiments": 55.0,
          "earningsReports": 80.0,
          "dividends": 40.0
        }
      },
      "recommendation": {
        "targetPrice": 195.0,
        "provider": "TipRanks",
        "rating": "BUY"
      },
      "reports": [
        {"id": "r1", "title": "Q3 Earnings Report", "provider": "Morningstar", "publishedAt": 1698796800}
      ],
      "sigDevs": [
        {"headline": "New product launch", "date": 1698796800}
      ]
    },
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

func newInsightsServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestGetReturnsInsightsResult(t *testing.T) {
	srv := newInsightsServer(insightsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", result.Symbol, "AAPL")
	}

	// KeyTechnicals
	if result.InstrumentInfo == nil || result.InstrumentInfo.KeyTechnicals == nil {
		t.Fatal("InstrumentInfo.KeyTechnicals is nil")
	}
	kt := result.InstrumentInfo.KeyTechnicals
	if kt.Support == nil || *kt.Support != 170.5 {
		t.Errorf("Support = %v, want 170.5", kt.Support)
	}
	if kt.Resistance == nil || *kt.Resistance != 185.0 {
		t.Errorf("Resistance = %v, want 185.0", kt.Resistance)
	}

	// Valuation
	if result.InstrumentInfo.Valuation == nil {
		t.Fatal("InstrumentInfo.Valuation is nil")
	}
	v := result.InstrumentInfo.Valuation
	if v.Description != "Undervalued" {
		t.Errorf("Valuation.Description = %q, want %q", v.Description, "Undervalued")
	}

	// Recommendation
	if result.Recommendation == nil {
		t.Fatal("Recommendation is nil")
	}
	if result.Recommendation.Rating != "BUY" {
		t.Errorf("Rating = %q, want %q", result.Recommendation.Rating, "BUY")
	}
	if result.Recommendation.TargetPrice == nil || *result.Recommendation.TargetPrice != 195.0 {
		t.Errorf("TargetPrice = %v, want 195.0", result.Recommendation.TargetPrice)
	}

	// Reports
	if len(result.Reports) != 1 {
		t.Fatalf("len(Reports) = %d, want 1", len(result.Reports))
	}
	if result.Reports[0].Title != "Q3 Earnings Report" {
		t.Errorf("Reports[0].Title = %q, want %q", result.Reports[0].Title, "Q3 Earnings Report")
	}

	// SigDevs
	if len(result.SigDevs) != 1 {
		t.Fatalf("len(SigDevs) = %d, want 1", len(result.SigDevs))
	}
	if result.SigDevs[0].Headline != "New product launch" {
		t.Errorf("SigDevs[0].Headline = %q, want %q", result.SigDevs[0].Headline, "New product launch")
	}
}

func TestGetEmptySymbolReturnsError(t *testing.T) {
	srv := newInsightsServer(insightsTestJSON)
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
	errorJSON := `{"finance":{"result":null,"error":{"code":"Not Found","description":"No insights found"}}}`
	srv := newInsightsServer(errorJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *internalErrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("error type = %T, want *HTTPError", err)
	}
}

func TestGetEmptyResultReturnsHTTPError(t *testing.T) {
	emptyJSON := `{"finance":{"result":null,"error":null}}`
	srv := newInsightsServer(emptyJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *internalErrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("error type = %T, want *HTTPError", err)
	}
}

func TestGetWithOptions(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(insightsTestJSON))
	}))
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", &Options{
		Lang:         "fr-FR",
		Region:       "FR",
		ReportsCount: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedQuery == "" {
		t.Fatal("expected query params, got empty")
	}
}
