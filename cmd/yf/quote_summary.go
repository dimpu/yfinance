package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/dimpu/yfinance/quotesummary"
)

var qsModules string

var quoteSummaryCmd = &cobra.Command{
	Use:   "quoteSummary <symbol>",
	Short: "Get detailed summary data for a symbol",
	Long: `Get detailed summary data including financials, profile, statistics, etc.

Use --modules to specify which modules to return (comma-separated).
Use --modules all to return all modules.
Default modules: price, summaryDetail

Available modules: assetProfile, balanceSheetHistory, balanceSheetHistoryQuarterly,
calendarEvents, cashflowStatementHistory, cashflowStatementHistoryQuarterly,
defaultKeyStatistics, earnings, earningsHistory, earningsTrend, financialData,
fundOwnership, fundPerformance, fundProfile, incomeStatementHistory,
incomeStatementHistoryQuarterly, indexTrend, industryTrend, insiderHolders,
insiderTransactions, institutionOwnership, majorDirectHolders,
majorHoldersBreakdown, netSharePurchaseActivity, price, quoteType,
recommendationTrend, secFilings, sectorTrend, summaryDetail, summaryProfile,
topHoldings, upgradeDowngradeHistory`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := &quotesummary.Options{}
		if qsModules != "" {
			opts.Modules = strings.Split(qsModules, ",")
		}
		result, err := newClient().QuoteSummary.Get(context.Background(), args[0], opts)
		if err != nil {
			fatal("quoteSummary failed", err)
		}
		printResult(result)
	},
}

func init() {
	quoteSummaryCmd.Flags().StringVar(&qsModules, "modules", "", "comma-separated list of modules (or 'all')")
	rootCmd.AddCommand(quoteSummaryCmd)
}
