package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/fundamentals"
)

var fundType, fundModule, fundLang, fundRegion string
var fundPeriod1, fundPeriod2 string
var fundMerge, fundPad bool

var fundamentalsCmd = &cobra.Command{
	Use:   "fundamentals <symbol>",
	Short: "Get financial statement time series data",
	Long: `Get fundamentals/financial statement time series data.

Types: quarterly, annual, trailing
Modules: financials, balance-sheet, cash-flow, all

Period dates in YYYY-MM-DD format. Defaults: period1=1 year ago, period2=now.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p1 := parseDate(fundPeriod1, time.Now().AddDate(-1, 0, 0))
		p2 := parseDate(fundPeriod2, time.Now())

		opts := &fundamentals.Options{
			Period1:       p1,
			Period2:       p2,
			Type:          fundType,
			Merge:         fundMerge,
			PadTimeSeries: fundPad,
			Lang:          fundLang,
			Region:        fundRegion,
			Module:        fundModule,
		}
		result, err := newClient().Fundamentals.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("fundamentals failed", err)
		}
		printResult(result)
	},
}

func init() {
	fundamentalsCmd.Flags().StringVar(&fundType, "type", "annual", "period type (quarterly, annual, trailing)")
	fundamentalsCmd.Flags().StringVar(&fundModule, "module", "all", "module (financials, balance-sheet, cash-flow, all)")
	fundamentalsCmd.Flags().StringVar(&fundPeriod1, "period1", "", "start date (YYYY-MM-DD)")
	fundamentalsCmd.Flags().StringVar(&fundPeriod2, "period2", "", "end date (YYYY-MM-DD)")
	fundamentalsCmd.Flags().StringVar(&fundLang, "lang", "", "language")
	fundamentalsCmd.Flags().StringVar(&fundRegion, "region", "", "region")
	fundamentalsCmd.Flags().BoolVar(&fundMerge, "merge", false, "merge results")
	fundamentalsCmd.Flags().BoolVar(&fundPad, "pad", false, "pad time series")
	rootCmd.AddCommand(fundamentalsCmd)
}
