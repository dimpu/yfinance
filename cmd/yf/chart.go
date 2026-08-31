package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/chart"
)

var chartInterval, chartLang, chartEvents string
var chartPeriod1, chartPeriod2 string
var chartPrePost bool

var chartCmd = &cobra.Command{
	Use:   "chart <symbol>",
	Short: "Get chart data for a symbol",
	Long: `Get chart/candlestick data for a symbol.

Intervals: 1m, 2m, 5m, 15m, 30m, 60m, 90m, 1h, 1d, 5d, 1wk, 1mo, 3mo

Period dates in YYYY-MM-DD format. Defaults: period1=1 year ago, period2=now.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p1 := parseDate(chartPeriod1, time.Now().AddDate(-1, 0, 0))
		p2 := parseDate(chartPeriod2, time.Now())

		opts := &chart.Options{
			Period1:        p1,
			Period2:        p2,
			Interval:       chartInterval,
			IncludePrePost: chartPrePost,
			Events:         chartEvents,
			Lang:           chartLang,
		}
		result, err := newClient().Chart.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("chart failed", err)
		}
		printResult(result)
	},
}

func init() {
	chartCmd.Flags().StringVar(&chartInterval, "interval", "1d", "data interval")
	chartCmd.Flags().StringVar(&chartPeriod1, "period1", "", "start date (YYYY-MM-DD)")
	chartCmd.Flags().StringVar(&chartPeriod2, "period2", "", "end date (YYYY-MM-DD)")
	chartCmd.Flags().BoolVar(&chartPrePost, "prepost", false, "include pre/post market data")
	chartCmd.Flags().StringVar(&chartEvents, "events", "", "event types to include")
	chartCmd.Flags().StringVar(&chartLang, "lang", "", "language")
	rootCmd.AddCommand(chartCmd)
}

func parseDate(s string, fallback time.Time) time.Time {
	if s == "" {
		return fallback
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		fatal("invalid date format (use YYYY-MM-DD)", err)
	}
	return t
}
