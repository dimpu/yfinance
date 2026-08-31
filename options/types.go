package options

import (
	"time"

	"github.com/dimpu/yfinance/quote"
)

// Options specifies parameters for the Options API call.
type Options struct {
	Formatted bool       `json:"formatted,omitempty"`
	Lang      string     `json:"lang,omitempty"`
	Region    string     `json:"region,omitempty"`
	Date      *time.Time `json:"date,omitempty"`
}

// Result represents the result of an options chain query.
type Result struct {
	UnderlyingSymbol string    `json:"underlyingSymbol"`
	ExpirationDates  []int64   `json:"expirationDates"`
	Strikes          []float64 `json:"strikes"`
	HasMiniOptions   bool      `json:"hasMiniOptions"`
	Quote            quote.Quote `json:"quote"`
	Options          []Option  `json:"options"`
}

// Option represents a single expiration date's option chain.
type Option struct {
	ExpirationDate int64       `json:"expirationDate"`
	HasMiniOptions bool        `json:"hasMiniOptions"`
	Calls          []CallOrPut `json:"calls"`
	Puts           []CallOrPut `json:"puts"`
}

// CallOrPut represents a single call or put option contract.
type CallOrPut struct {
	ContractSymbol    string   `json:"contractSymbol"`
	Strike            float64  `json:"strike"`
	Currency          string   `json:"currency,omitempty"`
	LastPrice         float64  `json:"lastPrice"`
	Change            float64  `json:"change,omitempty"`
	PercentChange     *float64 `json:"percentChange,omitempty"`
	Volume            *int64   `json:"volume,omitempty"`
	OpenInterest      *int64   `json:"openInterest,omitempty"`
	Bid               *float64 `json:"bid,omitempty"`
	Ask               *float64 `json:"ask,omitempty"`
	ContractSize      string   `json:"contractSize"`
	Expiration        int64    `json:"expiration"`
	LastTradeDate     int64    `json:"lastTradeDate"`
	ImpliedVolatility float64  `json:"impliedVolatility"`
	InTheMoney        bool     `json:"inTheMoney"`
}
