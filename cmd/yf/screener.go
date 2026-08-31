package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/screener"
)

var screenerLang, screenerRegion string
var screenerCount int

var screenerCmd = &cobra.Command{
	Use:   "screener <predefined-id>",
	Short: "Get predefined screener results",
	Long: `Get results from a predefined Yahoo Finance screener.

Available screener IDs:
  aggressive_small_caps, conservative_foreign_funds, day_gainers,
  day_losers, growth_technology_stocks, high_yield_bond,
  most_actives, most_shorted_stocks, portfolio_anchors,
  small_cap_gainers, solid_large_growth_funds,
  solid_midcap_growth_funds, top_mutual_funds,
  undervalued_growth_stocks, undervalued_large_caps`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := &screener.Options{
			Lang:   screenerLang,
			Region: screenerRegion,
			Count:  screenerCount,
		}
		result, err := newClient().Screener.Get(context.Background(), screener.PredefinedScreenerModule(args[0]), opts)
		if err != nil {
			fatal("screener failed", err)
		}
		printResult(result)
	},
}

func init() {
	screenerCmd.Flags().StringVar(&screenerLang, "lang", "", "language")
	screenerCmd.Flags().StringVar(&screenerRegion, "region", "", "region")
	screenerCmd.Flags().IntVar(&screenerCount, "count", 0, "max number of results")
	rootCmd.AddCommand(screenerCmd)
}
