package yahoofinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testQuoteSummaryJSON = `{
  "quoteSummary": {
    "result": [
      {
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
          "longBusinessSummary": "Apple Inc. designs, manufactures, and markets smartphones, personal computers, tablets, wearables, and accessories worldwide.",
          "fullTimeEmployees": 164000,
          "companyOfficers": [
            {
              "maxAge": 86400,
              "name": "Mr. Timothy D. Cook",
              "title": "CEO & Director",
              "yearBorn": 1960,
              "totalPay": {
                "raw": 9942000,
                "fmt": "9.94M",
                "longFmt": "9,942,000"
              },
              "exercisedValue": {
                "raw": 251525498,
                "fmt": "251.53M",
                "longFmt": "251,525,498"
              }
            }
          ]
        },
        "financialData": {
          "maxAge": 86400,
          "currentPrice": 178.72,
          "targetHighPrice": 220.0,
          "targetLowPrice": 140.0,
          "targetMeanPrice": 195.5,
          "targetMedianPrice": 200.0,
          "recommendationMean": 1.8,
          "recommendationKey": "buy",
          "numberOfAnalystOpinions": 42,
          "totalCash": 62655000320,
          "totalDebt": 112086999040,
          "totalRevenue": 383285001216,
          "ebitda": 125676996608,
          "grossMargins": 0.4475,
          "operatingMargins": 0.301,
          "profitMargins": 0.255,
          "financialCurrency": "USD"
        },
        "defaultKeyStatistics": {
          "maxAge": 86400,
          "enterpriseValue": 2819500000000,
          "forwardPE": 27.8,
          "profitMargins": 0.255,
          "floatShares": 15500000000,
          "sharesOutstanding": 15550000000,
          "beta": 1.28,
          "bookValue": 3.85,
          "priceToBook": 46.42,
          "trailingEps": 6.06,
          "forwardEps": 6.43,
          "pegRatio": 3.09
        },
        "price": {
          "averageDailyVolume10Day": 55000000,
          "averageDailyVolume3Month": 60000000,
          "exchange": "NMS",
          "exchangeName": "Nasdaq",
          "maxAge": 86400,
          "regularMarketPrice": 178.72,
          "regularMarketChange": 2.43,
          "regularMarketChangePercent": 1.38,
          "regularMarketVolume": 54321098,
          "regularMarketPreviousClose": 176.29,
          "quoteType": "EQUITY",
          "symbol": "AAPL",
          "shortName": "Apple Inc.",
          "longName": "Apple Inc.",
          "marketState": "REGULAR",
          "marketCap": 2780000000000,
          "currency": "USD"
        },
        "summaryDetail": {
          "maxAge": 86400,
          "previousClose": 176.29,
          "open": 176.5,
          "dayLow": 176.15,
          "dayHigh": 179.62,
          "regularMarketPreviousClose": 176.29,
          "regularMarketOpen": 176.5,
          "regularMarketDayLow": 176.15,
          "regularMarketDayHigh": 179.62,
          "regularMarketVolume": 54321098,
          "volume": 54321098,
          "averageVolume": 60000000,
          "marketCap": 2780000000000,
          "fiftyTwoWeekLow": 124.17,
          "fiftyTwoWeekHigh": 199.62,
          "trailingPE": 29.5,
          "forwardPE": 27.8,
          "dividendRate": 0.96,
          "dividendYield": 0.0054,
          "beta": 1.28,
          "currency": "USD"
        },
        "summaryProfile": {
          "address1": "One Apple Park Way",
          "city": "Cupertino",
          "state": "CA",
          "zip": "95014",
          "country": "United States",
          "phone": "408-996-1010",
          "website": "https://www.apple.com",
          "industry": "Consumer Electronics",
          "sector": "Technology",
          "longBusinessSummary": "Apple Inc. designs, manufactures, and markets smartphones, personal computers, tablets, wearables, and accessories worldwide.",
          "fullTimeEmployees": 164000,
          "maxAge": 86400
        },
        "earnings": {
          "maxAge": 86400,
          "financialCurrency": "USD",
          "earningsChart": {
            "quarterly": [
              {"date": "3Q2023", "actual": 1.26, "estimate": 1.21},
              {"date": "4Q2023", "actual": 1.46, "estimate": 1.39}
            ],
            "earningsDate": [
              {"raw": 1711843200, "fmt": "2024-03-31"},
              {"raw": 1714521600, "fmt": "2024-05-01"}
            ]
          },
          "financialsChart": {
            "yearly": [
              {"date": "2022", "revenue": 394328000000, "earnings": 99803000000},
              {"date": "2023", "revenue": 383285000000, "earnings": 96995000000}
            ],
            "quarterly": [
              {"date": "3Q2023", "revenue": 81797000000, "earnings": 19881000000},
              {"date": "4Q2023", "revenue": 119575000000, "earnings": 33916000000}
            ]
          }
        },
        "recommendationTrend": {
          "trend": [
            {
              "period": "0m",
              "strongBuy": 18,
              "buy": 14,
              "hold": 7,
              "sell": 2,
              "strongSell": 1
            },
            {
              "period": "-1m",
              "strongBuy": 17,
              "buy": 13,
              "hold": 8,
              "sell": 3,
              "strongSell": 1
            }
          ],
          "maxAge": 86400
        },
        "balanceSheetHistory": {
          "maxAge": 86400
        },
        "calendarEvents": {
          "maxAge": 86400
        },
        "quoteType": {
          "maxAge": 86400,
          "quoteType": "EQUITY",
          "exchange": "NMS",
          "symbol": "AAPL",
          "shortName": "Apple Inc.",
          "longName": "Apple Inc.",
          "marketCap": 2780000000000,
          "currency": "USD"
        }
      }
    ],
    "error": null
  }
}`

func TestQuoteSummary_ParsesAllModules(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v10/finance/quoteSummary/AAPL", func(w http.ResponseWriter, r *http.Request) {
		crumb := r.URL.Query().Get("crumb")
		if crumb != "testcrumb" {
			t.Errorf("expected crumb 'testcrumb', got %q", crumb)
		}
		modules := r.URL.Query().Get("modules")
		if !strings.Contains(modules, "price") {
			t.Errorf("expected modules to contain 'price', got %q", modules)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteSummaryJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	result, err := client.QuoteSummary(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// --- AssetProfile ---
	if result.AssetProfile == nil {
		t.Fatal("expected assetProfile to be non-nil")
	}
	ap := result.AssetProfile
	if ap.Address1 != "One Apple Park Way" {
		t.Errorf("expected address1 'One Apple Park Way', got %q", ap.Address1)
	}
	if ap.City != "Cupertino" {
		t.Errorf("expected city 'Cupertino', got %q", ap.City)
	}
	if ap.State != "CA" {
		t.Errorf("expected state 'CA', got %q", ap.State)
	}
	if ap.Country != "United States" {
		t.Errorf("expected country 'United States', got %q", ap.Country)
	}
	if ap.Website != "https://www.apple.com" {
		t.Errorf("expected website 'https://www.apple.com', got %q", ap.Website)
	}
	if ap.Industry != "Consumer Electronics" {
		t.Errorf("expected industry 'Consumer Electronics', got %q", ap.Industry)
	}
	if ap.Sector != "Technology" {
		t.Errorf("expected sector 'Technology', got %q", ap.Sector)
	}
	if ap.FullTimeEmployees == nil || *ap.FullTimeEmployees != 164000 {
		t.Errorf("expected fullTimeEmployees 164000, got %v", ap.FullTimeEmployees)
	}
	if len(ap.CompanyOfficers) != 1 {
		t.Fatalf("expected 1 company officer, got %d", len(ap.CompanyOfficers))
	}
	if ap.CompanyOfficers[0].Name != "Mr. Timothy D. Cook" {
		t.Errorf("expected officer name 'Mr. Timothy D. Cook', got %q", ap.CompanyOfficers[0].Name)
	}
	if ap.CompanyOfficers[0].Title != "CEO & Director" {
		t.Errorf("expected officer title 'CEO & Director', got %q", ap.CompanyOfficers[0].Title)
	}
	if ap.CompanyOfficers[0].YearBorn == nil || *ap.CompanyOfficers[0].YearBorn != 1960 {
		t.Errorf("expected officer yearBorn 1960, got %v", ap.CompanyOfficers[0].YearBorn)
	}

	// --- FinancialData ---
	if result.FinancialData == nil {
		t.Fatal("expected financialData to be non-nil")
	}
	fd := result.FinancialData
	if fd.CurrentPrice == nil || *fd.CurrentPrice != 178.72 {
		t.Errorf("expected currentPrice 178.72, got %v", fd.CurrentPrice)
	}
	if fd.TargetHighPrice == nil || *fd.TargetHighPrice != 220.0 {
		t.Errorf("expected targetHighPrice 220.0, got %v", fd.TargetHighPrice)
	}
	if fd.TargetLowPrice == nil || *fd.TargetLowPrice != 140.0 {
		t.Errorf("expected targetLowPrice 140.0, got %v", fd.TargetLowPrice)
	}
	if fd.TargetMeanPrice == nil || *fd.TargetMeanPrice != 195.5 {
		t.Errorf("expected targetMeanPrice 195.5, got %v", fd.TargetMeanPrice)
	}
	if fd.RecommendationKey != "buy" {
		t.Errorf("expected recommendationKey 'buy', got %q", fd.RecommendationKey)
	}
	if fd.NumberOfAnalystOpinions == nil || *fd.NumberOfAnalystOpinions != 42 {
		t.Errorf("expected numberOfAnalystOpinions 42, got %v", fd.NumberOfAnalystOpinions)
	}
	if fd.GrossMargins == nil || *fd.GrossMargins != 0.4475 {
		t.Errorf("expected grossMargins 0.4475, got %v", fd.GrossMargins)
	}
	if fd.FinancialCurrency != "USD" {
		t.Errorf("expected financialCurrency 'USD', got %q", fd.FinancialCurrency)
	}

	// --- DefaultKeyStatistics ---
	if result.DefaultKeyStatistics == nil {
		t.Fatal("expected defaultKeyStatistics to be non-nil")
	}
	dks := result.DefaultKeyStatistics
	if dks.EnterpriseValue == nil || *dks.EnterpriseValue != 2819500000000 {
		t.Errorf("expected enterpriseValue 2819500000000, got %v", dks.EnterpriseValue)
	}
	if dks.ForwardPE == nil || *dks.ForwardPE != 27.8 {
		t.Errorf("expected forwardPE 27.8, got %v", dks.ForwardPE)
	}
	if dks.Beta == nil || *dks.Beta != 1.28 {
		t.Errorf("expected beta 1.28, got %v", dks.Beta)
	}
	if dks.PriceToBook == nil || *dks.PriceToBook != 46.42 {
		t.Errorf("expected priceToBook 46.42, got %v", dks.PriceToBook)
	}
	if dks.PegRatio == nil || *dks.PegRatio != 3.09 {
		t.Errorf("expected pegRatio 3.09, got %v", dks.PegRatio)
	}

	// --- Price ---
	if result.Price == nil {
		t.Fatal("expected price to be non-nil")
	}
	p := result.Price
	if p.RegularMarketPrice == nil || *p.RegularMarketPrice != 178.72 {
		t.Errorf("expected regularMarketPrice 178.72, got %v", p.RegularMarketPrice)
	}
	if p.Symbol != "AAPL" {
		t.Errorf("expected symbol 'AAPL', got %q", p.Symbol)
	}
	if p.ShortName != "Apple Inc." {
		t.Errorf("expected shortName 'Apple Inc.', got %q", p.ShortName)
	}
	if p.MarketState != "REGULAR" {
		t.Errorf("expected marketState 'REGULAR', got %q", p.MarketState)
	}
	if p.Currency != "USD" {
		t.Errorf("expected currency 'USD', got %q", p.Currency)
	}
	if p.AverageDailyVolume10Day == nil || *p.AverageDailyVolume10Day != 55000000 {
		t.Errorf("expected averageDailyVolume10Day 55000000, got %v", p.AverageDailyVolume10Day)
	}

	// --- SummaryDetail ---
	if result.SummaryDetail == nil {
		t.Fatal("expected summaryDetail to be non-nil")
	}
	sd := result.SummaryDetail
	if sd.PreviousClose == nil || *sd.PreviousClose != 176.29 {
		t.Errorf("expected previousClose 176.29, got %v", sd.PreviousClose)
	}
	if sd.FiftyTwoWeekLow == nil || *sd.FiftyTwoWeekLow != 124.17 {
		t.Errorf("expected fiftyTwoWeekLow 124.17, got %v", sd.FiftyTwoWeekLow)
	}
	if sd.FiftyTwoWeekHigh == nil || *sd.FiftyTwoWeekHigh != 199.62 {
		t.Errorf("expected fiftyTwoWeekHigh 199.62, got %v", sd.FiftyTwoWeekHigh)
	}
	if sd.TrailingPE == nil || *sd.TrailingPE != 29.5 {
		t.Errorf("expected trailingPE 29.5, got %v", sd.TrailingPE)
	}
	if sd.DividendRate == nil || *sd.DividendRate != 0.96 {
		t.Errorf("expected dividendRate 0.96, got %v", sd.DividendRate)
	}
	if sd.Currency != "USD" {
		t.Errorf("expected currency 'USD', got %q", sd.Currency)
	}

	// --- SummaryProfile ---
	if result.SummaryProfile == nil {
		t.Fatal("expected summaryProfile to be non-nil")
	}
	sp := result.SummaryProfile
	if sp.City != "Cupertino" {
		t.Errorf("expected city 'Cupertino', got %q", sp.City)
	}
	if sp.Sector != "Technology" {
		t.Errorf("expected sector 'Technology', got %q", sp.Sector)
	}
	if sp.FullTimeEmployees == nil || *sp.FullTimeEmployees != 164000 {
		t.Errorf("expected fullTimeEmployees 164000, got %v", sp.FullTimeEmployees)
	}

	// --- Earnings ---
	if result.Earnings == nil {
		t.Fatal("expected earnings to be non-nil")
	}
	e := result.Earnings
	if e.FinancialCurrency != "USD" {
		t.Errorf("expected financialCurrency 'USD', got %q", e.FinancialCurrency)
	}
	if e.EarningsChart == nil {
		t.Fatal("expected earningsChart to be non-nil")
	}
	if len(e.EarningsChart.Quarterly) != 2 {
		t.Fatalf("expected 2 quarterly earnings, got %d", len(e.EarningsChart.Quarterly))
	}
	q1 := e.EarningsChart.Quarterly[0]
	if q1.Date != "3Q2023" {
		t.Errorf("expected quarterly date '3Q2023', got %q", q1.Date)
	}
	if q1.Actual == nil || *q1.Actual != 1.26 {
		t.Errorf("expected quarterly actual 1.26, got %v", q1.Actual)
	}
	if q1.Estimate == nil || *q1.Estimate != 1.21 {
		t.Errorf("expected quarterly estimate 1.21, got %v", q1.Estimate)
	}
	if len(e.EarningsChart.EarningsDate) != 2 {
		t.Fatalf("expected 2 earnings dates, got %d", len(e.EarningsChart.EarningsDate))
	}
	if e.FinancialsChart == nil {
		t.Fatal("expected financialsChart to be non-nil")
	}
	if len(e.FinancialsChart.Yearly) != 2 {
		t.Fatalf("expected 2 yearly financials, got %d", len(e.FinancialsChart.Yearly))
	}
	yr := e.FinancialsChart.Yearly[0]
	if yr.Date != "2022" {
		t.Errorf("expected yearly date '2022', got %q", yr.Date)
	}
	if yr.Revenue == nil || *yr.Revenue != 394328000000 {
		t.Errorf("expected yearly revenue 394328000000, got %v", yr.Revenue)
	}
	if len(e.FinancialsChart.Quarterly) != 2 {
		t.Fatalf("expected 2 quarterly financials, got %d", len(e.FinancialsChart.Quarterly))
	}

	// --- RecommendationTrend ---
	if result.RecommendationTrend == nil {
		t.Fatal("expected recommendationTrend to be non-nil")
	}
	rt := result.RecommendationTrend
	if len(rt.Trend) != 2 {
		t.Fatalf("expected 2 trend items, got %d", len(rt.Trend))
	}
	ti := rt.Trend[0]
	if ti.Period != "0m" {
		t.Errorf("expected period '0m', got %q", ti.Period)
	}
	if ti.StrongBuy != 18 {
		t.Errorf("expected strongBuy 18, got %d", ti.StrongBuy)
	}
	if ti.Buy != 14 {
		t.Errorf("expected buy 14, got %d", ti.Buy)
	}
	if ti.Hold != 7 {
		t.Errorf("expected hold 7, got %d", ti.Hold)
	}
	if ti.Sell != 2 {
		t.Errorf("expected sell 2, got %d", ti.Sell)
	}
	if ti.StrongSell != 1 {
		t.Errorf("expected strongSell 1, got %d", ti.StrongSell)
	}

	// --- QuoteType ---
	if result.QuoteType == nil {
		t.Fatal("expected quoteType to be non-nil")
	}
	qt := result.QuoteType
	if qt.QuoteType != "EQUITY" {
		t.Errorf("expected quoteType 'EQUITY', got %q", qt.QuoteType)
	}
	if qt.Symbol != "AAPL" {
		t.Errorf("expected symbol 'AAPL', got %q", qt.Symbol)
	}
}

func TestQuoteSummary_DefaultModules(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v10/finance/quoteSummary/AAPL", func(w http.ResponseWriter, r *http.Request) {
		modules := r.URL.Query().Get("modules")
		if modules != "price,summaryDetail" {
			t.Errorf("expected default modules 'price,summaryDetail', got %q", modules)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteSummaryJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.QuoteSummary(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestQuoteSummary_AllModules(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v10/finance/quoteSummary/AAPL", func(w http.ResponseWriter, r *http.Request) {
		modules := r.URL.Query().Get("modules")
		expectedModules := strings.Join(allQuoteSummaryModules, ",")
		if modules != expectedModules {
			t.Errorf("expected all modules %q, got %q", expectedModules, modules)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteSummaryJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.QuoteSummary(context.Background(), "AAPL", &QuoteSummaryOptions{Modules: []string{"all"}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestQuoteSummary_CustomModules(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb"))
	})
	mux.HandleFunc("/v10/finance/quoteSummary/AAPL", func(w http.ResponseWriter, r *http.Request) {
		modules := r.URL.Query().Get("modules")
		if modules != "financialData,defaultKeyStatistics" {
			t.Errorf("expected modules 'financialData,defaultKeyStatistics', got %q", modules)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testQuoteSummaryJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.QuoteSummary(context.Background(), "AAPL", &QuoteSummaryOptions{
		Modules: []string{"financialData", "defaultKeyStatistics"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestQuoteSummary_EmptySymbolError(t *testing.T) {
	client := NewClient(&Config{})
	_, err := client.QuoteSummary(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}
	var invOptsErr *InvalidOptionsError
	if !errorAs(err, &invOptsErr) {
		t.Errorf("expected InvalidOptionsError, got %T: %v", err, err)
	}
}

func TestQuoteSummary_APIError(t *testing.T) {
	const errorJSON = `{
	  "quoteSummary": {
	    "result": [],
	    "error": {
	      "code": "Not Found",
	      "description": "No data found for symbol"
	    }
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
	mux.HandleFunc("/v10/finance/quoteSummary/INVALID", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(errorJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.QuoteSummary(context.Background(), "INVALID", nil)
	if err == nil {
		t.Fatal("expected error for API error response, got nil")
	}
}

func TestQuoteSummary_EmptyResultError(t *testing.T) {
	const emptyResultJSON = `{
	  "quoteSummary": {
	    "result": [],
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
	mux.HandleFunc("/v10/finance/quoteSummary/UNKNOWN", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(emptyResultJSON))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	_, err := client.QuoteSummary(context.Background(), "UNKNOWN", nil)
	if err == nil {
		t.Fatal("expected error for empty result, got nil")
	}
	if _, ok := err.(*quoteSummaryParseError); !ok {
		t.Errorf("expected quoteSummaryParseError, got %T: %v", err, err)
	}
}

func TestQuoteSummary_ResolveModules_Defaults(t *testing.T) {
	modules := resolveModules(nil)
	if len(modules) != 2 || modules[0] != "price" || modules[1] != "summaryDetail" {
		t.Errorf("expected default modules [price, summaryDetail], got %v", modules)
	}

	modules = resolveModules(&QuoteSummaryOptions{})
	if len(modules) != 2 || modules[0] != "price" || modules[1] != "summaryDetail" {
		t.Errorf("expected default modules [price, summaryDetail], got %v", modules)
	}
}

func TestQuoteSummary_ResolveModules_All(t *testing.T) {
	modules := resolveModules(&QuoteSummaryOptions{Modules: []string{"all"}})
	if len(modules) != len(allQuoteSummaryModules) {
		t.Errorf("expected %d modules for 'all', got %d", len(allQuoteSummaryModules), len(modules))
	}
}

func TestQuoteSummary_ResolveModules_Custom(t *testing.T) {
	modules := resolveModules(&QuoteSummaryOptions{Modules: []string{"assetProfile", "financialData"}})
	if len(modules) != 2 || modules[0] != "assetProfile" || modules[1] != "financialData" {
		t.Errorf("expected [assetProfile, financialData], got %v", modules)
	}
}
