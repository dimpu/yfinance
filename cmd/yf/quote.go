package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/quote"
)

var quoteLang, quoteRegion string

var quoteCmd = &cobra.Command{
	Use:   "quote <symbol> [<symbol>...]",
	Short: "Get quote data for one or more symbols",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := &quote.Options{
			Lang:   quoteLang,
			Region: quoteRegion,
		}
		result, err := newClient().Quote.Get(context.Background(), args, opts)
		if err != nil {
			fatal("quote failed", err)
		}
		printResult(result)
	},
}

func init() {
	quoteCmd.Flags().StringVar(&quoteLang, "lang", "", "language (e.g. en-US)")
	quoteCmd.Flags().StringVar(&quoteRegion, "region", "", "region (e.g. US)")
	rootCmd.AddCommand(quoteCmd)
}
