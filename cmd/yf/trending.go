package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/trending"
)

var trendingLang string
var trendingCount int

var trendingCmd = &cobra.Command{
	Use:   "trending <region>",
	Short: "Get trending symbols for a region",
	Long: `Get trending symbols for a region (e.g. US, GB, DE, FR, JP, AU).

Common regions: US, GB, DE, FR, JP, AU, CA, HK, TW, IN, KR`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := &trending.Options{
			Lang:   trendingLang,
			Region: args[0],
			Count:  trendingCount,
		}
		result, err := newClient().Trending.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("trending failed", err)
		}
		printResult(result)
	},
}

func init() {
	trendingCmd.Flags().StringVar(&trendingLang, "lang", "", "language")
	trendingCmd.Flags().IntVar(&trendingCount, "count", 0, "max number of results")
	rootCmd.AddCommand(trendingCmd)
}
