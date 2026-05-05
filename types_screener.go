package yahoofinance

type ScreenerOptions struct {
	Lang   string `json:"lang,omitempty"`
	Region string `json:"region,omitempty"`
	Count  int    `json:"count,omitempty"`
}

// PredefinedScreenerModule are the 15 available screener IDs.
type PredefinedScreenerModule string

const (
	ScreenerAggressiveSmallCaps      PredefinedScreenerModule = "aggressive_small_caps"
	ScreenerConservativeForeignFunds PredefinedScreenerModule = "conservative_foreign_funds"
	ScreenerDayGainers               PredefinedScreenerModule = "day_gainers"
	ScreenerDayLosers                PredefinedScreenerModule = "day_losers"
	ScreenerGrowthTechStocks         PredefinedScreenerModule = "growth_technology_stocks"
	ScreenerHighYieldBond            PredefinedScreenerModule = "high_yield_bond"
	ScreenerMostActives              PredefinedScreenerModule = "most_actives"
	ScreenerMostShorted              PredefinedScreenerModule = "most_shorted_stocks"
	ScreenerPortfolioAnchors         PredefinedScreenerModule = "portfolio_anchors"
	ScreenerSmallCapGainers          PredefinedScreenerModule = "small_cap_gainers"
	ScreenerSolidLargeGrowthFunds    PredefinedScreenerModule = "solid_large_growth_funds"
	ScreenerSolidMidcapGrowthFunds   PredefinedScreenerModule = "solid_midcap_growth_funds"
	ScreenerTopMutualFunds           PredefinedScreenerModule = "top_mutual_funds"
	ScreenerUndervaluedGrowth        PredefinedScreenerModule = "undervalued_growth_stocks"
	ScreenerUndervaluedLargeCaps     PredefinedScreenerModule = "undervalued_large_caps"
)

type ScreenerResult struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	CanonicalName string          `json:"canonicalName"`
	Count         int             `json:"count"`
	Total         int             `json:"total"`
	Quotes        []ScreenerQuote `json:"quotes"`
}

type ScreenerQuote struct {
	Symbol                    string   `json:"symbol"`
	ShortName                 string   `json:"shortName,omitempty"`
	LongName                  string   `json:"longName,omitempty"`
	QuoteType                 string   `json:"quoteType"`
	Exchange                  string   `json:"exchange"`
	Currency                  string   `json:"currency"`
	MarketState               string   `json:"marketState"`
	RegularMarketPrice        *float64 `json:"regularMarketPrice"`
	RegularMarketChange       *float64 `json:"regularMarketChange"`
	RegularMarketChangePercent *float64 `json:"regularMarketChangePercent"`
	RegularMarketVolume       *int64   `json:"regularMarketVolume"`
	MarketCap                 *int64   `json:"marketCap,omitempty"`
	FiftyTwoWeekLow           *float64 `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh          *float64 `json:"fiftyTwoWeekHigh"`
	AverageDailyVolume3Month  *int64   `json:"averageDailyVolume3Month,omitempty"`
}
