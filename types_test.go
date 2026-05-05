package yahoofinance

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTwoNumberRangeUnmarshalString(t *testing.T) {
	var r TwoNumberRange
	if err := json.Unmarshal([]byte(`"180.68 - 589.07"`), &r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Low != 180.68 || r.High != 589.07 {
		t.Errorf("expected Low=180.68 High=589.07, got Low=%v High=%v", r.Low, r.High)
	}
}

func TestTwoNumberRangeUnmarshalObject(t *testing.T) {
	var r TwoNumberRange
	if err := json.Unmarshal([]byte(`{"low":100.5,"high":200.5}`), &r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Low != 100.5 || r.High != 200.5 {
		t.Errorf("expected Low=100.5 High=200.5, got Low=%v High=%v", r.Low, r.High)
	}
}

func TestTwoNumberRangeUnmarshalInvalidString(t *testing.T) {
	var r TwoNumberRange
	if err := json.Unmarshal([]byte(`"invalid"`), &r); err == nil {
		t.Error("expected error for invalid string, got nil")
	}
}

func TestYahooDateUnmarshalTimestamp(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`1700000000`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Unix(1700000000, 0)
	if !d.Time.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, d.Time)
	}
}

func TestYahooDateUnmarshalRawObject(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`{"raw":1700000000,"fmt":"2023-11-14"}`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Unix(1700000000, 0)
	if !d.Time.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, d.Time)
	}
}

func TestYahooDateUnmarshalDateString(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`"2023-11-14T00:00:00Z"`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Year() != 2023 || d.Month() != time.November || d.Day() != 14 {
		t.Errorf("expected 2023-11-14, got %v", d.Time)
	}
}

func TestYahooDateUnmarshalDateOnly(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`"2023-11-14"`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Year() != 2023 || d.Month() != time.November || d.Day() != 14 {
		t.Errorf("expected 2023-11-14, got %v", d.Time)
	}
}

func TestYahooDateUnmarshalInvalid(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`"not-a-date"`), &d); err == nil {
		t.Error("expected error for invalid date, got nil")
	}
}
