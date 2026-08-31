package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/search"
)

var searchLang, searchRegion string
var searchQuotesCount, searchNewsCount int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for quotes and news",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := &search.Options{
			Lang:        searchLang,
			Region:      searchRegion,
			QuotesCount: searchQuotesCount,
			NewsCount:   searchNewsCount,
		}
		result, err := newClient().Search.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("search failed", err)
		}
		printResult(result)
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchLang, "lang", "", "language (e.g. en-US)")
	searchCmd.Flags().StringVar(&searchRegion, "region", "", "region (e.g. US)")
	searchCmd.Flags().IntVar(&searchQuotesCount, "quotes", 0, "max number of quotes")
	searchCmd.Flags().IntVar(&searchNewsCount, "news", 0, "max number of news items")
	rootCmd.AddCommand(searchCmd)
}
