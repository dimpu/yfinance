package screener

// Options controls optional parameters for the Screener endpoint.
type Options struct {
	Lang   string `json:"lang,omitempty"`
	Region string `json:"region,omitempty"`
	Count  int    `json:"count,omitempty"`
}

// PredefinedScreenerModule are the 15 available screener IDs.
type PredefinedScreenerModule string

const (
	AggressiveSmallCaps      PredefinedScreenerModule = "aggressive_small_caps"
	ConservativeForeignFunds PredefinedScreenerModule = "conservative_foreign_funds"
	DayGainers               PredefinedScreenerModule = "day_gainers"
	DayLosers                PredefinedScreenerModule = "day_losers"
	GrowthTechStocks         PredefinedScreenerModule = "growth_technology_stocks"
	HighYieldBond            PredefinedScreenerModule = "high_yield_bond"
	MostActives              PredefinedScreenerModule = "most_actives"
	MostShorted              PredefinedScreenerModule = "most_shorted_stocks"
	PortfolioAnchors         PredefinedScreenerModule = "portfolio_anchors"
	SmallCapGainers          PredefinedScreenerModule = "small_cap_gainers"
	SolidLargeGrowthFunds    PredefinedScreenerModule = "solid_large_growth_funds"
	SolidMidcapGrowthFunds   PredefinedScreenerModule = "solid_midcap_growth_funds"
	TopMutualFunds           PredefinedScreenerModule = "top_mutual_funds"
	UndervaluedGrowth        PredefinedScreenerModule = "undervalued_growth_stocks"
	UndervaluedLargeCaps     PredefinedScreenerModule = "undervalued_large_caps"
)

// Result holds the response from the Screener endpoint.
type Result struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	CanonicalName string          `json:"canonicalName"`
	Count         int             `json:"count"`
	Total         int             `json:"total"`
	Quotes        []ScreenerQuote `json:"quotes"`
}

// ScreenerQuote represents a single quote from a screener result.
type ScreenerQuote struct {
	Symbol                     string   `json:"symbol"`
	ShortName                  string   `json:"shortName,omitempty"`
	LongName                   string   `json:"longName,omitempty"`
	QuoteType                  string   `json:"quoteType"`
	Exchange                   string   `json:"exchange"`
	Currency                   string   `json:"currency"`
	MarketState                string   `json:"marketState"`
	RegularMarketPrice         *float64 `json:"regularMarketPrice"`
	RegularMarketChange        *float64 `json:"regularMarketChange"`
	RegularMarketChangePercent *float64 `json:"regularMarketChangePercent"`
	RegularMarketVolume        *int64   `json:"regularMarketVolume"`
	MarketCap                  *int64   `json:"marketCap,omitempty"`
	FiftyTwoWeekLow            *float64 `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh           *float64 `json:"fiftyTwoWeekHigh"`
	AverageDailyVolume3Month   *int64   `json:"averageDailyVolume3Month,omitempty"`
}
