package quotesummary

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	internalErrors "github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/types"
)

const quoteSummaryTestJSON = `{
  "quoteSummary": {
    "result": [{
      "assetProfile": {
        "maxAge": 86400,
        "address1": "One Apple Park Way",
        "city": "Cupertino",
        "state": "CA",
        "zip": "95014",
        "country": "United States",
        "phone": "408-996-1010",
        "website": "https://www.apple.com",
        "industry": "Consumer Electronics",
        "sector": "Technology",
        "fullTimeEmployees": 164000,
        "companyOfficers": []
      },
      "financialData": {
        "maxAge": 86400,
        "currentPrice": 178.72,
        "targetHighPrice": 220.0,
        "targetLowPrice": 140.0,
        "targetMeanPrice": 195.5,
        "recommendationKey": "buy",
        "financialCurrency": "USD"
      },
      "defaultKeyStatistics": {
        "maxAge": 86400,
        "enterpriseValue": 2780000000000,
        "forwardPE": 28.5,
        "beta": 1.29
      },
      "price": {
        "exchange": "NMS",
        "maxAge": 86400,
        "regularMarketPrice": 178.72,
        "regularMarketVolume": 54321098,
        "symbol": "AAPL",
        "shortName": "Apple Inc.",
        "currency": "USD"
      },
      "summaryDetail": {
        "maxAge": 86400,
        "previousClose": 177.5,
        "regularMarketVolume": 54321098,
        "marketCap": 2780000000000,
        "fiftyTwoWeekLow": 124.17,
        "fiftyTwoWeekHigh": 199.62,
        "currency": "USD"
      },
      "recommendationTrend": {
        "trend": [
          {"period": "0m", "strongBuy": 15, "buy": 20, "hold": 5, "sell": 2, "strongSell": 1}
        ],
        "maxAge": 86400
      },
      "earnings": {
        "maxAge": 86400,
        "financialCurrency": "USD",
        "earningsChart": {
          "quarterly": [
            {"date": "3Q2023", "actual": 1.46, "estimate": 1.39}
          ]
        },
        "financialsChart": {
          "yearly": [
            {"date": "2023", "revenue": 383285000000, "earnings": 96995000000}
          ]
        }
      }
    }],
    "error": null
  }
}`

// redirectTransport routes all requests to the test server.
type redirectTransport struct {
	server *httptest.Server
	base   http.RoundTripper
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := *req.URL
	newURL.Scheme = "http"
	newURL.Host = t.server.Listener.Addr().String()
	req.URL = &newURL
	return t.base.RoundTrip(req)
}

func newTestFetcher(srv *httptest.Server) *fetch.Fetcher {
	jar, _ := cookiejar.New(nil)
	return fetch.NewFetcher(fetch.Config{
		QueryHost:  srv.URL,
		HTTPClient: &http.Client{Jar: jar, Transport: &redirectTransport{server: srv, base: http.DefaultTransport}},
		Concurrency: 4,
		Logger:      types.NewDefaultLogger(),
	})
}

func newQuoteSummaryServer(body string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	})
	return httptest.NewServer(mux)
}

func TestGetReturnsQuoteSummaryResult(t *testing.T) {
	srv := newQuoteSummaryServer(quoteSummaryTestJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	result, err := svc.Get(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AssetProfile
	if result.AssetProfile == nil {
		t.Fatal("AssetProfile is nil")
	}
	if result.AssetProfile.City != "Cupertino" {
		t.Errorf("AssetProfile.City = %q, want %q", result.AssetProfile.City, "Cupertino")
	}
	if result.AssetProfile.Sector != "Technology" {
		t.Errorf("AssetProfile.Sector = %q, want %q", result.AssetProfile.Sector, "Technology")
	}

	// FinancialData
	if result.FinancialData == nil {
		t.Fatal("FinancialData is nil")
	}
	if result.FinancialData.CurrentPrice == nil || *result.FinancialData.CurrentPrice != 178.72 {
		t.Errorf("FinancialData.CurrentPrice = %v, want 178.72", result.FinancialData.CurrentPrice)
	}
	if result.FinancialData.RecommendationKey != "buy" {
		t.Errorf("FinancialData.RecommendationKey = %q, want %q", result.FinancialData.RecommendationKey, "buy")
	}

	// DefaultKeyStatistics
	if result.DefaultKeyStatistics == nil {
		t.Fatal("DefaultKeyStatistics is nil")
	}
	if result.DefaultKeyStatistics.ForwardPE == nil || *result.DefaultKeyStatistics.ForwardPE != 28.5 {
		t.Errorf("DefaultKeyStatistics.ForwardPE = %v, want 28.5", result.DefaultKeyStatistics.ForwardPE)
	}

	// Price
	if result.Price == nil {
		t.Fatal("Price is nil")
	}
	if result.Price.Symbol != "AAPL" {
		t.Errorf("Price.Symbol = %q, want %q", result.Price.Symbol, "AAPL")
	}

	// SummaryDetail
	if result.SummaryDetail == nil {
		t.Fatal("SummaryDetail is nil")
	}
	if result.SummaryDetail.FiftyTwoWeekHigh == nil || *result.SummaryDetail.FiftyTwoWeekHigh != 199.62 {
		t.Errorf("SummaryDetail.FiftyTwoWeekHigh = %v, want 199.62", result.SummaryDetail.FiftyTwoWeekHigh)
	}

	// RecommendationTrend
	if result.RecommendationTrend == nil {
		t.Fatal("RecommendationTrend is nil")
	}
	if len(result.RecommendationTrend.Trend) != 1 {
		t.Fatalf("len(Trend) = %d, want 1", len(result.RecommendationTrend.Trend))
	}
	if result.RecommendationTrend.Trend[0].StrongBuy != 15 {
		t.Errorf("Trend[0].StrongBuy = %d, want 15", result.RecommendationTrend.Trend[0].StrongBuy)
	}

	// Earnings
	if result.Earnings == nil {
		t.Fatal("Earnings is nil")
	}
	if result.Earnings.EarningsChart == nil {
		t.Fatal("Earnings.EarningsChart is nil")
	}
	if len(result.Earnings.EarningsChart.Quarterly) != 1 {
		t.Fatalf("len(Quarterly) = %d, want 1", len(result.Earnings.EarningsChart.Quarterly))
	}
	q := result.Earnings.EarningsChart.Quarterly[0]
	if q.Actual == nil || *q.Actual != 1.46 {
		t.Errorf("Quarterly[0].Actual = %v, want 1.46", q.Actual)
	}
}

func TestGetEmptySymbolReturnsError(t *testing.T) {
	srv := newQuoteSummaryServer(quoteSummaryTestJSON)
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

func TestGetAPIErrorReturnsError(t *testing.T) {
	errorJSON := `{"quoteSummary":{"result":null,"error":{"code":"Not Found","description":"No data found for AAPL"}}}`
	srv := newQuoteSummaryServer(errorJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetEmptyResultReturnsError(t *testing.T) {
	emptyJSON := `{"quoteSummary":{"result":[],"error":null}}`
	srv := newQuoteSummaryServer(emptyJSON)
	defer srv.Close()

	svc := NewService(newTestFetcher(srv))
	_, err := svc.Get(context.Background(), "AAPL", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveModulesDefault(t *testing.T) {
	modules := resolveModules(nil)
	if len(modules) != 2 || modules[0] != "price" || modules[1] != "summaryDetail" {
		t.Errorf("resolveModules(nil) = %v, want [price summaryDetail]", modules)
	}
}

func TestResolveModulesEmptyModules(t *testing.T) {
	modules := resolveModules(&Options{})
	if len(modules) != 2 || modules[0] != "price" || modules[1] != "summaryDetail" {
		t.Errorf("resolveModules(&Options{}) = %v, want [price summaryDetail]", modules)
	}
}

func TestResolveModulesAll(t *testing.T) {
	modules := resolveModules(&Options{Modules: []string{"all"}})
	if len(modules) != len(allModules) {
		t.Errorf("resolveModules(all) returned %d modules, want %d", len(modules), len(allModules))
	}
}

func TestResolveModulesSpecific(t *testing.T) {
	modules := resolveModules(&Options{Modules: []string{"price", "financialData"}})
	if len(modules) != 2 {
		t.Errorf("resolveModules([price, financialData]) returned %d modules, want 2", len(modules))
	}
}
