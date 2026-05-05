package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testSearchJSON = `{
  "count": 3,
  "quotes": [
    {
      "symbol": "AAPL",
      "isYahooFinance": true,
      "exchange": "NMS",
      "exchDisp": "Nasdaq",
      "shortname": "Apple Inc.",
      "longname": "Apple Inc.",
      "quoteType": "EQUITY",
      "typeDisp": "Equity",
      "score": 100.0,
      "sector": "Technology",
      "industry": "Consumer Electronics"
    },
    {
      "symbol": "AAPL34.SA",
      "isYahooFinance": true,
      "exchange": "SAO",
      "exchDisp": "Sao Paulo",
      "shortname": "APPLE DRN",
      "longname": "Apple Inc. - Unit",
      "quoteType": "EQUITY",
      "typeDisp": "Equity",
      "score": 50.5
    }
  ],
  "news": [
    {
      "uuid": "abc123",
      "title": "Apple Reports Record Q4 Earnings",
      "publisher": "Reuters",
      "link": "https://finance.yahoo.com/news/apple-reports-record-q4-earnings",
      "providerPublishTime": 1704267600,
      "type": "STORY"
    }
  ],
  "explains": []
}`

func TestSearch_ParsesQuotesAndNews(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameter
		q := r.URL.Query().Get("q")
		if q != "Apple" {
			t.Errorf("expected q='Apple', got %q", q)
		}
		// Verify default lang/region
		lang := r.URL.Query().Get("lang")
		if lang != "en-US" {
			t.Errorf("expected lang='en-US', got %q", lang)
		}
		region := r.URL.Query().Get("region")
		if region != "US" {
			t.Errorf("expected region='US', got %q", region)
		}
		// Verify no crumb is sent (this endpoint doesn't need one)
		if crumb := r.URL.Query().Get("crumb"); crumb != "" {
			t.Errorf("expected no crumb parameter, got %q", crumb)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testSearchJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.Search(context.Background(), "Apple", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify count
	if result.Count != 3 {
		t.Errorf("expected count 3, got %d", result.Count)
	}

	// Verify quotes
	if len(result.Quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(result.Quotes))
	}

	aapl := result.Quotes[0]
	if aapl.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", aapl.Symbol)
	}
	if !aapl.IsYahooFinance {
		t.Errorf("expected isYahooFinance true, got false")
	}
	if aapl.Exchange != "NMS" {
		t.Errorf("expected exchange NMS, got %s", aapl.Exchange)
	}
	if aapl.ExchDisp != "Nasdaq" {
		t.Errorf("expected exchDisp Nasdaq, got %s", aapl.ExchDisp)
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
	if aapl.TypeDisp != "Equity" {
		t.Errorf("expected typeDisp Equity, got %s", aapl.TypeDisp)
	}
	if aapl.Score != 100.0 {
		t.Errorf("expected score 100.0, got %f", aapl.Score)
	}
	if aapl.Sector != "Technology" {
		t.Errorf("expected sector Technology, got %s", aapl.Sector)
	}
	if aapl.Industry != "Consumer Electronics" {
		t.Errorf("expected industry Consumer Electronics, got %s", aapl.Industry)
	}

	// Second quote (no sector/industry)
	aapl34 := result.Quotes[1]
	if aapl34.Symbol != "AAPL34.SA" {
		t.Errorf("expected symbol AAPL34.SA, got %s", aapl34.Symbol)
	}
	if aapl34.Score != 50.5 {
		t.Errorf("expected score 50.5, got %f", aapl34.Score)
	}
	if aapl34.Sector != "" {
		t.Errorf("expected empty sector for second quote, got %s", aapl34.Sector)
	}

	// Verify news
	if len(result.News) != 1 {
		t.Fatalf("expected 1 news item, got %d", len(result.News))
	}
	news := result.News[0]
	if news.UUID != "abc123" {
		t.Errorf("expected uuid abc123, got %s", news.UUID)
	}
	if news.Title != "Apple Reports Record Q4 Earnings" {
		t.Errorf("expected title 'Apple Reports Record Q4 Earnings', got %s", news.Title)
	}
	if news.Publisher != "Reuters" {
		t.Errorf("expected publisher Reuters, got %s", news.Publisher)
	}
	if news.Link != "https://finance.yahoo.com/news/apple-reports-record-q4-earnings" {
		t.Errorf("unexpected link: %s", news.Link)
	}
	if news.ProviderPublishTime != 1704267600 {
		t.Errorf("expected providerPublishTime 1704267600, got %d", news.ProviderPublishTime)
	}
	if news.Type != "STORY" {
		t.Errorf("expected type STORY, got %s", news.Type)
	}
}

func TestSearch_WithOptions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		// Verify custom options
		lang := r.URL.Query().Get("lang")
		if lang != "fr-FR" {
			t.Errorf("expected lang='fr-FR', got %q", lang)
		}
		region := r.URL.Query().Get("region")
		if region != "FR" {
			t.Errorf("expected region='FR', got %q", region)
		}
		qc := r.URL.Query().Get("quotesCount")
		if qc != "5" {
			t.Errorf("expected quotesCount='5', got %q", qc)
		}
		nc := r.URL.Query().Get("newsCount")
		if nc != "3" {
			t.Errorf("expected newsCount='3', got %q", nc)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":0,"quotes":[],"news":[]}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	opts := &SearchOptions{
		Lang:        "fr-FR",
		Region:      "FR",
		QuotesCount: 5,
		NewsCount:   3,
	}
	_, err := client.Search(context.Background(), "Apple", opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSearch_EmptyQueryError(t *testing.T) {
	client := NewClient(nil)

	_, err := client.Search(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}

	var invOptsErr *InvalidOptionsError
	if !errorAs(err, &invOptsErr) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}
