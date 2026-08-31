package yahoofinance

import (
	"testing"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient(nil)
	if c.Chart == nil {
		t.Error("expected Chart service to be initialized")
	}
	if c.Quote == nil {
		t.Error("expected Quote service to be initialized")
	}
	if c.Historical == nil {
		t.Error("expected Historical service to be initialized")
	}
	if c.Options == nil {
		t.Error("expected Options service to be initialized")
	}
	if c.Fundamentals == nil {
		t.Error("expected Fundamentals service to be initialized")
	}
	if c.Insights == nil {
		t.Error("expected Insights service to be initialized")
	}
	if c.Screener == nil {
		t.Error("expected Screener service to be initialized")
	}
	if c.Search == nil {
		t.Error("expected Search service to be initialized")
	}
	if c.Trending == nil {
		t.Error("expected Trending service to be initialized")
	}
	if c.Recommendations == nil {
		t.Error("expected Recommendations service to be initialized")
	}
	if c.QuoteSummary == nil {
		t.Error("expected QuoteSummary service to be initialized")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	c := NewClient(&Config{
		QueryHost:   "https://custom.host.com",
		Concurrency: 8,
	})
	if c.Chart == nil {
		t.Error("expected Chart service to be initialized")
	}
}
