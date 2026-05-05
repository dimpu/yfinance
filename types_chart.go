package yahoofinance

import "time"

// ChartOptions specifies parameters for the Chart API call.
type ChartOptions struct {
	Period1        time.Time `validate:"required" json:"period1"`
	Period2        time.Time `json:"period2,omitempty"`
	Interval       string    `validate:"omitempty,oneof=1m 2m 5m 15m 30m 60m 90m 1h 1d 5d 1wk 1mo 3mo" json:"interval,omitempty"`
	IncludePrePost bool      `json:"includePrePost,omitempty"`
	Events         string    `json:"events,omitempty"`
	Lang           string    `json:"lang,omitempty"`
	Return         string    `json:"-"` // "array" (default) or "object", not sent to Yahoo
}

// ChartResult is the parsed result from the Chart API.
type ChartResult struct {
	Meta   ChartMeta   `json:"meta"`
	Quotes []ChartQuote `json:"quotes,omitempty"`
	Events *ChartEvents `json:"events,omitempty"`
}

// ChartMeta contains metadata about the chart result.
type ChartMeta struct {
	Currency             string   `json:"currency"`
	Symbol               string   `json:"symbol"`
	ExchangeName         string   `json:"exchangeName"`
	InstrumentType       string   `json:"instrumentType"`
	FiftyTwoWeekHigh     *float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      *float64 `json:"fiftyTwoWeekLow"`
	FirstTradeDate       *int64   `json:"firstTradeDate"`
	FullExchangeName     string   `json:"fullExchangeName"`
	RegularMarketTime    *int64   `json:"regularMarketTime"`
	GMTOffset            *int64   `json:"gmtoffset"`
	HasPrePostMarketData bool     `json:"hasPrePostMarketData"`
	Timezone             string   `json:"timezone"`
	ExchangeTimezoneName string   `json:"exchangeTimezoneName"`
	RegularMarketPrice   *float64 `json:"regularMarketPrice"`
	ChartPreviousClose   *float64 `json:"chartPreviousClose"`
	PreviousClose        *float64 `json:"previousClose"`
	RegularMarketDayHigh *float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow  *float64 `json:"regularMarketDayLow"`
	RegularMarketVolume  *int64   `json:"regularMarketVolume"`
	LongName             string   `json:"longName"`
	ShortName            string   `json:"shortName"`
	Scale                *int64   `json:"scale"`
	PriceHint            *int64   `json:"priceHint"`
	DataGranularity      string   `json:"dataGranularity"`
	Range                string   `json:"range"`
	ValidRanges          []string `json:"validRanges"`
}

// ChartQuote represents a single OHLCV data point.
type ChartQuote struct {
	Date     time.Time `json:"date"`
	Open     *float64  `json:"open"`
	High     *float64  `json:"high"`
	Low      *float64  `json:"low"`
	Close    *float64  `json:"close"`
	Volume   *int64    `json:"volume"`
	AdjClose *float64  `json:"adjclose,omitempty"`
}

// ChartEvents contains dividend and split events.
type ChartEvents struct {
	Dividends map[int64]ChartEventDividend `json:"dividends,omitempty"`
	Splits    map[int64]ChartEventSplit    `json:"splits,omitempty"`
}

// ChartEventDividend represents a dividend event.
type ChartEventDividend struct {
	Amount      float64 `json:"amount"`
	Date        int64   `json:"date"`
	Description string  `json:"description"`
}

// ChartEventSplit represents a stock split event.
type ChartEventSplit struct {
	Date        int64   `json:"date"`
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
	SplitRatio  string  `json:"splitRatio"`
}
