package fundamentals

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalErrors "github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

const fundamentalsTestJSON = `{
  "finance": {
    "result": [{
      "timeseries": {
        "quarterlyTotalRevenue": [
          {"asOfDate": "2023-09-30", "reportedValue": {"raw": 89498}},
          {"asOfDate": "2023-06-30", "reportedValue": {"raw": 81479}}
        ],
        "quarterlyNetIncome": [
          {"asOfDate": "2023-09-30", "reportedValue": {"raw": 22956}},
          {"asOfDate": "2023-06-30", "reportedValue": {"raw": 19881}}
        ]
      },
      "meta": {"currency": "USD", "symbol": "AAPL", "type": "quarterlyTotalRevenue,quarterlyNetIncome"}
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

func newFundamentalsServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestGetReturnsResults(t *testing.T) {
	srv := newFundamentalsServer(fundamentalsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	svc.SetQueryHost(srv.URL)

	results, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Module:  "financials",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	// First result should be 2023-09-30
	if results[0].Date.Year() != 2023 || results[0].Date.Month() != 9 || results[0].Date.Day() != 30 {
		t.Errorf("results[0].Date = %v, want 2023-09-30", results[0].Date)
	}
	if results[0].PeriodType != "quarterly" {
		t.Errorf("results[0].PeriodType = %q, want %q", results[0].PeriodType, "quarterly")
	}

	revVal := results[0].Fields["TotalRevenue"]
	if revVal == nil || *revVal != 89498 {
		t.Errorf("results[0].Fields[TotalRevenue] = %v, want 89498", revVal)
	}
	niVal := results[0].Fields["NetIncome"]
	if niVal == nil || *niVal != 22956 {
		t.Errorf("results[0].Fields[NetIncome] = %v, want 22956", niVal)
	}

	// Second result should be 2023-06-30
	revVal2 := results[1].Fields["TotalRevenue"]
	if revVal2 == nil || *revVal2 != 81479 {
		t.Errorf("results[1].Fields[TotalRevenue] = %v, want 81479", revVal2)
	}
}

func TestGetEmptySymbolReturnsError(t *testing.T) {
	srv := newFundamentalsServer(fundamentalsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Module:  "financials",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *InvalidOptionsError", err)
	}
}

func TestGetNilOptionsReturnsError(t *testing.T) {
	srv := newFundamentalsServer(fundamentalsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var invOptsErr *internalErrors.InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("error type = %T, want *InvalidOptionsError", err)
	}
}

func TestGetInvalidModuleReturnsError(t *testing.T) {
	srv := newFundamentalsServer(fundamentalsTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Module:  "invalid",
	})
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
	srv := newFundamentalsServer(errorJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	svc.SetQueryHost(srv.URL)

	_, err := svc.Get(context.Background(), "AAPL", &Options{
		Period1: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Module:  "financials",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *internalErrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("error type = %T, want *HTTPError", err)
	}
}

func TestBuildTimeseriesQueryKeys(t *testing.T) {
	tests := []struct {
		name       string
		periodType string
		module     string
		wantEmpty  bool
	}{
		{"financials", "quarterly", "financials", false},
		{"balance-sheet", "annual", "balance-sheet", false},
		{"cash-flow", "quarterly", "cash-flow", false},
		{"all", "quarterly", "all", false},
		{"invalid module", "quarterly", "bogus", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTimeseriesQueryKeys(tt.periodType, tt.module)
			if (got == "") != tt.wantEmpty {
				t.Errorf("buildTimeseriesQueryKeys(%q, %q) = %q, wantEmpty = %v", tt.periodType, tt.module, got, tt.wantEmpty)
			}
			if got != "" && !contains(got, tt.periodType) {
				t.Errorf("result %q should contain periodType %q", got, tt.periodType)
			}
		})
	}
}

func TestParseFundamentalsTimeSeriesNilResponse(t *testing.T) {
	results, err := parseFundamentalsTimeSeries(nil, "quarterly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for nil response, got %v", results)
	}
}

func TestParseFundamentalsTimeSeriesEmptyResults(t *testing.T) {
	raw := &fundamentalsTimeSeriesRawResponse{}
	results, err := parseFundamentalsTimeSeries(raw, "quarterly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty response, got %v", results)
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
