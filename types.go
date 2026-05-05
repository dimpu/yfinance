package yahoofinance

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TwoNumberRange represents a Yahoo range like "180.68 - 589.07".
type TwoNumberRange struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

// UnmarshalJSON handles both "180.68 - 589.07" string format
// and {"low":180.68,"high":589.07} object format.
func (r *TwoNumberRange) UnmarshalJSON(data []byte) error {
	// Try string format "low - high" first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parts := strings.SplitN(s, " - ", 2)
		if len(parts) == 2 {
			low, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			high, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				r.Low = low
				r.High = high
				return nil
			}
		}
		return fmt.Errorf("invalid TwoNumberRange string: %s", s)
	}
	// Try object format
	type alias struct {
		Low  float64 `json:"low"`
		High float64 `json:"high"`
	}
	var obj alias
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	r.Low = obj.Low
	r.High = obj.High
	return nil
}

// YahooDate handles Yahoo's {"raw": timestamp, "fmt": "date_string"} format.
type YahooDate struct {
	time.Time
}

func (d *YahooDate) UnmarshalJSON(data []byte) error {
	// Try numeric timestamp first
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		d.Time = time.Unix(int64(num), 0)
		return nil
	}
	// Try {"raw": timestamp, "fmt": "..."} object
	var obj struct {
		Raw float64 `json:"raw"`
		Fmt string  `json:"fmt"`
	}
	if err := json.Unmarshal(data, &obj); err == nil && obj.Raw != 0 {
		d.Time = time.Unix(int64(obj.Raw), 0)
		return nil
	}
	// Try RFC3339 / ISO8601 string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t, err = time.Parse("2006-01-02", s)
		}
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	}
	return fmt.Errorf("cannot unmarshal YahooDate from: %s", string(data))
}

// ModuleOptions controls per-call behavior.
type ModuleOptions struct {
	ValidateResult bool
	ValidateOptions bool
	FetchOptions   *FetchOptions
}

// FetchOptions carries additional HTTP headers.
type FetchOptions struct {
	Headers map[string]string
}

// ValidationOpts controls validation behavior.
type ValidationOpts struct {
	LogErrors            bool
	AllowAdditionalProps bool
}
