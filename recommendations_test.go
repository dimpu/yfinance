package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testRecommendationsJSON = `{
  "finance": {
    "result": [
      {
        "recommendedSymbols": [
          {"score": 2.5, "symbol": "MSFT"},
          {"score": 1.8, "symbol": "GOOGL"},
          {"score": 1.2, "symbol": "AMZN"}
        ],
        "symbol": "AAPL"
      },
      {
        "recommendedSymbols": [
          {"score": 3.0, "symbol": "AAPL"},
          {"score": 2.1, "symbol": "META"}
        ],
        "symbol": "MSFT"
      }
    ],
    "error": null
  }
}`

func TestRecommendations_ParsesResults(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v6/finance/recommendationsbysymbol/", func(w http.ResponseWriter, r *http.Request) {
		// Verify the symbols are in the path
		expectedPath := "/v6/finance/recommendationsbysymbol/AAPL,MSFT"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, r.URL.Path)
		}
		// Verify no crumb is sent (this endpoint doesn't need one)
		if crumb := r.URL.Query().Get("crumb"); crumb != "" {
			t.Errorf("expected no crumb parameter, got %q", crumb)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRecommendationsJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	results, err := client.RecommendationsBySymbol(context.Background(), []string{"AAPL", "MSFT"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify AAPL recommendations
	aapl := results[0]
	if aapl.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", aapl.Symbol)
	}
	if len(aapl.RecommendedSymbols) != 3 {
		t.Fatalf("expected 3 recommended symbols for AAPL, got %d", len(aapl.RecommendedSymbols))
	}
	if aapl.RecommendedSymbols[0].Symbol != "MSFT" {
		t.Errorf("expected first recommended symbol MSFT, got %s", aapl.RecommendedSymbols[0].Symbol)
	}
	if aapl.RecommendedSymbols[0].Score != 2.5 {
		t.Errorf("expected first recommended score 2.5, got %f", aapl.RecommendedSymbols[0].Score)
	}
	if aapl.RecommendedSymbols[1].Symbol != "GOOGL" {
		t.Errorf("expected second recommended symbol GOOGL, got %s", aapl.RecommendedSymbols[1].Symbol)
	}
	if aapl.RecommendedSymbols[2].Symbol != "AMZN" {
		t.Errorf("expected third recommended symbol AMZN, got %s", aapl.RecommendedSymbols[2].Symbol)
	}

	// Verify MSFT recommendations
	msft := results[1]
	if msft.Symbol != "MSFT" {
		t.Errorf("expected symbol MSFT, got %s", msft.Symbol)
	}
	if len(msft.RecommendedSymbols) != 2 {
		t.Fatalf("expected 2 recommended symbols for MSFT, got %d", len(msft.RecommendedSymbols))
	}
	if msft.RecommendedSymbols[0].Symbol != "AAPL" {
		t.Errorf("expected first recommended symbol AAPL, got %s", msft.RecommendedSymbols[0].Symbol)
	}
	if msft.RecommendedSymbols[0].Score != 3.0 {
		t.Errorf("expected first recommended score 3.0, got %f", msft.RecommendedSymbols[0].Score)
	}
}

func TestRecommendations_EmptySymbolsError(t *testing.T) {
	client := NewClient(nil)

	_, err := client.RecommendationsBySymbol(context.Background(), []string{}, nil)
	if err == nil {
		t.Fatal("expected error for empty symbols, got nil")
	}

	var invOptsErr *InvalidOptionsError
	if !errorAs(err, &invOptsErr) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestRecommendations_SingleSymbol(t *testing.T) {
	const singleSymbolJSON = `{
  "finance": {
    "result": [
      {
        "recommendedSymbols": [
          {"score": 5.0, "symbol": "TSLA"}
        ],
        "symbol": "AAPL"
      }
    ],
    "error": null
  }
}`

	mux := http.NewServeMux()
	mux.HandleFunc("/v6/finance/recommendationsbysymbol/", func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v6/finance/recommendationsbysymbol/AAPL"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(singleSymbolJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	results, err := client.RecommendationsBySymbol(context.Background(), []string{"AAPL"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", results[0].Symbol)
	}
	if len(results[0].RecommendedSymbols) != 1 {
		t.Fatalf("expected 1 recommended symbol, got %d", len(results[0].RecommendedSymbols))
	}
	if results[0].RecommendedSymbols[0].Symbol != "TSLA" {
		t.Errorf("expected recommended symbol TSLA, got %s", results[0].RecommendedSymbols[0].Symbol)
	}
	if results[0].RecommendedSymbols[0].Score != 5.0 {
		t.Errorf("expected score 5.0, got %f", results[0].RecommendedSymbols[0].Score)
	}
}

func TestRecommendations_ApiError(t *testing.T) {
	const errorJSON = `{
  "finance": {
    "result": null,
    "error": {
      "code": "Bad Request",
      "description": "Invalid symbols"
    }
  }
}`

	mux := http.NewServeMux()
	mux.HandleFunc("/v6/finance/recommendationsbysymbol/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(errorJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.RecommendationsBySymbol(context.Background(), []string{"INVALID"}, nil)
	if err == nil {
		t.Fatal("expected error for API error response, got nil")
	}
}
