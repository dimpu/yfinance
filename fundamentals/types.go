package fundamentals

import "time"

// Options contains options for the FundamentalsTimeSeries API.
type Options struct {
	Period1       time.Time `validate:"required"`
	Period2       time.Time
	Type          string `validate:"omitempty,oneof=quarterly annual trailing"`
	Merge         bool
	PadTimeSeries bool
	Lang          string
	Region        string
	Module        string `validate:"required,oneof=financials balance-sheet cash-flow all"`
}

// Result represents a single time series data point.
type Result struct {
	Date       time.Time           `json:"date"`
	Type       string              `json:"TYPE"`
	PeriodType string              `json:"periodType"`
	Fields     map[string]*float64 `json:"-"`
}
