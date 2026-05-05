package yahoofinance

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testInsightsJSON = `{
  "finance": {
    "result": {
      "symbol": "AAPL",
      "instrumentInfo": {
        "keyTechnicals": {
          "provider": "Markit",
          "support": 170.5,
          "resistance": 195.0,
          "stopLoss": 165.0
        },
        "technicalEvents": {
          "provider": "Markit",
          "sector": "Technology",
          "shortTermOutlook": {
            "stateDescription": "Bullish",
            "direction": "Bullish",
            "score": 8.5,
            "scoreDescription": "Strong",
            "sectorDirection": "Bullish",
            "sectorScore": 7.0,
            "indexDirection": "Bullish",
            "indexScore": 6.5
          },
          "intermediateTermOutlook": {
            "stateDescription": "Neutral",
            "direction": "Neutral",
            "score": 5.0
          },
          "longTermOutlook": {
            "stateDescription": "Bullish",
            "direction": "Bullish",
            "score": 7.5
          }
        },
        "valuation": {
          "color": "green",
          "description": "Undervalued",
          "discount": 0.15,
          "provider": "Markit",
          "relativeValue": 1.15
        }
      },
      "companySnapshot": {
        "sectorInfo": "Technology",
        "company": {
          "innovativeness": 9.0,
          "hiring": 7.5,
          "sustainability": 8.0,
          "insiderSentiments": 6.5,
          "earningsReports": 8.5,
          "dividends": 7.0
        },
        "sector": {
          "innovativeness": 7.0,
          "hiring": 6.5,
          "sustainability": 7.5,
          "insiderSentiments": 5.5,
          "earningsReports": 6.0,
          "dividends": 6.5
        }
      },
      "recommendation": {
        "targetPrice": 200.0,
        "provider": "Markit",
        "rating": "BUY"
      },
      "events": [],
      "reports": [
        {
          "id": "report-123",
          "title": "Apple Q4 Earnings Analysis",
          "provider": "Yahoo Finance",
          "publishedAt": 1704267600
        }
      ],
      "sigDevs": [
        {
          "headline": "Apple announces new product line",
          "date": 1704000000
        }
      ]
    }
  }
}`

func TestInsights_ParsesAllFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/insights/v2/finance/insights", func(w http.ResponseWriter, r *http.Request) {
		// Verify symbol parameter
		symbol := r.URL.Query().Get("symbol")
		if symbol != "AAPL" {
			t.Errorf("expected symbol='AAPL', got %q", symbol)
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
		// Verify no crumb is sent
		if crumb := r.URL.Query().Get("crumb"); crumb != "" {
			t.Errorf("expected no crumb parameter, got %q", crumb)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testInsightsJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.Insights(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify symbol
	if result.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", result.Symbol)
	}

	// Verify instrumentInfo
	if result.InstrumentInfo == nil {
		t.Fatal("expected instrumentInfo, got nil")
	}

	// Verify keyTechnicals
	kt := result.InstrumentInfo.KeyTechnicals
	if kt == nil {
		t.Fatal("expected keyTechnicals, got nil")
	}
	if kt.Provider != "Markit" {
		t.Errorf("expected provider Markit, got %s", kt.Provider)
	}
	if kt.Support == nil || *kt.Support != 170.5 {
		t.Errorf("expected support 170.5, got %v", kt.Support)
	}
	if kt.Resistance == nil || *kt.Resistance != 195.0 {
		t.Errorf("expected resistance 195.0, got %v", kt.Resistance)
	}
	if kt.StopLoss == nil || *kt.StopLoss != 165.0 {
		t.Errorf("expected stopLoss 165.0, got %v", kt.StopLoss)
	}

	// Verify technicalEvents
	te := result.InstrumentInfo.TechnicalEvents
	if te == nil {
		t.Fatal("expected technicalEvents, got nil")
	}
	if te.Provider != "Markit" {
		t.Errorf("expected provider Markit, got %s", te.Provider)
	}
	if te.Sector != "Technology" {
		t.Errorf("expected sector Technology, got %s", te.Sector)
	}

	// Verify shortTermOutlook
	sto := te.ShortTermOutlook
	if sto == nil {
		t.Fatal("expected shortTermOutlook, got nil")
	}
	if sto.StateDescription != "Bullish" {
		t.Errorf("expected stateDescription Bullish, got %s", sto.StateDescription)
	}
	if sto.Direction != "Bullish" {
		t.Errorf("expected direction Bullish, got %s", sto.Direction)
	}
	if sto.Score == nil || *sto.Score != 8.5 {
		t.Errorf("expected score 8.5, got %v", sto.Score)
	}

	// Verify valuation
	v := result.InstrumentInfo.Valuation
	if v == nil {
		t.Fatal("expected valuation, got nil")
	}
	if v.Provider != "Markit" {
		t.Errorf("expected provider Markit, got %s", v.Provider)
	}
	if v.Color != "green" {
		t.Errorf("expected color green, got %s", v.Color)
	}
	if v.Description != "Undervalued" {
		t.Errorf("expected description Undervalued, got %s", v.Description)
	}
	if v.Discount == nil || *v.Discount != 0.15 {
		t.Errorf("expected discount 0.15, got %v", v.Discount)
	}
	if v.RelativeValue == nil || *v.RelativeValue != 1.15 {
		t.Errorf("expected relativeValue 1.15, got %v", v.RelativeValue)
	}

	// Verify companySnapshot
	cs := result.CompanySnapshot
	if cs == nil {
		t.Fatal("expected companySnapshot, got nil")
	}
	if cs.SectorInfo != "Technology" {
		t.Errorf("expected sectorInfo Technology, got %s", cs.SectorInfo)
	}

	// Verify company
	company := cs.Company
	if company == nil {
		t.Fatal("expected company, got nil")
	}
	if company.Innovativeness == nil || *company.Innovativeness != 9.0 {
		t.Errorf("expected innovativeness 9.0, got %v", company.Innovativeness)
	}

	// Verify sector info
	sector := cs.Sector
	if sector == nil {
		t.Fatal("expected sector, got nil")
	}
	if sector.Innovativeness != 7.0 {
		t.Errorf("expected sector innovativeness 7.0, got %f", sector.Innovativeness)
	}

	// Verify recommendation
	rec := result.Recommendation
	if rec == nil {
		t.Fatal("expected recommendation, got nil")
	}
	if rec.Provider != "Markit" {
		t.Errorf("expected provider Markit, got %s", rec.Provider)
	}
	if rec.Rating != "BUY" {
		t.Errorf("expected rating BUY, got %s", rec.Rating)
	}
	if rec.TargetPrice == nil || *rec.TargetPrice != 200.0 {
		t.Errorf("expected targetPrice 200.0, got %v", rec.TargetPrice)
	}

	// Verify reports
	if len(result.Reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(result.Reports))
	}
	report := result.Reports[0]
	if report.ID != "report-123" {
		t.Errorf("expected id report-123, got %s", report.ID)
	}
	if report.Title != "Apple Q4 Earnings Analysis" {
		t.Errorf("expected title 'Apple Q4 Earnings Analysis', got %s", report.Title)
	}
	if report.Provider != "Yahoo Finance" {
		t.Errorf("expected provider 'Yahoo Finance', got %s", report.Provider)
	}
	if report.PublishedAt != 1704267600 {
		t.Errorf("expected publishedAt 1704267600, got %d", report.PublishedAt)
	}

	// Verify sigDevs
	if len(result.SigDevs) != 1 {
		t.Fatalf("expected 1 sigDev, got %d", len(result.SigDevs))
	}
	sigDev := result.SigDevs[0]
	if sigDev.Headline != "Apple announces new product line" {
		t.Errorf("expected headline 'Apple announces new product line', got %s", sigDev.Headline)
	}
	if sigDev.Date != 1704000000 {
		t.Errorf("expected date 1704000000, got %d", sigDev.Date)
	}
}

func TestInsights_WithOptions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/insights/v2/finance/insights", func(w http.ResponseWriter, r *http.Request) {
		// Verify custom options
		lang := r.URL.Query().Get("lang")
		if lang != "fr-FR" {
			t.Errorf("expected lang='fr-FR', got %q", lang)
		}
		region := r.URL.Query().Get("region")
		if region != "FR" {
			t.Errorf("expected region='FR', got %q", region)
		}
		rc := r.URL.Query().Get("reportsCount")
		if rc != "5" {
			t.Errorf("expected reportsCount='5', got %q", rc)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"finance":{"result":{"symbol":"AAPL"}}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	opts := &InsightsOptions{
		Lang:         "fr-FR",
		Region:       "FR",
		ReportsCount: 5,
	}
	result, err := client.Insights(context.Background(), "AAPL", opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", result.Symbol)
	}
}

func TestInsights_EmptySymbolError(t *testing.T) {
	client := NewClient(nil)

	_, err := client.Insights(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	var invOptsErr *InvalidOptionsError
	if !errors.As(err, &invOptsErr) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestInsights_ApiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/insights/v2/finance/insights", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"finance":{"error":{"code":"Not Found","description":"Symbol not found"}}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.Insights(context.Background(), "INVALID", nil)
	if err == nil {
		t.Fatal("expected error for API error, got nil")
	}

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("expected HTTPError, got %T: %v", err, err)
	}
	if httpErr.Body != "Symbol not found" {
		t.Errorf("expected body 'Symbol not found', got %q", httpErr.Body)
	}
}

func TestInsights_EmptyResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/insights/v2/finance/insights", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"finance":{}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.Insights(context.Background(), "AAPL", nil)
	if err == nil {
		t.Fatal("expected error for empty result, got nil")
	}

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("expected HTTPError, got %T: %v", err, err)
	}
}
