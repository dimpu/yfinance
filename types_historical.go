package yahoofinance

import "time"

// HistoricalOptions specifies parameters for the Historical API call.
type HistoricalOptions struct {
	Period1              time.Time `validate:"required"`
	Period2              time.Time
	Interval             string `validate:"omitempty,oneof=1d 1wk 1mo"`
	Events               string `validate:"omitempty,oneof=history dividends split"`
	IncludeAdjustedClose bool
}

// HistoricalRowHistory represents a single price history row.
type HistoricalRowHistory struct {
	Date     time.Time `json:"date"`
	Open     *float64  `json:"open"`
	High     *float64  `json:"high"`
	Low      *float64  `json:"low"`
	Close    *float64  `json:"close"`
	AdjClose *float64  `json:"adjClose"`
	Volume   *int64    `json:"volume"`
}

// HistoricalRowDividend represents a single dividend event row.
type HistoricalRowDividend struct {
	Date      time.Time `json:"date"`
	Dividends float64   `json:"dividends"`
}

// HistoricalRowStockSplit represents a single stock split event row.
type HistoricalRowStockSplit struct {
	Date        time.Time `json:"date"`
	StockSplits string    `json:"stockSplits"`
}
