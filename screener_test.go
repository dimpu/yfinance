package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testScreenerJSON = `{
  "finance": {
    "result": [
      {
        "id": "most_actives",
        "title": "Most Actives",
        "description": "Stocks with the highest trading volume today",
        "canonicalName": "most_actives",
        "count": 3,
        "total": 100,
        "quotes": [
          {
            "symbol": "AAPL",
            "shortName": "Apple Inc.",
            "longName": "Apple Inc.",
            "quoteType": "EQUITY",
            "exchange": "NMS",
            "currency": "USD",
            "marketState": "REGULAR",
            "regularMarketPrice": 178.72,
            "regularMarketChange": 2.43,
            "regularMarketChangePercent": 1.38,
            "regularMarketVolume": 54321098,
            "marketCap": 2850000000000,
            "fiftyTwoWeekLow": 124.17,
            "fiftyTwoWeekHigh": 199.62,
            "averageDailyVolume3Month": 60000000
          },
          {
            "symbol": "TSLA",
            "shortName": "Tesla, Inc.",
            "longName": "Tesla, Inc.",
            "quoteType": "EQUITY",
            "exchange": "NMS",
            "currency": "USD",
            "marketState": "REGULAR",
            "regularMarketPrice": 248.50,
            "regularMarketChange": -3.25,
            "regularMarketChangePercent": -1.29,
            "regularMarketVolume": 42156789,
            "marketCap": 790000000000,
            "fiftyTwoWeekLow": 138.80,
            "fiftyTwoWeekHigh": 299.29,
            "averageDailyVolume3Month": 45000000
          },
          {
            "symbol": "NVDA",
            "shortName": "NVIDIA Corporation",
            "longName": "NVIDIA Corporation",
            "quoteType": "EQUITY",
            "exchange": "NMS",
            "currency": "USD",
            "marketState": "REGULAR",
            "regularMarketPrice": 495.22,
            "regularMarketChange": 8.15,
            "regularMarketChangePercent": 1.67,
            "regularMarketVolume": 38912456,
            "fiftyTwoWeekLow": 138.84,
            "fiftyTwoWeekHigh": 505.48
          }
        ]
      }
    ],
    "error": null
  }
}`

func TestScreener_ParsesQuotes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/screener/predefined/saved", func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		scrIds := r.URL.Query().Get("scrIds")
		if scrIds != "most_actives" {
			t.Errorf("expected scrIds='most_actives', got %q", scrIds)
		}
		count := r.URL.Query().Get("count")
		if count != "25" {
			t.Errorf("expected count='25', got %q", count)
		}
		// Verify crumb is sent (this endpoint needs crumb)
		if crumb := r.URL.Query().Get("crumb"); crumb == "" {
			t.Error("expected crumb parameter, got empty")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testScreenerJSON))
	})

	// Handle crumb endpoints
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb123"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.Screener(context.Background(), ScreenerMostActives, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify result metadata
	if result.ID != "most_actives" {
		t.Errorf("expected id 'most_actives', got %s", result.ID)
	}
	if result.Title != "Most Actives" {
		t.Errorf("expected title 'Most Actives', got %s", result.Title)
	}
	if result.Description != "Stocks with the highest trading volume today" {
		t.Errorf("unexpected description: %s", result.Description)
	}
	if result.Count != 3 {
		t.Errorf("expected count 3, got %d", result.Count)
	}
	if result.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Total)
	}

	// Verify quotes
	if len(result.Quotes) != 3 {
		t.Fatalf("expected 3 quotes, got %d", len(result.Quotes))
	}

	// First quote (AAPL)
	aapl := result.Quotes[0]
	if aapl.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", aapl.Symbol)
	}
	if aapl.ShortName != "Apple Inc." {
		t.Errorf("expected shortName 'Apple Inc.', got %s", aapl.ShortName)
	}
	if aapl.LongName != "Apple Inc." {
		t.Errorf("expected longName 'Apple Inc.', got %s", aapl.LongName)
	}
	if aapl.QuoteType != "EQUITY" {
		t.Errorf("expected quoteType EQUITY, got %s", aapl.QuoteType)
	}
	if aapl.Exchange != "NMS" {
		t.Errorf("expected exchange NMS, got %s", aapl.Exchange)
	}
	if aapl.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", aapl.Currency)
	}
	if aapl.MarketState != "REGULAR" {
		t.Errorf("expected marketState REGULAR, got %s", aapl.MarketState)
	}
	if aapl.RegularMarketPrice == nil || *aapl.RegularMarketPrice != 178.72 {
		t.Errorf("expected regularMarketPrice 178.72, got %v", aapl.RegularMarketPrice)
	}
	if aapl.RegularMarketChange == nil || *aapl.RegularMarketChange != 2.43 {
		t.Errorf("expected regularMarketChange 2.43, got %v", aapl.RegularMarketChange)
	}
	if aapl.RegularMarketChangePercent == nil || *aapl.RegularMarketChangePercent != 1.38 {
		t.Errorf("expected regularMarketChangePercent 1.38, got %v", aapl.RegularMarketChangePercent)
	}
	if aapl.RegularMarketVolume == nil || *aapl.RegularMarketVolume != 54321098 {
		t.Errorf("expected regularMarketVolume 54321098, got %v", aapl.RegularMarketVolume)
	}
	if aapl.MarketCap == nil || *aapl.MarketCap != 2850000000000 {
		t.Errorf("expected marketCap 2850000000000, got %v", aapl.MarketCap)
	}
	if aapl.FiftyTwoWeekLow == nil || *aapl.FiftyTwoWeekLow != 124.17 {
		t.Errorf("expected fiftyTwoWeekLow 124.17, got %v", aapl.FiftyTwoWeekLow)
	}
	if aapl.FiftyTwoWeekHigh == nil || *aapl.FiftyTwoWeekHigh != 199.62 {
		t.Errorf("expected fiftyTwoWeekHigh 199.62, got %v", aapl.FiftyTwoWeekHigh)
	}
	if aapl.AverageDailyVolume3Month == nil || *aapl.AverageDailyVolume3Month != 60000000 {
		t.Errorf("expected averageDailyVolume3Month 60000000, got %v", aapl.AverageDailyVolume3Month)
	}

	// Second quote (TSLA) - verify negative change
	tsla := result.Quotes[1]
	if tsla.Symbol != "TSLA" {
		t.Errorf("expected symbol TSLA, got %s", tsla.Symbol)
	}
	if tsla.RegularMarketChange == nil || *tsla.RegularMarketChange != -3.25 {
		t.Errorf("expected regularMarketChange -3.25, got %v", tsla.RegularMarketChange)
	}

	// Third quote (NVDA) - verify missing marketCap (optional field)
	nvda := result.Quotes[2]
	if nvda.Symbol != "NVDA" {
		t.Errorf("expected symbol NVDA, got %s", nvda.Symbol)
	}
	if nvda.MarketCap != nil {
		t.Errorf("expected marketCap nil for NVDA, got %v", nvda.MarketCap)
	}
}

func TestScreener_WithOptions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/screener/predefined/saved", func(w http.ResponseWriter, r *http.Request) {
		// Verify custom count
		count := r.URL.Query().Get("count")
		if count != "50" {
			t.Errorf("expected count='50', got %q", count)
		}
		scrIds := r.URL.Query().Get("scrIds")
		if scrIds != "day_gainers" {
			t.Errorf("expected scrIds='day_gainers', got %q", scrIds)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"finance":{"result":[{"id":"day_gainers","title":"Day Gainers","count":0,"total":0,"quotes":[]}],"error":null}}`))
	})

	// Handle crumb endpoints
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb123"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	opts := &ScreenerOptions{
		Count: 50,
	}
	result, err := client.Screener(context.Background(), ScreenerDayGainers, opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != "day_gainers" {
		t.Errorf("expected id 'day_gainers', got %s", result.ID)
	}
}

func TestScreener_AllPredefinedScreeners(t *testing.T) {
	// Verify all screener constants are valid
	screeners := []PredefinedScreenerModule{
		ScreenerAggressiveSmallCaps,
		ScreenerConservativeForeignFunds,
		ScreenerDayGainers,
		ScreenerDayLosers,
		ScreenerGrowthTechStocks,
		ScreenerHighYieldBond,
		ScreenerMostActives,
		ScreenerMostShorted,
		ScreenerPortfolioAnchors,
		ScreenerSmallCapGainers,
		ScreenerSolidLargeGrowthFunds,
		ScreenerSolidMidcapGrowthFunds,
		ScreenerTopMutualFunds,
		ScreenerUndervaluedGrowth,
		ScreenerUndervaluedLargeCaps,
	}

	expected := []string{
		"aggressive_small_caps",
		"conservative_foreign_funds",
		"day_gainers",
		"day_losers",
		"growth_technology_stocks",
		"high_yield_bond",
		"most_actives",
		"most_shorted_stocks",
		"portfolio_anchors",
		"small_cap_gainers",
		"solid_large_growth_funds",
		"solid_midcap_growth_funds",
		"top_mutual_funds",
		"undervalued_growth_stocks",
		"undervalued_large_caps",
	}

	for i, scr := range screeners {
		if string(scr) != expected[i] {
			t.Errorf("screener %d: expected %s, got %s", i, expected[i], scr)
		}
	}
}

func TestScreener_ApiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/screener/predefined/saved", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"finance":{"result":[],"error":{"code":"Bad Request","description":"Invalid screener ID"}}}`))
	})

	// Handle crumb endpoints
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb123"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.Screener(context.Background(), ScreenerMostActives, nil)
	if err == nil {
		t.Fatal("expected error for API error response, got nil")
	}
}

func TestScreener_EmptyResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/screener/predefined/saved", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"finance":{"result":[],"error":null}}`))
	})

	// Handle crumb endpoints
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb123"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.Screener(context.Background(), ScreenerMostActives, nil)
	if err == nil {
		t.Fatal("expected error for empty result, got nil")
	}
}
