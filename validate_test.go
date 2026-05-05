package yahoofinance

import (
	"testing"
	"time"
)

func TestValidateOptionsValid(t *testing.T) {
	type testOpts struct {
		Period1  time.Time `validate:"required"`
		Interval string   `validate:"omitempty,oneof=1m 2m 5m 15m 30m 60m 90m 1h 1d 5d 1wk 1mo 3mo"`
	}
	opts := testOpts{
		Period1:  time.Now(),
		Interval: "1d",
	}
	if err := validateOptions(opts); err != nil {
		t.Errorf("expected valid, got %v", err)
	}
}

func TestValidateOptionsInvalid(t *testing.T) {
	type testOpts struct {
		Period1  time.Time `validate:"required"`
		Interval string   `validate:"omitempty,oneof=1m 2m 5m 15m 30m 60m 90m 1h 1d 5d 1wk 1mo 3mo"`
	}
	opts := testOpts{
		Period1:  time.Time{},
		Interval: "invalid",
	}
	if err := validateOptions(opts); err == nil {
		t.Error("expected validation error, got nil")
	}
}
