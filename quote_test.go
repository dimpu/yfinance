package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testQuoteJSON = `{
  "quoteResponse": {
    "result": [
      {
        "language": "en-US",
        "region": "US",
        "quoteType": "EQUITY",
        "typeDisp": "Equity",
        "quoteSourceName": "Nasdaq Real Time Price",
        "triggerable": true,
        "customPriceAlertConfidence": "HIGH",
        "currency": "USD",
        "marketState": "REGULAR",
        "tradeable": false,
        "cryptoTradeable": false,
        "exchange": "NMS",
        "shortName": "Apple Inc.",
        "longName": "Apple Inc.",
        "messageBoardId": "finmb_249",
        "exchangeTimezoneName": "America/New_York",
        "exchangeTimezoneShortName": "EST",
        "gmtOffSetMilliseconds": -18000000,
        "market": "us_market",
        "esgPopulated": false,
        "regularMarketPrice": 178.72,
        "regularMarketChange": 2.4299927,
        "regularMarketChangePercent": 1.3791553,
        "regularMarketVolume": 54321098,
        "regularMarketOpen": 176.5,
        "regularMarketDayHigh": 179.62,
        "regularMarketDayLow": 176.15,
        "regularMarketTime": 1704267600,
        "regularMarketPreviousClose": 176.29,
        "fiftyTwoWeekLow": 124.17,
        "fiftyTwoWeekHigh": 199.62,
        "fiftyTwoWeekRange": "124.17 - 199.62",
        "fiftyTwoWeekLowChange": 54.55002,
        "fiftyTwoWeekLowChangePercent": 43.93992,
        "fiftyTwoWeekHighChange": -20.90001,
        "fiftyTwoWeekHighChangePercent": -10.470558,
        "fiftyTwoWeekChangePercent": 45.23452,
        "fiftyDayAverage": 181.5,
        "fiftyDayAverageChange": -2.78,
        "fiftyDayAverageChangePercent": -1.53168,
        "twoHundredDayAverage": 175.25,
        "twoHundredDayAverageChange": 3.47,
        "twoHundredDayAverageChangePercent": 1.98003,
        "marketCap": 2850000000000,
        "sharesOutstanding": 15550000000,
        "floatShares": 15500000000,
        "trailingPE": 29.5,
        "forwardPE": 27.8,
        "epsTrailingTwelveMonths": 6.06,
        "epsForward": 6.43,
        "trailingAnnualDividendRate": 0.96,
        "trailingAnnualDividendYield": 0.0054,
        "bookValue": 3.85,
        "priceToBook": 46.42,
        "averageDailyVolume3Month": 60000000,
        "averageDailyVolume10Day": 55000000,
        "bid": 178.71,
        "ask": 178.72,
        "bidSize": 10,
        "askSize": 9,
        "fullExchangeName": "NasdaqGS",
        "financialCurrency": "USD",
        "dividendDate": {"raw": 1703980800, "fmt": "2023-12-31"},
        "earningsTimestamp": 1704326400,
        "postMarketPrice": 178.85,
        "postMarketChange": 0.13,
        "postMarketChangePercent": 0.07,
        "preMarketPrice": 177.5,
        "preMarketChange": 1.21,
        "preMarketChangePercent": 0.68,
        "symbol": "AAPL",
        "dividendRate": 0.96,
        "dividendYield": 0.0054,
        "beta": 1.28
      },
      {
        "language": "en-US",
        "region": "US",
        "quoteType": "ETF",
        "typeDisp": "ETF",
        "quoteSourceName": "Delayed Quote",
        "triggerable": true,
        "currency": "USD",
        "marketState": "CLOSED",
        "tradeable": true,
        "cryptoTradeable": false,
        "exchange": "PCX",
        "shortName": "Vanguard S&P 500 ETF",
        "longName": "Vanguard S&P 500 ETF",
        "exchangeTimezoneName": "America/New_York",
        "exchangeTimezoneShortName": "EST",
        "gmtOffSetMilliseconds": -18000000,
        "market": "us_market",
        "esgPopulated": false,
        "regularMarketPrice": 475.25,
        "regularMarketChange": 3.5,
        "regularMarketChangePercent": 0.74,
        "regularMarketVolume": 4500000,
        "regularMarketTime": 1704267600,
        "regularMarketPreviousClose": 471.75,
        "fiftyTwoWeekLow": 370.0,
        "fiftyTwoWeekHigh": 480.0,
        "fiftyTwoWeekRange": "370.0 - 480.0",
        "marketCap": 420000000000,
        "sharesOutstanding": 880000000,
        "trailingPE": 25.5,
        "netAssets": 420000000000,
        "symbol": "VOO",
        "dividendYield": 0.015,
        "beta": 1.0
      }
    ],
    "error": null
  }
}`

func TestQuote_ParsesMultipleSymbols(t *testing.T) {
	mux := http.NewServeMux()
	// Endpoint hit by fetchCookies
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Endpoint hit by fetchCrumb
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	// Quote endpoint
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		// Verify crumb is present
		crumb := r.URL.Query().Get("crumb")
		if crumb != "testcrumb" {
			t.Errorf("expected crumb 'testcrumb', got %q", crumb)
		}
		// Verify symbols parameter
		symbols := r.URL.Query().Get("symbols")
		if symbols != "AAPL,VOO" {
			t.Errorf("expected symbols 'AAPL,VOO', got %q", symbols)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	quotes, err := client.Quote(context.Background(), []string{"AAPL", "VOO"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	// Check AAPL quote
	aapl := quotes[0]
	if aapl.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", aapl.Symbol)
	}
	if aapl.QuoteType != "EQUITY" {
		t.Errorf("expected quoteType EQUITY, got %s", aapl.QuoteType)
	}
	if aapl.MarketState != "REGULAR" {
		t.Errorf("expected marketState REGULAR, got %s", aapl.MarketState)
	}
	if aapl.RegularMarketPrice == nil || *aapl.RegularMarketPrice != 178.72 {
		t.Errorf("expected regularMarketPrice 178.72, got %v", aapl.RegularMarketPrice)
	}
	if aapl.RegularMarketChange == nil || *aapl.RegularMarketChange != 2.4299927 {
		t.Errorf("expected regularMarketChange 2.4299927, got %v", aapl.RegularMarketChange)
	}
	if aapl.RegularMarketVolume == nil || *aapl.RegularMarketVolume != 54321098 {
		t.Errorf("expected regularMarketVolume 54321098, got %v", aapl.RegularMarketVolume)
	}
	if aapl.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", aapl.Currency)
	}
	if aapl.ShortName != "Apple Inc." {
		t.Errorf("expected shortName 'Apple Inc.', got %s", aapl.ShortName)
	}

	// Check VOO quote
	voo := quotes[1]
	if voo.Symbol != "VOO" {
		t.Errorf("expected symbol VOO, got %s", voo.Symbol)
	}
	if voo.QuoteType != "ETF" {
		t.Errorf("expected quoteType ETF, got %s", voo.QuoteType)
	}
	if voo.MarketState != "CLOSED" {
		t.Errorf("expected marketState CLOSED, got %s", voo.MarketState)
	}
	if voo.RegularMarketPrice == nil || *voo.RegularMarketPrice != 475.25 {
		t.Errorf("expected regularMarketPrice 475.25, got %v", voo.RegularMarketPrice)
	}
}

func TestQuote_FiltersNoneQuoteType(t *testing.T) {
	const jsonWithNone = `{
  "quoteResponse": {
    "result": [
      {
        "language": "en-US",
        "region": "US",
        "quoteType": "EQUITY",
        "marketState": "REGULAR",
        "tradeable": true,
        "esgPopulated": false,
        "symbol": "AAPL",
        "regularMarketPrice": 150.0
      },
      {
        "quoteType": "NONE",
        "symbol": "DELISTED"
      }
    ],
    "error": null
  }
}`

	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonWithNone))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	quotes, err := client.Quote(context.Background(), []string{"AAPL", "DELISTED"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(quotes) != 1 {
		t.Fatalf("expected 1 quote (NONE filtered), got %d", len(quotes))
	}
	if quotes[0].Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", quotes[0].Symbol)
	}
}

func TestQuote_ParsesTwoNumberRange(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	quotes, err := client.Quote(context.Background(), []string{"AAPL"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	// Check FiftyTwoWeekRange parsing
	aapl := quotes[0]
	if aapl.FiftyTwoWeekRange == nil {
		t.Fatal("expected fiftyTwoWeekRange to be non-nil")
	}
	if aapl.FiftyTwoWeekRange.Low != 124.17 {
		t.Errorf("expected fiftyTwoWeekRange.Low 124.17, got %f", aapl.FiftyTwoWeekRange.Low)
	}
	if aapl.FiftyTwoWeekRange.High != 199.62 {
		t.Errorf("expected fiftyTwoWeekRange.High 199.62, got %f", aapl.FiftyTwoWeekRange.High)
	}
}

func TestQuote_ParsesYahooDate(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	quotes, err := client.Quote(context.Background(), []string{"AAPL"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	// Check DividendDate parsing
	aapl := quotes[0]
	if aapl.DividendDate == nil {
		t.Fatal("expected dividendDate to be non-nil")
	}
	// 1703980800 = 2023-12-31 00:00:00 UTC
	expectedUnix := int64(1703980800)
	if aapl.DividendDate.Unix() != expectedUnix {
		t.Errorf("expected dividendDate unix %d, got %d", expectedUnix, aapl.DividendDate.Unix())
	}
}

func TestQuote_WithFieldsOption(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		// Verify fields parameter
		fields := r.URL.Query().Get("fields")
		if fields != "symbol,regularMarketPrice,currency" {
			t.Errorf("expected fields 'symbol,regularMarketPrice,currency', got %q", fields)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	opts := &QuoteOptions{
		Fields: []string{"symbol", "regularMarketPrice", "currency"},
	}
	_, err := client.Quote(context.Background(), []string{"AAPL"}, opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestQuote_EmptySymbolsError(t *testing.T) {
	client := NewClient(&Config{})

	_, err := client.Quote(context.Background(), []string{}, nil)
	if err == nil {
		t.Fatal("expected error for empty symbols, got nil")
	}

	var invOptsErr *InvalidOptionsError
	if !errorAs(err, &invOptsErr) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

// Helper for error type assertion (Go 1.13+ compatible)
func errorAs(err error, target interface{}) bool {
	return errorAsImpl(err, target)
}

func errorAsImpl(err error, target interface{}) bool {
	// Use errors.As if available (Go 1.13+)
	// For simplicity, just check the type directly
	if e, ok := err.(*InvalidOptionsError); ok {
		*(target.(**InvalidOptionsError)) = e
		return true
	}
	return false
}
