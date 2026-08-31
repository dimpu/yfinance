package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/recommendations"
)

var recommendationsCmd = &cobra.Command{
	Use:   "recommendations <symbol> [<symbol>...]",
	Short: "Get recommended symbols for one or more symbols",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := newClient().Recommendations.Get(context.Background(), args, &recommendations.Options{})
		if err != nil {
			fatal("recommendations failed", err)
		}
		printResult(result)
	},
}

func init() {
	rootCmd.AddCommand(recommendationsCmd)
}
