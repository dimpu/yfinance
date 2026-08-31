package quote

import "github.com/dimpu/yfinance/internal/types"

// Options specifies parameters for the Quote API call.
type Options struct {
	Fields []string `json:"fields,omitempty"`
	Lang   string   `json:"lang,omitempty"`
	Region string   `json:"region,omitempty"`
}

// Quote represents quote data for a financial instrument.
// Yahoo returns a discriminated union by quoteType, but we use a single struct
// with all possible fields (flat, like the JS lib's QuoteBase + subtype fields combined).
type Quote struct {
	// Base fields
	Language  string `json:"language"`
	Region    string `json:"region"`
	QuoteType string `json:"quoteType"` // discriminator: "EQUITY", "ETF", "CRYPTOCURRENCY", "OPTION", "FUTURE", "INDEX", "MUTUALFUND", "CURRENCY", "MONEYMARKET", "ECNQUOTE", "ALTSYMBOL"
	TypeDisp  string `json:"typeDisp,omitempty"`

	Currency    string `json:"currency,omitempty"`
	MarketState string `json:"marketState"`

	Symbol    string `json:"symbol"`
	ShortName string `json:"shortName,omitempty"`
	LongName  string `json:"longName,omitempty"`

	// Regular market data
	RegularMarketPrice         *float64 `json:"regularMarketPrice,omitempty"`
	RegularMarketChange        *float64 `json:"regularMarketChange,omitempty"`
	RegularMarketChangePercent *float64 `json:"regularMarketChangePercent,omitempty"`
	RegularMarketVolume        *int64   `json:"regularMarketVolume,omitempty"`
	RegularMarketOpen          *float64 `json:"regularMarketOpen,omitempty"`
	RegularMarketDayHigh       *float64 `json:"regularMarketDayHigh,omitempty"`
	RegularMarketDayLow        *float64 `json:"regularMarketDayLow,omitempty"`
	RegularMarketTime          *int64   `json:"regularMarketTime,omitempty"`
	RegularMarketPreviousClose *float64 `json:"regularMarketPreviousClose,omitempty"`

	// 52-week data
	FiftyTwoWeekLow                *float64              `json:"fiftyTwoWeekLow,omitempty"`
	FiftyTwoWeekHigh               *float64              `json:"fiftyTwoWeekHigh,omitempty"`
	FiftyTwoWeekRange              *types.TwoNumberRange `json:"fiftyTwoWeekRange,omitempty"`
	FiftyTwoWeekLowChange          *float64              `json:"fiftyTwoWeekLowChange,omitempty"`
	FiftyTwoWeekLowChangePercent   *float64              `json:"fiftyTwoWeekLowChangePercent,omitempty"`
	FiftyTwoWeekHighChange         *float64              `json:"fiftyTwoWeekHighChange,omitempty"`
	FiftyTwoWeekHighChangePercent  *float64              `json:"fiftyTwoWeekHighChangePercent,omitempty"`
	FiftyTwoWeekChangePercent      *float64              `json:"fiftyTwoWeekChangePercent,omitempty"`

	// Moving averages
	FiftyDayAverage                   *float64 `json:"fiftyDayAverage,omitempty"`
	FiftyDayAverageChange             *float64 `json:"fiftyDayAverageChange,omitempty"`
	FiftyDayAverageChangePercent      *float64 `json:"fiftyDayAverageChangePercent,omitempty"`
	TwoHundredDayAverage              *float64 `json:"twoHundredDayAverage,omitempty"`
	TwoHundredDayAverageChange        *float64 `json:"twoHundredDayAverageChange,omitempty"`
	TwoHundredDayAverageChangePercent *float64 `json:"twoHundredDayAverageChangePercent,omitempty"`

	// Valuation metrics
	MarketCap         *int64   `json:"marketCap,omitempty"`
	SharesOutstanding *int64   `json:"sharesOutstanding,omitempty"`
	FloatShares       *int64   `json:"floatShares,omitempty"`
	TrailingPE        *float64 `json:"trailingPE,omitempty"`
	ForwardPE         *float64 `json:"forwardPE,omitempty"`
	EPS               *float64 `json:"epsTrailingTwelveMonths,omitempty"`
	EPSForward        *float64 `json:"epsForward,omitempty"`
	BookValue         *float64 `json:"bookValue,omitempty"`
	PriceToBook       *float64 `json:"priceToBook,omitempty"`

	// Dividend data
	TrailingAnnualDividendRate  *float64        `json:"trailingAnnualDividendRate,omitempty"`
	TrailingAnnualDividendYield *float64        `json:"trailingAnnualDividendYield,omitempty"`
	DividendDate                *types.YahooDate `json:"dividendDate,omitempty"`

	// Volume data
	AverageDailyVolume3Month *int64 `json:"averageDailyVolume3Month,omitempty"`
	AverageDailyVolume10Day  *int64 `json:"averageDailyVolume10Day,omitempty"`

	// Bid/Ask
	Bid     *float64 `json:"bid,omitempty"`
	Ask     *float64 `json:"ask,omitempty"`
	BidSize *int64   `json:"bidSize,omitempty"`
	AskSize *int64   `json:"askSize,omitempty"`

	// Exchange info
	FullExchangeName  string `json:"fullExchangeName,omitempty"`
	FinancialCurrency string `json:"financialCurrency,omitempty"`
	Exchange          string `json:"exchange,omitempty"`

	// Earnings
	EarningsTimestamp *int64 `json:"earningsTimestamp,omitempty"`

	// Post-market data
	PostMarketPrice         *float64 `json:"postMarketPrice,omitempty"`
	PostMarketChange        *float64 `json:"postMarketChange,omitempty"`
	PostMarketChangePercent *float64 `json:"postMarketChangePercent,omitempty"`

	// Pre-market data
	PreMarketPrice         *float64 `json:"preMarketPrice,omitempty"`
	PreMarketChange        *float64 `json:"preMarketChange,omitempty"`
	PreMarketChangePercent *float64 `json:"preMarketChangePercent,omitempty"`

	// Flags
	Tradeable       bool `json:"tradeable"`
	CryptoTradeable bool `json:"cryptoTradeable,omitempty"`
	ESGPopulated    bool `json:"esgPopulated"`
	Triggerable     bool `json:"triggerable"`

	// Subtype-specific fields
	// Equity/ETF
	DividendRate  *float64 `json:"dividendRate,omitempty"`
	DividendYield *float64 `json:"dividendYield,omitempty"`

	// Crypto
	CirculatingSupply *float64 `json:"circulatingSupply,omitempty"`
	FromCurrency      string   `json:"fromCurrency,omitempty"`
	ToCurrency        string   `json:"toCurrency,omitempty"`

	// Option/Future
	Strike           *float64         `json:"strike,omitempty"`
	OpenInterest     *int64           `json:"openInterest,omitempty"`
	ExpireDate       *types.YahooDate `json:"expireDate,omitempty"`
	UnderlyingSymbol string           `json:"underlyingSymbol,omitempty"`

	// ETF/MutualFund
	NetAssets *int64   `json:"netAssets,omitempty"`
	Beta      *float64 `json:"beta,omitempty"`

	// Additional fields from JS implementation
	QuoteSourceName            string  `json:"quoteSourceName,omitempty"`
	CustomPriceAlertConfidence string  `json:"customPriceAlertConfidence,omitempty"`
	MessageBoardId             string  `json:"messageBoardId,omitempty"`
	ExchangeTimezoneName       string  `json:"exchangeTimezoneName,omitempty"`
	ExchangeTimezoneShortName  string  `json:"exchangeTimezoneShortName,omitempty"`
	GmtOffSetMilliseconds      int64   `json:"gmtOffSetMilliseconds,omitempty"`
	Market                     string  `json:"market,omitempty"`
	SourceInterval             int     `json:"sourceInterval,omitempty"`
	ExchangeDataDelayedBy      int     `json:"exchangeDataDelayedBy,omitempty"`
	PriceHint                  *int64  `json:"priceHint,omitempty"`
	DisplayName                string  `json:"displayName,omitempty"`
	FirstTradeDateMilliseconds *int64  `json:"firstTradeDateMilliseconds,omitempty"`
	HasPrePostMarketData       bool    `json:"hasPrePostMarketData,omitempty"`
}
