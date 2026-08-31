package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/insights"
)

var insightsLang, insightsRegion string
var insightsReports int

var insightsCmd = &cobra.Command{
	Use:   "insights <symbol>",
	Short: "Get insights data for a symbol",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := &insights.Options{
			Lang:         insightsLang,
			Region:       insightsRegion,
			ReportsCount: insightsReports,
		}
		result, err := newClient().Insights.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("insights failed", err)
		}
		printResult(result)
	},
}

func init() {
	insightsCmd.Flags().StringVar(&insightsLang, "lang", "", "language")
	insightsCmd.Flags().StringVar(&insightsRegion, "region", "", "region")
	insightsCmd.Flags().IntVar(&insightsReports, "reports", 0, "number of reports")
	rootCmd.AddCommand(insightsCmd)
}
