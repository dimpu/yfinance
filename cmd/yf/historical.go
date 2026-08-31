package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/historical"
)

var histInterval, histEvents string
var histPeriod1, histPeriod2 string
var histAdjClose bool

var historicalCmd = &cobra.Command{
	Use:   "historical <symbol>",
	Short: "Get historical price data",
	Long: `Get historical price data for a symbol.

Subcommands: dividends, splits
Intervals: 1d, 1wk, 1mo
Events: history, dividends, split

Period dates in YYYY-MM-DD format. Defaults: period1=1 year ago, period2=now.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p1 := parseDate(histPeriod1, time.Now().AddDate(-1, 0, 0))
		p2 := parseDate(histPeriod2, time.Now())

		opts := &historical.Options{
			Period1:              p1,
			Period2:              p2,
			Interval:             histInterval,
			Events:               histEvents,
			IncludeAdjustedClose: histAdjClose,
		}
		result, err := newClient().Historical.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("historical failed", err)
		}
		printResult(result)
	},
}

var dividendsCmd = &cobra.Command{
	Use:   "dividends <symbol>",
	Short: "Get historical dividend data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p1 := parseDate(histPeriod1, time.Now().AddDate(-1, 0, 0))
		p2 := parseDate(histPeriod2, time.Now())

		opts := &historical.Options{
			Period1: p1,
			Period2: p2,
		}
		result, err := newClient().Historical.Dividends(context.Background(), args[0], opts)
		if err != nil {
			fatal("dividends failed", err)
		}
		printResult(result)
	},
}

var splitsCmd = &cobra.Command{
	Use:   "splits <symbol>",
	Short: "Get historical stock split data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p1 := parseDate(histPeriod1, time.Now().AddDate(-1, 0, 0))
		p2 := parseDate(histPeriod2, time.Now())

		opts := &historical.Options{
			Period1: p1,
			Period2: p2,
		}
		result, err := newClient().Historical.Splits(context.Background(), args[0], opts)
		if err != nil {
			fatal("splits failed", err)
		}
		printResult(result)
	},
}

func init() {
	historicalCmd.Flags().StringVar(&histInterval, "interval", "1d", "data interval (1d, 1wk, 1mo)")
	historicalCmd.Flags().StringVar(&histPeriod1, "period1", "", "start date (YYYY-MM-DD)")
	historicalCmd.Flags().StringVar(&histPeriod2, "period2", "", "end date (YYYY-MM-DD)")
	historicalCmd.Flags().StringVar(&histEvents, "events", "history", "event type (history, dividends, split)")
	historicalCmd.Flags().BoolVar(&histAdjClose, "adjclose", true, "include adjusted close")

	dividendsCmd.Flags().StringVar(&histPeriod1, "period1", "", "start date (YYYY-MM-DD)")
	dividendsCmd.Flags().StringVar(&histPeriod2, "period2", "", "end date (YYYY-MM-DD)")
	splitsCmd.Flags().StringVar(&histPeriod1, "period1", "", "start date (YYYY-MM-DD)")
	splitsCmd.Flags().StringVar(&histPeriod2, "period2", "", "end date (YYYY-MM-DD)")

	historicalCmd.AddCommand(dividendsCmd)
	historicalCmd.AddCommand(splitsCmd)
	rootCmd.AddCommand(historicalCmd)
}
