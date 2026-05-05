package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Search queries the Yahoo Finance search endpoint for matching quotes and news.
// No crumb authentication is required for this endpoint.
func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) (*SearchResult, error) {
	if query == "" {
		return nil, &InvalidOptionsError{Field: "query", Msg: "query is required"}
	}

	// Build URL with query params
	params := url.Values{}
	params.Set("q", query)

	// Apply defaults and options
	lang := "en-US"
	region := "US"
	if opts != nil {
		if opts.Lang != "" {
			lang = opts.Lang
		}
		if opts.Region != "" {
			region = opts.Region
		}
		if opts.QuotesCount > 0 {
			params.Set("quotesCount", fmt.Sprintf("%d", opts.QuotesCount))
		}
		if opts.NewsCount > 0 {
			params.Set("newsCount", fmt.Sprintf("%d", opts.NewsCount))
		}
	}
	params.Set("lang", lang)
	params.Set("region", region)

	fullURL := fmt.Sprintf("${YF_QUERY_HOST}/v1/finance/search?%s", params.Encode())

	// Fetch without crumb
	body, err := c.fetch(ctx, fetchConfig{url: fullURL, needsCrumb: false}, nil)
	if err != nil {
		return nil, err
	}

	// Parse response directly (no wrapping like quoteResponse)
	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	return &result, nil
}
