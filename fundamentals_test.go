package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testFundamentalsTimeSeriesJSON = `{
  "finance": {
    "result": [{
      "timeseries": {
        "quarterlyTotalRevenue": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 89498000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 81433000000}}
        ],
        "quarterlyNetIncome": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 22956000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 19881000000}}
        ],
        "quarterlyOperatingIncome": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 26969000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 23298000000}}
        ]
      },
      "meta": {
        "currency": "USD",
        "symbol": "AAPL",
        "type": "quarterly"
      }
    }],
    "error": null
  }
}`

const testFundamentalsTimeSeriesBalanceSheetJSON = `{
  "finance": {
    "result": [{
      "timeseries": {
        "quarterlyTotalAssets": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 352583000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 335516000000}}
        ],
        "quarterlyTotalLiabilities": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 290437000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 274828000000}}
        ],
        "quarterlyStockholdersEquity": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 62146000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 60688000000}}
        ]
      },
      "meta": {
        "currency": "USD",
        "symbol": "AAPL",
        "type": "quarterly"
      }
    }],
    "error": null
  }
}`

const testFundamentalsTimeSeriesCashFlowJSON = `{
  "finance": {
    "result": [{
      "timeseries": {
        "quarterlyFreeCashFlow": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 22505000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 19165000000}}
        ],
        "quarterlyOperatingCashFlow": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": 26543000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": 23048000000}}
        ],
        "quarterlyCapitalExpenditure": [
          {"asOfDate": "2023-09-30", "periodType": "3M", "reportedValue": {"raw": -4038000000}},
          {"asOfDate": "2023-06-30", "periodType": "3M", "reportedValue": {"raw": -3883000000}}
        ]
      },
      "meta": {
        "currency": "USD",
        "symbol": "AAPL",
        "type": "quarterly"
      }
    }],
    "error": null
  }
}`

const testFundamentalsTimeSeriesEmptyJSON = `{
  "finance": {
    "result": [],
    "error": null
  }
}`

const testFundamentalsTimeSeriesErrorJSON = `{
  "finance": {
    "result": [],
    "error": {
      "code": "Not Found",
      "description": "No data found for symbol"
    }
  }
}`

func newFundamentalsTestClient(srv *httptest.Server) *Client {
	client := NewClient(&Config{})
	srvClient := srv.Client()
	client.httpClient.Transport = srvClient.Transport
	client.fundamentalsQueryHost = srv.URL
	return client
}

func TestFundamentalsTimeSeriesReturnsData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the path
		if r.URL.Path != "/ws/fundamentals-timeseries/v1/finance/timeseries/AAPL" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify type parameter exists
		typeParam := r.URL.Query().Get("type")
		if typeParam == "" {
			t.Error("type parameter is required")
		}

		// Verify it contains quarterly prefix
		if !containsPrefix(typeParam, "quarterly") {
			t.Errorf("expected type to contain quarterly prefix, got: %s", typeParam)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	results, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "financials",
		Type:    "quarterly",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify first result
	r0 := results[0]
	wantDate0 := time.Date(2023, 9, 30, 0, 0, 0, 0, time.UTC)
	if !r0.Date.Equal(wantDate0) {
		t.Errorf("result[0] date: expected %v, got %v", wantDate0, r0.Date)
	}

	// Verify fields
	if r0.Fields == nil {
		t.Fatal("result[0] fields should not be nil")
	}

	if totalRevenue := r0.Fields["TotalRevenue"]; totalRevenue == nil || *totalRevenue != 89498000000 {
		t.Errorf("result[0] TotalRevenue: expected 89498000000, got %v", totalRevenue)
	}

	if netIncome := r0.Fields["NetIncome"]; netIncome == nil || *netIncome != 22956000000 {
		t.Errorf("result[0] NetIncome: expected 22956000000, got %v", netIncome)
	}
}

func TestFundamentalsTimeSeriesBalanceSheet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesBalanceSheetJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	results, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "balance-sheet",
		Type:    "quarterly",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if totalAssets := results[0].Fields["TotalAssets"]; totalAssets == nil || *totalAssets != 352583000000 {
		t.Errorf("result[0] TotalAssets: expected 352583000000, got %v", totalAssets)
	}
}

func TestFundamentalsTimeSeriesCashFlow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesCashFlowJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	results, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "cash-flow",
		Type:    "quarterly",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if freeCashFlow := results[0].Fields["FreeCashFlow"]; freeCashFlow == nil || *freeCashFlow != 22505000000 {
		t.Errorf("result[0] FreeCashFlow: expected 22505000000, got %v", freeCashFlow)
	}
}

func TestFundamentalsTimeSeriesEmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesEmptyJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	results, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "financials",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty response, got %d", len(results))
	}
}

func TestFundamentalsTimeSeriesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesErrorJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "financials",
	})
	if err == nil {
		t.Fatal("expected error for API error response, got nil")
	}

	var httpErr *HTTPError
	if !isHTTPError(err) {
		t.Errorf("expected HTTPError, got %T: %v", err, err)
	}
	_ = httpErr
}

func TestFundamentalsTimeSeriesValidationMissingSymbol(t *testing.T) {
	client := NewClient(&Config{})

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "financials",
	})
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !isInvalidOptionsError(err) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestFundamentalsTimeSeriesValidationMissingModule(t *testing.T) {
	client := NewClient(&Config{})

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
	})
	if err == nil {
		t.Fatal("expected error for missing module, got nil")
	}

	if !isInvalidOptionsError(err) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestFundamentalsTimeSeriesValidationInvalidModule(t *testing.T) {
	client := NewClient(&Config{})

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid module, got nil")
	}

	if !isInvalidOptionsError(err) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestFundamentalsTimeSeriesValidationInvalidType(t *testing.T) {
	client := NewClient(&Config{})

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "financials",
		Type:    "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}

	if !isInvalidOptionsError(err) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestFundamentalsTimeSeriesValidationMissingPeriod1(t *testing.T) {
	client := NewClient(&Config{})

	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Module: "financials",
	})
	if err == nil {
		t.Fatal("expected error for missing Period1, got nil")
	}

	if !isInvalidOptionsError(err) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestFundamentalsTimeSeriesAnnual(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify type parameter contains annual prefix
		typeParam := r.URL.Query().Get("type")
		if !containsPrefix(typeParam, "annual") {
			t.Errorf("expected type to contain annual prefix, got: %s", typeParam)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "financials",
		Type:    "annual",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFundamentalsTimeSeriesAllModule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		typeParam := r.URL.Query().Get("type")

		// Verify type contains keys from all three modules
		hasFinancials := containsPrefix(typeParam, "quarterlyTotalRevenue")
		hasBalanceSheet := containsPrefix(typeParam, "quarterlyTotalAssets")
		hasCashFlow := containsPrefix(typeParam, "quarterlyFreeCashFlow")

		if !hasFinancials || !hasBalanceSheet || !hasCashFlow {
			t.Errorf("expected type to contain keys from all modules, got: %s", typeParam)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "all",
		Type:    "quarterly",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFundamentalsTimeSeriesPeriod2Default(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period2 := r.URL.Query().Get("period2")
		if period2 == "" {
			t.Error("expected period2 to be set automatically")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testFundamentalsTimeSeriesJSON))
	}))
	defer srv.Close()

	client := newFundamentalsTestClient(srv)

	period1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FundamentalsTimeSeries(context.Background(), "AAPL", &FundamentalsTimeSeriesOptions{
		Period1: period1,
		Module:  "financials",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBuildTimeseriesQueryKeys(t *testing.T) {
	tests := []struct {
		periodType string
		module     string
		wantPrefix string
	}{
		{"quarterly", "financials", "quarterlyTotalRevenue"},
		{"annual", "financials", "annualTotalRevenue"},
		{"trailing", "financials", "trailingTotalRevenue"},
		{"quarterly", "balance-sheet", "quarterlyNetDebt"},
		{"quarterly", "cash-flow", "quarterlyFreeCashFlow"},
	}

	for _, tt := range tests {
		result := buildTimeseriesQueryKeys(tt.periodType, tt.module)
		if !containsPrefix(result, tt.wantPrefix) {
			t.Errorf("buildTimeseriesQueryKeys(%q, %q) should contain %q, got: %s",
				tt.periodType, tt.module, tt.wantPrefix, result)
		}
	}
}

func TestBuildTimeseriesQueryKeysInvalidModule(t *testing.T) {
	result := buildTimeseriesQueryKeys("quarterly", "invalid")
	if result != "" {
		t.Errorf("expected empty string for invalid module, got: %s", result)
	}
}

// Helper functions

func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix ||
		len(s) > len(prefix) && containsSubstring(s, prefix+",")
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func isHTTPError(err error) bool {
	for e := err; e != nil; e = unwrapError(e) {
		if _, ok := e.(*HTTPError); ok {
			return true
		}
	}
	return false
}
