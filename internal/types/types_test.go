package types

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
	if r.Low != 180.68 {
		t.Errorf("Low = %v, want 180.68", r.Low)
	}
	if r.High != 589.07 {
		t.Errorf("High = %v, want 589.07", r.High)
	}
}

func TestTwoNumberRangeUnmarshalObject(t *testing.T) {
	var r TwoNumberRange
	if err := json.Unmarshal([]byte(`{"low":100.5,"high":200.5}`), &r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Low != 100.5 {
		t.Errorf("Low = %v, want 100.5", r.Low)
	}
	if r.High != 200.5 {
		t.Errorf("High = %v, want 200.5", r.High)
	}
}

func TestTwoNumberRangeUnmarshalInvalidString(t *testing.T) {
	var r TwoNumberRange
	if err := json.Unmarshal([]byte(`"invalid"`), &r); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestYahooDateUnmarshalTimestamp(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`1700000000`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(1700000000, 0)
	if !d.Time.Equal(want) {
		t.Errorf("Time = %v, want %v", d.Time, want)
	}
}

func TestYahooDateUnmarshalRawObject(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`{"raw":1700000000,"fmt":"2023-11-14"}`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(1700000000, 0)
	if !d.Time.Equal(want) {
		t.Errorf("Time = %v, want %v", d.Time, want)
	}
}

func TestYahooDateUnmarshalDateString(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`"2023-11-14T00:00:00Z"`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2023, time.November, 14, 0, 0, 0, 0, time.UTC)
	if !d.Time.Equal(want) {
		t.Errorf("Time = %v, want %v", d.Time, want)
	}
}

func TestYahooDateUnmarshalDateOnly(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`"2023-11-14"`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2023, time.November, 14, 0, 0, 0, 0, time.UTC)
	if !d.Time.Equal(want) {
		t.Errorf("Time = %v, want %v", d.Time, want)
	}
}

func TestYahooDateUnmarshalInvalid(t *testing.T) {
	var d YahooDate
	if err := json.Unmarshal([]byte(`"not-a-date"`), &d); err == nil {
		t.Fatal("expected error, got nil")
	}
}
