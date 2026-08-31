package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/options"
)

var optDate string

var optionsCmd = &cobra.Command{
	Use:   "options <symbol>",
	Short: "Get options chain data for a symbol",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := &options.Options{}
		if optDate != "" {
			t, err := time.Parse("2006-01-02", optDate)
			if err != nil {
				fatal("invalid date format (use YYYY-MM-DD)", err)
			}
			opts.Date = &t
		}
		result, err := newClient().Options.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("options failed", err)
		}
		printResult(result)
	},
}

func init() {
	optionsCmd.Flags().StringVar(&optDate, "date", "", "expiration date (YYYY-MM-DD)")
	rootCmd.AddCommand(optionsCmd)
}
