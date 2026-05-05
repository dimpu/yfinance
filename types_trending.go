package yahoofinance

// TrendingOptions controls optional parameters for the TrendingSymbols endpoint.
type TrendingOptions struct {
	Lang   string `json:"lang,omitempty"`
	Region string `json:"region,omitempty"`
	Count  int    `json:"count,omitempty"`
}

// TrendingSymbol represents a single trending symbol.
type TrendingSymbol struct {
	Symbol string `json:"symbol"`
}

// TrendingResult holds the trending symbols data for a region.
type TrendingResult struct {
	Count         int              `json:"count"`
	Quotes        []TrendingSymbol `json:"quotes,omitempty"`
	JobTimestamp  int64            `json:"jobTimestamp"`
	StartInterval int64            `json:"startInterval"`
}
