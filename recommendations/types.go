package recommendations

// Options controls optional parameters for the Recommendations endpoint.
type Options struct{}

// RecommendedSymbol represents a single recommended symbol with its score.
type RecommendedSymbol struct {
	Score  float64 `json:"score"`
	Symbol string  `json:"symbol"`
}

// Result holds the recommendation data for a single symbol.
type Result struct {
	RecommendedSymbols []RecommendedSymbol `json:"recommendedSymbols"`
	Symbol             string              `json:"symbol"`
}
