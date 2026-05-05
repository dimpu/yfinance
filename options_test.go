package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testOptionsJSON = `{
  "optionChain": {
    "result": [
      {
        "underlyingSymbol": "AAPL",
        "expirationDates": [1703980800, 1704585600, 1705190400],
        "strikes": [150.0, 155.0, 160.0, 165.0, 170.0, 175.0, 180.0],
        "hasMiniOptions": false,
        "quote": {
          "language": "en-US",
          "region": "US",
          "quoteType": "EQUITY",
          "currency": "USD",
          "marketState": "REGULAR",
          "tradeable": false,
          "esgPopulated": false,
          "symbol": "AAPL",
          "regularMarketPrice": 178.72
        },
        "options": [
          {
            "expirationDate": 1703980800,
            "hasMiniOptions": false,
            "calls": [
              {
                "contractSymbol": "AAPL231230C00150000",
                "strike": 150.0,
                "currency": "USD",
                "lastPrice": 29.5,
                "change": 1.2,
                "percentChange": 4.24,
                "volume": 10,
                "openInterest": 25,
                "bid": 29.2,
                "ask": 29.8,
                "contractSize": "REGULAR",
                "expiration": 1703980800,
                "lastTradeDate": 1703755200,
                "impliedVolatility": 0.35,
                "inTheMoney": true
              },
              {
                "contractSymbol": "AAPL231230C00180000",
                "strike": 180.0,
                "currency": "USD",
                "lastPrice": 2.15,
                "change": -0.35,
                "percentChange": -14.0,
                "volume": 5000,
                "openInterest": 12000,
                "bid": 2.1,
                "ask": 2.2,
                "contractSize": "REGULAR",
                "expiration": 1703980800,
                "lastTradeDate": 1703755200,
                "impliedVolatility": 0.42,
                "inTheMoney": false
              }
            ],
            "puts": [
              {
                "contractSymbol": "AAPL231230P00150000",
                "strike": 150.0,
                "currency": "USD",
                "lastPrice": 0.45,
                "change": -0.1,
                "percentChange": -18.18,
                "volume": 50,
                "openInterest": 200,
                "bid": 0.4,
                "ask": 0.5,
                "contractSize": "REGULAR",
                "expiration": 1703980800,
                "lastTradeDate": 1703755200,
                "impliedVolatility": 0.38,
                "inTheMoney": false
              },
              {
                "contractSymbol": "AAPL231230P00180000",
                "strike": 180.0,
                "currency": "USD",
                "lastPrice": 3.8,
                "change": 0.55,
                "percentChange": 16.92,
                "volume": 3000,
                "openInterest": 8500,
                "bid": 3.75,
                "ask": 3.85,
                "contractSize": "REGULAR",
                "expiration": 1703980800,
                "lastTradeDate": 1703755200,
                "impliedVolatility": 0.40,
                "inTheMoney": true
              }
            ]
          }
        ]
      }
    ],
    "error": null
  }
}`

func TestOptions_ParsesStrikes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/options/AAPL", func(w http.ResponseWriter, r *http.Request) {
		// Verify crumb is present
		crumb := r.URL.Query().Get("crumb")
		if crumb != "testcrumb" {
			t.Errorf("expected crumb 'testcrumb', got %q", crumb)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testOptionsJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.Options(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify strikes
	expectedStrikes := []float64{150.0, 155.0, 160.0, 165.0, 170.0, 175.0, 180.0}
	if len(result.Strikes) != len(expectedStrikes) {
		t.Fatalf("expected %d strikes, got %d", len(expectedStrikes), len(result.Strikes))
	}
	for i, s := range result.Strikes {
		if s != expectedStrikes[i] {
			t.Errorf("strike[%d]: expected %f, got %f", i, expectedStrikes[i], s)
		}
	}

	// Verify underlying symbol
	if result.UnderlyingSymbol != "AAPL" {
		t.Errorf("expected underlyingSymbol AAPL, got %s", result.UnderlyingSymbol)
	}

	// Verify expiration dates
	if len(result.ExpirationDates) != 3 {
		t.Errorf("expected 3 expiration dates, got %d", len(result.ExpirationDates))
	}

	// Verify hasMiniOptions
	if result.HasMiniOptions {
		t.Errorf("expected hasMiniOptions false, got true")
	}

	// Verify quote
	if result.Quote.Symbol != "AAPL" {
		t.Errorf("expected quote symbol AAPL, got %s", result.Quote.Symbol)
	}
}

func TestOptions_ParsesCalls(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/options/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testOptionsJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.Options(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Options) != 1 {
		t.Fatalf("expected 1 option expiration, got %d", len(result.Options))
	}

	opt := result.Options[0]
	if len(opt.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(opt.Calls))
	}

	// First call (in the money)
	call0 := opt.Calls[0]
	if call0.ContractSymbol != "AAPL231230C00150000" {
		t.Errorf("expected contractSymbol AAPL231230C00150000, got %s", call0.ContractSymbol)
	}
	if call0.Strike != 150.0 {
		t.Errorf("expected strike 150.0, got %f", call0.Strike)
	}
	if call0.LastPrice != 29.5 {
		t.Errorf("expected lastPrice 29.5, got %f", call0.LastPrice)
	}
	if call0.ImpliedVolatility != 0.35 {
		t.Errorf("expected impliedVolatility 0.35, got %f", call0.ImpliedVolatility)
	}
	if !call0.InTheMoney {
		t.Errorf("expected inTheMoney true, got false")
	}
	if call0.Volume == nil || *call0.Volume != 10 {
		t.Errorf("expected volume 10, got %v", call0.Volume)
	}
	if call0.OpenInterest == nil || *call0.OpenInterest != 25 {
		t.Errorf("expected openInterest 25, got %v", call0.OpenInterest)
	}
	if call0.Bid == nil || *call0.Bid != 29.2 {
		t.Errorf("expected bid 29.2, got %v", call0.Bid)
	}
	if call0.Ask == nil || *call0.Ask != 29.8 {
		t.Errorf("expected ask 29.8, got %v", call0.Ask)
	}
	if call0.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", call0.Currency)
	}
	if call0.ContractSize != "REGULAR" {
		t.Errorf("expected contractSize REGULAR, got %s", call0.ContractSize)
	}
	if call0.Expiration != 1703980800 {
		t.Errorf("expected expiration 1703980800, got %d", call0.Expiration)
	}

	// Second call (out of the money)
	call1 := opt.Calls[1]
	if call1.Strike != 180.0 {
		t.Errorf("expected strike 180.0, got %f", call1.Strike)
	}
	if call1.InTheMoney {
		t.Errorf("expected inTheMoney false, got true")
	}
	if call1.PercentChange == nil || *call1.PercentChange != -14.0 {
		t.Errorf("expected percentChange -14.0, got %v", call1.PercentChange)
	}
}

func TestOptions_ParsesPuts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/options/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testOptionsJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.Options(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	opt := result.Options[0]
	if len(opt.Puts) != 2 {
		t.Fatalf("expected 2 puts, got %d", len(opt.Puts))
	}

	// First put (out of the money)
	put0 := opt.Puts[0]
	if put0.ContractSymbol != "AAPL231230P00150000" {
		t.Errorf("expected contractSymbol AAPL231230P00150000, got %s", put0.ContractSymbol)
	}
	if put0.Strike != 150.0 {
		t.Errorf("expected strike 150.0, got %f", put0.Strike)
	}
	if put0.LastPrice != 0.45 {
		t.Errorf("expected lastPrice 0.45, got %f", put0.LastPrice)
	}
	if put0.InTheMoney {
		t.Errorf("expected inTheMoney false, got true")
	}

	// Second put (in the money)
	put1 := opt.Puts[1]
	if put1.ContractSymbol != "AAPL231230P00180000" {
		t.Errorf("expected contractSymbol AAPL231230P00180000, got %s", put1.ContractSymbol)
	}
	if put1.Strike != 180.0 {
		t.Errorf("expected strike 180.0, got %f", put1.Strike)
	}
	if !put1.InTheMoney {
		t.Errorf("expected inTheMoney true, got false")
	}
	if put1.PercentChange == nil || *put1.PercentChange != 16.92 {
		t.Errorf("expected percentChange 16.92, got %v", put1.PercentChange)
	}
}

func TestOptions_WithDateOption(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/options/AAPL", func(w http.ResponseWriter, r *http.Request) {
		// Verify date query param is set
		dateParam := r.URL.Query().Get("date")
		if dateParam != "1703980800" {
			t.Errorf("expected date '1703980800', got %q", dateParam)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testOptionsJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	date := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	_, err := client.Options(context.Background(), "AAPL", &OptionsOptions{
		Date: &date,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOptions_WithLangRegion(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/options/AAPL", func(w http.ResponseWriter, r *http.Request) {
		lang := r.URL.Query().Get("lang")
		if lang != "en-US" {
			t.Errorf("expected lang 'en-US', got %q", lang)
		}
		region := r.URL.Query().Get("region")
		if region != "US" {
			t.Errorf("expected region 'US', got %q", region)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testOptionsJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.Options(context.Background(), "AAPL", &OptionsOptions{
		Lang:   "en-US",
		Region: "US",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
