package yahoofinance

import "testing"

func TestNewClientDefaults(t *testing.T) {
	c := NewClient(nil)
	if c.queryHost != defaultQueryHost {
		t.Errorf("expected queryHost %s, got %s", defaultQueryHost, c.queryHost)
	}
	if c.validation.LogErrors != true {
		t.Error("expected LogErrors true by default")
	}
	if c.validation.AllowAdditionalProps != true {
		t.Error("expected AllowAdditionalProps true by default")
	}
}

func TestNewClientConfig(t *testing.T) {
	c := NewClient(&Config{
		QueryHost:   "custom.host.com",
		Concurrency: 8,
	})
	if c.queryHost != "custom.host.com" {
		t.Errorf("expected custom host, got %s", c.queryHost)
	}
}
