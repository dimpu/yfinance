package yahoofinance

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// FundamentalsTimeSeriesOptions contains options for the FundamentalsTimeSeries API.
type FundamentalsTimeSeriesOptions struct {
	Period1       time.Time `validate:"required"`
	Period2       time.Time
	Type          string `validate:"omitempty,oneof=quarterly annual trailing"` // default "quarterly"
	Merge         bool
	PadTimeSeries bool
	Lang          string
	Region        string
	Module        string `validate:"required,oneof=financials balance-sheet cash-flow all"`
}

// FundamentalsTimeSeriesResult represents a single time series data point.
// Since there are 145+ financials fields, 196+ balance sheet fields, and 148+ cash flow fields,
// we use a generic map[string]*float64 for Fields rather than individual struct fields.
type FundamentalsTimeSeriesResult struct {
	Date       time.Time           `json:"date"`
	Type       string              `json:"TYPE"`
	PeriodType string              `json:"periodType"`
	Fields     map[string]*float64 `json:"-"`
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

// parseFundamentalsTimeSeries parses the raw response into structured results.
// The response format is:
//
//	{
//	  "finance": {
//	    "result": [{
//	      "timeseries": {
//	        "quarterlyTotalRevenue": [{
//	          "asOfDate": "2023-09-30",
//	          "periodType": "3M",
//	          "reportedValue": { "raw": 89498000000 }
//	        }, ...]
//	      }
//	    }]
//	  }
//	}
func parseFundamentalsTimeSeries(raw *fundamentalsTimeSeriesRawResponse, periodType string) ([]FundamentalsTimeSeriesResult, error) {
	if raw == nil || len(raw.Finance.Result) == 0 {
		return nil, nil
	}

	result := raw.Finance.Result[0]
	if result.TimeSeries == nil {
		return nil, nil
	}

	// First pass: collect all unique dates and build date -> index mapping
	dateIndices := make(map[string]int)
	var orderedDates []string

	// Iterate through all fields to find all unique dates
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

	// Initialize results slice
	results := make([]FundamentalsTimeSeriesResult, len(orderedDates))
	for i, dateStr := range orderedDates {
		date, err := parseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("parsing date %s: %w", dateStr, err)
		}
		results[i] = FundamentalsTimeSeriesResult{
			Date:       date,
			PeriodType: periodType,
			Fields:     make(map[string]*float64),
		}
	}

	// Second pass: populate fields
	for fieldName, fieldData := range result.TimeSeries {
		entries, ok := fieldData.([]interface{})
		if !ok {
			continue
		}

		// Strip the periodType prefix to get the clean field name
		cleanFieldName := strings.TrimPrefix(fieldName, periodType)
		if cleanFieldName == fieldName {
			// If prefix wasn't present, try other period types
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

			// Extract reportedValue.raw
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

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
