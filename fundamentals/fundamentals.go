package fundamentals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/fetch"
	"github.com/dimpu/yfinance/internal/validate"
)

const defaultFundamentalsQueryHost = "https://query1.finance.yahoo.com"

// Service provides access to the Yahoo Finance fundamentals API.
type Service struct {
	fetcher *fetch.Fetcher
	// queryHost override for fundamentals (different from main queryHost)
	queryHost string
}

// NewService creates a new fundamentals Service.
func NewService(f *fetch.Fetcher) *Service {
	return &Service{fetcher: f, queryHost: defaultFundamentalsQueryHost}
}

// SetQueryHost overrides the default fundamentals query host (for testing).
func (s *Service) SetQueryHost(host string) {
	s.queryHost = host
}

// Get fetches fundamental data (financials, balance sheet, cash flow)
// for a given symbol over time. This endpoint uses query1.finance.yahoo.com directly
// and does not require crumb authentication.
func (s *Service) Get(ctx context.Context, symbol string, opts *Options) ([]Result, error) {
	if symbol == "" {
		return nil, &errors.InvalidOptionsError{Field: "symbol", Msg: "symbol is required"}
	}

	if opts == nil {
		return nil, &errors.InvalidOptionsError{Field: "opts", Msg: "options are required"}
	}

	if err := validate.Struct(opts); err != nil {
		return nil, &errors.InvalidOptionsError{Field: "FundamentalsTimeSeriesOptions", Msg: err.Error()}
	}

	periodType := opts.Type
	if periodType == "" {
		periodType = "quarterly"
	}
	queryKeys := buildTimeseriesQueryKeys(periodType, opts.Module)
	if queryKeys == "" {
		return nil, &errors.InvalidOptionsError{Field: "Module", Msg: "invalid module value"}
	}

	params := url.Values{}
	params.Set("type", queryKeys)
	params.Set("period1", fmt.Sprintf("%d", opts.Period1.Unix()))

	if !opts.Period2.IsZero() {
		params.Set("period2", fmt.Sprintf("%d", opts.Period2.Unix()))
	} else {
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

	fullURL := fmt.Sprintf("%s/ws/fundamentals-timeseries/v1/finance/timeseries/%s?%s",
		s.queryHost, url.PathEscape(symbol), params.Encode())

	body, err := s.fetcher.Do(ctx, fullURL)
	if err != nil {
		return nil, err
	}

	var raw fundamentalsTimeSeriesRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing fundamentals timeseries response: %w", err)
	}

	if raw.Finance.Error != nil && raw.Finance.Error.Code != "" {
		return nil, &errors.HTTPError{StatusCode: 200, Body: raw.Finance.Error.Description}
	}

	results, err := parseFundamentalsTimeSeries(&raw, periodType)
	if err != nil {
		return nil, fmt.Errorf("processing timeseries data: %w", err)
	}

	return results, nil
}

// fundamentalsTimeSeriesRawResponse represents the raw JSON response from the API.
type fundamentalsTimeSeriesRawResponse struct {
	Finance struct {
		Result []struct {
			TimeSeries map[string]interface{} `json:"timeseries"`
			Meta       struct {
				Currency string `json:"currency"`
				Symbol   string `json:"symbol"`
				Type     string `json:"type"`
			} `json:"meta"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"finance"`
}

func parseFundamentalsTimeSeries(raw *fundamentalsTimeSeriesRawResponse, periodType string) ([]Result, error) {
	if raw == nil || len(raw.Finance.Result) == 0 {
		return nil, nil
	}

	result := raw.Finance.Result[0]
	if result.TimeSeries == nil {
		return nil, nil
	}

	dateIndices := make(map[string]int)
	var orderedDates []string

	for _, fieldData := range result.TimeSeries {
		entries, ok := fieldData.([]interface{})
		if !ok {
			continue
		}
		for _, entry := range entries {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			asOfDate, _ := entryMap["asOfDate"].(string)
			if asOfDate == "" {
				continue
			}
			if _, exists := dateIndices[asOfDate]; !exists {
				dateIndices[asOfDate] = len(orderedDates)
				orderedDates = append(orderedDates, asOfDate)
			}
		}
	}

	results := make([]Result, len(orderedDates))
	for i, dateStr := range orderedDates {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parsing date %s: %w", dateStr, err)
		}
		results[i] = Result{
			Date:       date,
			PeriodType: periodType,
			Fields:     make(map[string]*float64),
		}
	}

	for fieldName, fieldData := range result.TimeSeries {
		entries, ok := fieldData.([]interface{})
		if !ok {
			continue
		}

		cleanFieldName := strings.TrimPrefix(fieldName, periodType)
		if cleanFieldName == fieldName {
			for _, pt := range []string{"quarterly", "annual", "trailing"} {
				if strings.HasPrefix(fieldName, pt) {
					cleanFieldName = strings.TrimPrefix(fieldName, pt)
					break
				}
			}
		}

		for _, entry := range entries {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			asOfDate, _ := entryMap["asOfDate"].(string)
			if asOfDate == "" {
				continue
			}
			idx, exists := dateIndices[asOfDate]
			if !exists {
				continue
			}

			var value *float64
			if reportedValue, ok := entryMap["reportedValue"].(map[string]interface{}); ok {
				if rawVal, ok := reportedValue["raw"]; ok {
					switch v := rawVal.(type) {
					case float64:
						value = &v
					case int:
						f := float64(v)
						value = &f
					case json.Number:
						f, err := v.Float64()
						if err == nil {
							value = &f
						}
					}
				}
			}

			results[idx].Fields[cleanFieldName] = value
		}
	}

	return results, nil
}
