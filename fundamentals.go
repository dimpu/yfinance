package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const fundamentalsQueryHost = "https://query1.finance.yahoo.com"

// FundamentalsTimeSeries fetches fundamental data (financials, balance sheet, cash flow)
// for a given symbol over time. This endpoint uses query1.finance.yahoo.com directly
// and does not require crumb authentication.
func (c *Client) FundamentalsTimeSeries(ctx context.Context, symbol string, opts *FundamentalsTimeSeriesOptions) ([]FundamentalsTimeSeriesResult, error) {
	if symbol == "" {
		return nil, &InvalidOptionsError{Field: "symbol", Msg: "symbol is required"}
	}

	if opts == nil {
		return nil, &InvalidOptionsError{Field: "opts", Msg: "options are required"}
	}

	// Validate options
	if err := validateOptions(opts); err != nil {
		return nil, &InvalidOptionsError{Field: "FundamentalsTimeSeriesOptions", Msg: err.Error()}
	}

	// Build query keys
	periodType := opts.Type
	if periodType == "" {
		periodType = "quarterly"
	}
	queryKeys := buildTimeseriesQueryKeys(periodType, opts.Module)
	if queryKeys == "" {
		return nil, &InvalidOptionsError{Field: "Module", Msg: "invalid module value"}
	}

	// Build URL
	params := url.Values{}
	params.Set("type", queryKeys)
	params.Set("period1", fmt.Sprintf("%d", opts.Period1.Unix()))

	if !opts.Period2.IsZero() {
		params.Set("period2", fmt.Sprintf("%d", opts.Period2.Unix()))
	} else {
		// Default to current time if Period2 not specified
		params.Set("period2", fmt.Sprintf("%d", time.Now().Unix()))
	}

	if opts.Merge {
		params.Set("merge", "true")
	}
	if opts.PadTimeSeries {
		params.Set("padTimeSeries", "true")
	}

	lang := opts.Lang
	if lang == "" {
		lang = "en-US"
	}
	params.Set("lang", lang)

	region := opts.Region
	if region == "" {
		region = "US"
	}
	params.Set("region", region)

	// Build the URL - use the test query host if configured for testing
	queryHost := fundamentalsQueryHost
	if c.fundamentalsQueryHost != "" {
		queryHost = c.fundamentalsQueryHost
	}
	fullURL := fmt.Sprintf("%s/ws/fundamentals-timeseries/v1/finance/timeseries/%s?%s",
		queryHost, url.PathEscape(symbol), params.Encode())

	// Make direct HTTP request (no crumb needed for this endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, &BadRequestError{Message: string(body)}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	// Parse response
	var raw fundamentalsTimeSeriesRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing fundamentals timeseries response: %w", err)
	}

	// Check for API error
	if raw.Finance.Error != nil && raw.Finance.Error.Code != "" {
		return nil, &HTTPError{StatusCode: 200, Body: raw.Finance.Error.Description}
	}

	// Parse the nested structure
	results, err := parseFundamentalsTimeSeries(&raw, periodType)
	if err != nil {
		return nil, fmt.Errorf("processing timeseries data: %w", err)
	}

	return results, nil
}
