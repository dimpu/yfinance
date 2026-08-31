package search

// Options controls optional parameters for the Search endpoint.
type Options struct {
	Lang                       string `json:"lang,omitempty"`
	Region                     string `json:"region,omitempty"`
	QuotesCount                int    `json:"quotesCount,omitempty"`
	NewsCount                  int    `json:"newsCount,omitempty"`
	EnableFuzzyQuery           bool   `json:"enableFuzzyQuery,omitempty"`
	EnableCb                   bool   `json:"enableCb,omitempty"`
	EnableNavLinks             bool   `json:"enableNavLinks,omitempty"`
	EnableEnhancedTrivialQuery bool   `json:"enableEnhancedTrivialQuery,omitempty"`
}

// Result holds the response from the Search endpoint.
type Result struct {
	Count    int             `json:"count"`
	Quotes   []SearchQuote   `json:"quotes,omitempty"`
	News     []SearchNews    `json:"news,omitempty"`
	Explains []SearchExplain `json:"explains,omitempty"`
}

// SearchQuote represents a single quote result from a search.
type SearchQuote struct {
	Symbol         string  `json:"symbol"`
	IsYahooFinance bool    `json:"isYahooFinance"`
	Exchange       string  `json:"exchange"`
	ExchDisp       string  `json:"exchDisp,omitempty"`
	ShortName      string  `json:"shortname,omitempty"`
	LongName       string  `json:"longname,omitempty"`
	QuoteType      string  `json:"quoteType"`
	TypeDisp       string  `json:"typeDisp,omitempty"`
	Score          float64 `json:"score"`
	Sector         string  `json:"sector,omitempty"`
	Industry       string  `json:"industry,omitempty"`
}

// SearchNews represents a single news result from a search.
type SearchNews struct {
	UUID                string `json:"uuid"`
	Title               string `json:"title"`
	Publisher           string `json:"publisher"`
	Link                string `json:"link"`
	ProviderPublishTime int64  `json:"providerPublishTime"`
	Type                string `json:"type"`
}

// SearchExplain represents an explain result from a search.
type SearchExplain struct{}
