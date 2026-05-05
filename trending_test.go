package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testTrendingJSON = `{
  "finance": {
    "result": [
      {
        "count": 5,
        "quotes": [
          {"symbol": "AAPL"},
          {"symbol": "MSFT"},
          {"symbol": "GOOGL"},
          {"symbol": "AMZN"},
          {"symbol": "TSLA"}
        ],
        "jobTimestamp": 1714800000,
        "startInterval": 1714796400
      }
    ],
    "error": null
  }
}`

func TestTrending_ParsesSymbols(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/trending/US", func(w http.ResponseWriter, r *http.Request) {
		// Verify no crumb is sent (this endpoint doesn't need one)
		if crumb := r.URL.Query().Get("crumb"); crumb != "" {
			t.Errorf("expected no crumb parameter, got %q", crumb)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTrendingJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.TrendingSymbols(context.Background(), "US", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Count != 5 {
		t.Errorf("expected count 5, got %d", result.Count)
	}
	if len(result.Quotes) != 5 {
		t.Fatalf("expected 5 quotes, got %d", len(result.Quotes))
	}
	if result.Quotes[0].Symbol != "AAPL" {
		t.Errorf("expected first symbol AAPL, got %s", result.Quotes[0].Symbol)
	}
	if result.Quotes[1].Symbol != "MSFT" {
		t.Errorf("expected second symbol MSFT, got %s", result.Quotes[1].Symbol)
	}
	if result.Quotes[4].Symbol != "TSLA" {
		t.Errorf("expected fifth symbol TSLA, got %s", result.Quotes[4].Symbol)
	}
	if result.JobTimestamp != 1714800000 {
		t.Errorf("expected jobTimestamp 1714800000, got %d", result.JobTimestamp)
	}
	if result.StartInterval != 1714796400 {
		t.Errorf("expected startInterval 1714796400, got %d", result.StartInterval)
	}
}

func TestTrending_EmptyRegionError(t *testing.T) {
	client := NewClient(nil)

	_, err := client.TrendingSymbols(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty region, got nil")
	}

	var invOptsErr *InvalidOptionsError
	if !errorAs(err, &invOptsErr) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestTrending_WithOpts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/trending/GB", func(w http.ResponseWriter, r *http.Request) {
		// Verify query params from options
		if lang := r.URL.Query().Get("lang"); lang != "en-GB" {
			t.Errorf("expected lang en-GB, got %q", lang)
		}
		if count := r.URL.Query().Get("count"); count != "3" {
			t.Errorf("expected count 3, got %q", count)
		}

		const gbTrendingJSON = `{
		  "finance": {
		    "result": [
		      {
		        "count": 3,
		        "quotes": [
		          {"symbol": "BARC.L"},
		          {"symbol": "HSBA.L"},
		          {"symbol": "BP.L"}
		        ],
		        "jobTimestamp": 1714800000,
		        "startInterval": 1714796400
		      }
		    ],
		    "error": null
		  }
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(gbTrendingJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.TrendingSymbols(context.Background(), "GB", &TrendingOptions{
		Lang:  "en-GB",
		Count: 3,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Count != 3 {
		t.Errorf("expected count 3, got %d", result.Count)
	}
	if len(result.Quotes) != 3 {
		t.Fatalf("expected 3 quotes, got %d", len(result.Quotes))
	}
	if result.Quotes[0].Symbol != "BARC.L" {
		t.Errorf("expected first symbol BARC.L, got %s", result.Quotes[0].Symbol)
	}
}

func TestTrending_ApiError(t *testing.T) {
	const errorJSON = `{
	  "finance": {
	    "result": null,
	    "error": {
	      "code": "Bad Request",
	      "description": "Invalid region"
	    }
	  }
	}`

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/trending/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(errorJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.TrendingSymbols(context.Background(), "INVALID", nil)
	if err == nil {
		t.Fatal("expected error for API error response, got nil")
	}
}

func TestTrending_EmptyResult(t *testing.T) {
	const emptyResultJSON = `{
	  "finance": {
	    "result": [],
	    "error": null
	  }
	}`

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/trending/XX", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(emptyResultJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.TrendingSymbols(context.Background(), "XX", nil)
	if err == nil {
		t.Fatal("expected error for empty results, got nil")
	}
}
