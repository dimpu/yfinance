package yahoofinance

// RecommendationsOptions controls optional parameters for the RecommendationsBySymbol endpoint.
type RecommendationsOptions struct{}

// RecommendedSymbol represents a single recommended symbol with its score.
type RecommendedSymbol struct {
	Score  float64 `json:"score"`
	Symbol string  `json:"symbol"`
}

// RecommendationsResult holds the recommendation data for a single symbol.
type RecommendationsResult struct {
	RecommendedSymbols []RecommendedSymbol `json:"recommendedSymbols"`
	Symbol             string              `json:"symbol"`
}
