package trending

// Options controls optional parameters for the Trending endpoint.
type Options struct {
	Lang   string `json:"lang,omitempty"`
	Region string `json:"region,omitempty"`
	Count  int    `json:"count,omitempty"`
}

// Symbol represents a single trending symbol.
type Symbol struct {
	Symbol string `json:"symbol"`
}

// Result holds the trending symbols data for a region.
type Result struct {
	Count         int      `json:"count"`
	Quotes        []Symbol `json:"quotes,omitempty"`
	JobTimestamp  int64    `json:"jobTimestamp"`
	StartInterval int64    `json:"startInterval"`
}
