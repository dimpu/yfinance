package historical

import "time"

// Options specifies parameters for the Historical API call.
type Options struct {
	Period1              time.Time `validate:"required"`
	Period2              time.Time
	Interval             string `validate:"omitempty,oneof=1d 1wk 1mo"`
	Events               string `validate:"omitempty,oneof=history dividends split"`
	IncludeAdjustedClose bool
}

// RowHistory represents a single price history row.
type RowHistory struct {
	Date     time.Time `json:"date"`
	Open     *float64  `json:"open"`
	High     *float64  `json:"high"`
	Low      *float64  `json:"low"`
	Close    *float64  `json:"close"`
	AdjClose *float64  `json:"adjClose"`
	Volume   *int64    `json:"volume"`
}

// RowDividend represents a single dividend event row.
type RowDividend struct {
	Date      time.Time `json:"date"`
	Dividends float64   `json:"dividends"`
}

// RowStockSplit represents a single stock split event row.
type RowStockSplit struct {
	Date        time.Time `json:"date"`
	StockSplits string    `json:"stockSplits"`
}
