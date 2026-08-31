package quotesummary

import (
	"encoding/json"
	"fmt"

	"github.com/dimpu/yfinance/internal/types"
)

// Options specifies parameters for the QuoteSummary API call.
type Options struct {
	Formatted bool     `json:"formatted,omitempty"`
	Modules   []string `json:"modules,omitempty"`
}

var allModules = []string{
	"assetProfile",
	"balanceSheetHistory",
	"balanceSheetHistoryQuarterly",
	"calendarEvents",
	"cashflowStatementHistory",
	"cashflowStatementHistoryQuarterly",
	"defaultKeyStatistics",
	"earnings",
	"earningsHistory",
	"earningsTrend",
	"financialData",
	"fundOwnership",
	"fundPerformance",
	"fundProfile",
	"incomeStatementHistory",
	"incomeStatementHistoryQuarterly",
	"indexTrend",
	"industryTrend",
	"insiderHolders",
	"insiderTransactions",
	"institutionOwnership",
	"majorDirectHolders",
	"majorHoldersBreakdown",
	"netSharePurchaseActivity",
	"price",
	"quoteType",
	"recommendationTrend",
	"secFilings",
	"sectorTrend",
	"summaryDetail",
	"summaryProfile",
	"topHoldings",
	"upgradeDowngradeHistory",
}

// Result holds all optional sub-modules returned by the
// /v10/finance/quoteSummary endpoint.
type Result struct {
	AssetProfile                      *AssetProfile                      `json:"assetProfile,omitempty"`
	BalanceSheetHistory               *BalanceSheetHistory               `json:"balanceSheetHistory,omitempty"`
	BalanceSheetHistoryQuarterly      *BalanceSheetHistoryQuarterly      `json:"balanceSheetHistoryQuarterly,omitempty"`
	CalendarEvents                    *CalendarEvents                    `json:"calendarEvents,omitempty"`
	CashflowStatementHistory          *CashflowStatementHistory          `json:"cashflowStatementHistory,omitempty"`
	CashflowStatementHistoryQuarterly *CashflowStatementHistoryQuarterly `json:"cashflowStatementHistoryQuarterly,omitempty"`
	DefaultKeyStatistics              *DefaultKeyStatistics              `json:"defaultKeyStatistics,omitempty"`
	Earnings                          *Earnings                          `json:"earnings,omitempty"`
	EarningsHistory                   *EarningsHistory                   `json:"earningsHistory,omitempty"`
	EarningsTrend                     *EarningsTrend                     `json:"earningsTrend,omitempty"`
	FinancialData                     *FinancialData                     `json:"financialData,omitempty"`
	FundOwnership                     *FundOwnership                     `json:"fundOwnership,omitempty"`
	FundPerformance                   *FundPerformance                   `json:"fundPerformance,omitempty"`
	FundProfile                       *FundProfile                       `json:"fundProfile,omitempty"`
	IncomeStatementHistory            *IncomeStatementHistory            `json:"incomeStatementHistory,omitempty"`
	IncomeStatementHistoryQuarterly   *IncomeStatementHistoryQuarterly   `json:"incomeStatementHistoryQuarterly,omitempty"`
	IndexTrend                        *IndexTrend                        `json:"indexTrend,omitempty"`
	IndustryTrend                     *IndustryTrend                     `json:"industryTrend,omitempty"`
	InsiderHolders                    *InsiderHolders                    `json:"insiderHolders,omitempty"`
	InsiderTransactions               *InsiderTransactions               `json:"insiderTransactions,omitempty"`
	InstitutionOwnership              *InstitutionOwnership              `json:"institutionOwnership,omitempty"`
	MajorDirectHolders                *MajorDirectHolders                `json:"majorDirectHolders,omitempty"`
	MajorHoldersBreakdown             *MajorHoldersBreakdown             `json:"majorHoldersBreakdown,omitempty"`
	NetSharePurchaseActivity          *NetSharePurchaseActivity          `json:"netSharePurchaseActivity,omitempty"`
	Price                             *Price                             `json:"price,omitempty"`
	QuoteType                         *QuoteTypeSummary                  `json:"quoteType,omitempty"`
	RecommendationTrend               *RecommendationTrend               `json:"recommendationTrend,omitempty"`
	SECFilings                        *SECFilings                        `json:"secFilings,omitempty"`
	SectorTrend                       *SectorTrend                       `json:"sectorTrend,omitempty"`
	SummaryDetail                     *SummaryDetail                     `json:"summaryDetail,omitempty"`
	SummaryProfile                    *SummaryProfile                    `json:"summaryProfile,omitempty"`
	TopHoldings                       *TopHoldings                       `json:"topHoldings,omitempty"`
	UpgradeDowngradeHistory           *UpgradeDowngradeHistory           `json:"upgradeDowngradeHistory,omitempty"`
}

// CompanyOfficer represents a company officer in AssetProfile.
type CompanyOfficer struct {
	MaxAge    int64  `json:"maxAge"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	YearBorn  *int64 `json:"yearBorn,omitempty"`
	TotalPay  *struct {
		Raw     float64 `json:"raw"`
		Fmt     string  `json:"fmt"`
		LongFmt string  `json:"longFmt"`
	} `json:"totalPay,omitempty"`
	ExercisedValue *struct {
		Raw     float64 `json:"raw"`
		Fmt     string  `json:"fmt"`
		LongFmt string  `json:"longFmt"`
	} `json:"exercisedValue,omitempty"`
}

// AssetProfile provides company profile information.
type AssetProfile struct {
	MaxAge              int64            `json:"maxAge"`
	Address1            string           `json:"address1,omitempty"`
	Address2            string           `json:"address2,omitempty"`
	Address3            string           `json:"address3,omitempty"`
	City                string           `json:"city,omitempty"`
	State               string           `json:"state,omitempty"`
	Zip                 string           `json:"zip,omitempty"`
	Country             string           `json:"country,omitempty"`
	Phone               string           `json:"phone,omitempty"`
	Website             string           `json:"website,omitempty"`
	Industry            string           `json:"industry,omitempty"`
	Sector              string           `json:"sector,omitempty"`
	LongBusinessSummary string           `json:"longBusinessSummary,omitempty"`
	FullTimeEmployees   *int64           `json:"fullTimeEmployees,omitempty"`
	CompanyOfficers     []CompanyOfficer `json:"companyOfficers,omitempty"`
	AuditRisk           *json.RawMessage `json:"auditRisk,omitempty"`
	BoardRisk           *json.RawMessage `json:"boardRisk,omitempty"`
	CompensationRisk    *json.RawMessage `json:"compensationRisk,omitempty"`
	ShareHolderRightsRisk *json.RawMessage `json:"shareHolderRightsRisk,omitempty"`
	OverallRisk         *json.RawMessage `json:"overallRisk,omitempty"`
	GovernanceEpochDate *types.YahooDate `json:"governanceEpochDate,omitempty"`
	CompensationAsOfEpochDate *types.YahooDate `json:"compensationAsOfEpochDate,omitempty"`
	MaxAgeSeconds       int64            `json:"maxAgeSeconds,omitempty"`
}

// FinancialData provides financial metrics and analyst recommendations.
type FinancialData struct {
	MaxAge                     int64     `json:"maxAge"`
	CurrentPrice              *float64  `json:"currentPrice,omitempty"`
	TargetHighPrice           *float64  `json:"targetHighPrice,omitempty"`
	TargetLowPrice            *float64  `json:"targetLowPrice,omitempty"`
	TargetMeanPrice           *float64  `json:"targetMeanPrice,omitempty"`
	TargetMedianPrice         *float64  `json:"targetMedianPrice,omitempty"`
	RecommendationMean        *float64  `json:"recommendationMean,omitempty"`
	RecommendationKey         string    `json:"recommendationKey,omitempty"`
	NumberOfAnalystOpinions   *int64    `json:"numberOfAnalystOpinions,omitempty"`
	TotalCash                 *int64    `json:"totalCash,omitempty"`
	TotalCashPerShare         *float64  `json:"totalCashPerShare,omitempty"`
	TotalDebt                 *int64    `json:"totalDebt,omitempty"`
	TotalDebtToEquity         *float64  `json:"totalDebtToEquity,omitempty"`
	TotalRevenue              *int64    `json:"totalRevenue,omitempty"`
	RevenuePerShare           *float64  `json:"revenuePerShare,omitempty"`
	RevenueGrowth             *float64  `json:"revenueGrowth,omitempty"`
	Ebitda                    *int64    `json:"ebitda,omitempty"`
	GrossMargins              *float64  `json:"grossMargins,omitempty"`
	OperatingMargins          *float64  `json:"operatingMargins,omitempty"`
	ProfitMargins             *float64  `json:"profitMargins,omitempty"`
	FinancialCurrency         string    `json:"financialCurrency,omitempty"`
	CurrentPriceRaw           *float64  `json:"currentPriceRaw,omitempty"`
	TargetHighPriceRaw        *float64  `json:"targetHighPriceRaw,omitempty"`
	TargetLowPriceRaw         *float64  `json:"targetLowPriceRaw,omitempty"`
	TargetMeanPriceRaw        *float64  `json:"targetMeanPriceRaw,omitempty"`
	TargetMedianPriceRaw      *float64  `json:"targetMedianPriceRaw,omitempty"`
}

// DefaultKeyStatistics provides key valuation and statistical metrics.
type DefaultKeyStatistics struct {
	MaxAge                   int64           `json:"maxAge"`
	EnterpriseValue          *int64          `json:"enterpriseValue,omitempty"`
	ForwardPE               *float64        `json:"forwardPE,omitempty"`
	ProfitMargins            *float64        `json:"profitMargins,omitempty"`
	FloatShares              *int64          `json:"floatShares,omitempty"`
	SharesOutstanding        *int64          `json:"sharesOutstanding,omitempty"`
	Beta                    *float64        `json:"beta,omitempty"`
	BookValue               *float64        `json:"bookValue,omitempty"`
	PriceToBook             *float64        `json:"priceToBook,omitempty"`
	TrailingEps             *float64        `json:"trailingEps,omitempty"`
	ForwardEps              *float64        `json:"forwardEps,omitempty"`
	PegRatio                *float64        `json:"pegRatio,omitempty"`
	EnterpriseToRevenue     *float64        `json:"enterpriseToRevenue,omitempty"`
	EnterpriseToEbitda      *float64        `json:"enterpriseToEbitda,omitempty"`
	FiftyTwoWeekChange      *float64        `json:"52WeekChange,omitempty"`
	SandP52WeekChange       *float64        `json:"SandP52WeekChange,omitempty"`
	LastFiscalYearEnd       *types.YahooDate `json:"lastFiscalYearEnd,omitempty"`
	NextFiscalYearEnd       *types.YahooDate `json:"nextFiscalYearEnd,omitempty"`
	MostRecentQuarter       *types.YahooDate `json:"mostRecentQuarter,omitempty"`
	EarningsQuarterlyGrowth *float64        `json:"earningsQuarterlyGrowth,omitempty"`
	RevenueQuarterlyGrowth  *float64        `json:"revenueQuarterlyGrowth,omitempty"`
	HeldPercentInsiders     *float64        `json:"heldPercentInsiders,omitempty"`
	HeldPercentInstitutions *float64        `json:"heldPercentInstitutions,omitempty"`
	ShortRatio              *float64        `json:"shortRatio,omitempty"`
	ShortPercentOfFloat     *float64        `json:"shortPercentOfFloat,omitempty"`
	PercentHeldByInsiders   *float64        `json:"percentHeldByInsiders,omitempty"`
	PercentHeldByInstitutions *float64      `json:"percentHeldByInstitutions,omitempty"`
}

// Price provides price data within quote summary.
type Price struct {
	AverageDailyVolume10Day   *int64          `json:"averageDailyVolume10Day,omitempty"`
	AverageDailyVolume3Month  *int64          `json:"averageDailyVolume3Month,omitempty"`
	Exchange                  string          `json:"exchange,omitempty"`
	ExchangeName              string          `json:"exchangeName,omitempty"`
	ExchangeDataDelayedBy     int64           `json:"exchangeDataDelayedBy,omitempty"`
	MaxAge                    int64           `json:"maxAge"`
	RegularMarketPrice        *float64        `json:"regularMarketPrice,omitempty"`
	RegularMarketChange       *float64        `json:"regularMarketChange,omitempty"`
	RegularMarketChangePercent *float64       `json:"regularMarketChangePercent,omitempty"`
	RegularMarketVolume       *int64          `json:"regularMarketVolume,omitempty"`
	RegularMarketPreviousClose *float64       `json:"regularMarketPreviousClose,omitempty"`
	RegularMarketTime         *types.YahooDate `json:"regularMarketTime,omitempty"`
	RegularMarketOpen         *float64        `json:"regularMarketOpen,omitempty"`
	RegularMarketDayHigh      *float64        `json:"regularMarketDayHigh,omitempty"`
	RegularMarketDayLow       *float64        `json:"regularMarketDayLow,omitempty"`
	QuoteType                 string          `json:"quoteType,omitempty"`
	Symbol                    string          `json:"symbol,omitempty"`
	ShortName                 string          `json:"shortName,omitempty"`
	LongName                  string          `json:"longName,omitempty"`
	MarketState               string          `json:"marketState,omitempty"`
	MarketCap                 *int64          `json:"marketCap,omitempty"`
	Currency                  string          `json:"currency,omitempty"`
	CirculatingSupply         *float64        `json:"circulatingSupply,omitempty"`
	LastMarket                string          `json:"lastMarket,omitempty"`
	PriceHint                 *int64          `json:"priceHint,omitempty"`
	PreMarketPrice            *float64        `json:"preMarketPrice,omitempty"`
	PreMarketChange           *float64        `json:"preMarketChange,omitempty"`
	PreMarketChangePercent    *float64        `json:"preMarketChangePercent,omitempty"`
	PostMarketPrice           *float64        `json:"postMarketPrice,omitempty"`
	PostMarketChange          *float64        `json:"postMarketChange,omitempty"`
	PostMarketChangePercent   *float64        `json:"postMarketChangePercent,omitempty"`
}

// SummaryDetail provides key summary financial metrics.
type SummaryDetail struct {
	MaxAge                       int64           `json:"maxAge"`
	PreviousClose               *float64        `json:"previousClose,omitempty"`
	Open                        *float64        `json:"open,omitempty"`
	DayLow                      *float64        `json:"dayLow,omitempty"`
	DayHigh                     *float64        `json:"dayHigh,omitempty"`
	RegularMarketPreviousClose  *float64        `json:"regularMarketPreviousClose,omitempty"`
	RegularMarketOpen           *float64        `json:"regularMarketOpen,omitempty"`
	RegularMarketDayLow         *float64        `json:"regularMarketDayLow,omitempty"`
	RegularMarketDayHigh        *float64        `json:"regularMarketDayHigh,omitempty"`
	RegularMarketVolume         *int64          `json:"regularMarketVolume,omitempty"`
	Volume                      *int64          `json:"volume,omitempty"`
	AverageVolume               *int64          `json:"averageVolume,omitempty"`
	AverageDailyVolume10Day     *int64          `json:"averageDailyVolume10Day,omitempty"`
	AverageDailyVolume3Month    *int64          `json:"averageDailyVolume3Month,omitempty"`
	MarketCap                   *int64          `json:"marketCap,omitempty"`
	FiftyTwoWeekLow             *float64        `json:"fiftyTwoWeekLow,omitempty"`
	FiftyTwoWeekHigh            *float64        `json:"fiftyTwoWeekHigh,omitempty"`
	FiftyTwoWeekChange          *float64        `json:"fiftyTwoWeekChange,omitempty"`
	TrailingPE                  *float64        `json:"trailingPE,omitempty"`
	ForwardPE                   *float64        `json:"forwardPE,omitempty"`
	DividendRate                *float64        `json:"dividendRate,omitempty"`
	DividendYield               *float64        `json:"dividendYield,omitempty"`
	Beta                        *float64        `json:"beta,omitempty"`
	Currency                    string          `json:"currency,omitempty"`
	YtdReturn                   *float64        `json:"ytdReturn,omitempty"`
	LastFiscalYearEnd           *types.YahooDate `json:"lastFiscalYearEnd,omitempty"`
	NextFiscalYearEnd           *types.YahooDate `json:"nextFiscalYearEnd,omitempty"`
	TrailingAnnualDividendRate  *float64        `json:"trailingAnnualDividendRate,omitempty"`
	TrailingAnnualDividendYield *float64        `json:"trailingAnnualDividendYield,omitempty"`
}

// SummaryProfile provides company profile summary.
type SummaryProfile struct {
	Address1            string          `json:"address1,omitempty"`
	Address2            string          `json:"address2,omitempty"`
	Address3            string          `json:"address3,omitempty"`
	City                string          `json:"city,omitempty"`
	State               string          `json:"state,omitempty"`
	Zip                 string          `json:"zip,omitempty"`
	Country             string          `json:"country,omitempty"`
	Phone               string          `json:"phone,omitempty"`
	Website             string          `json:"website,omitempty"`
	Industry            string          `json:"industry,omitempty"`
	Sector              string          `json:"sector,omitempty"`
	LongBusinessSummary string          `json:"longBusinessSummary,omitempty"`
	FullTimeEmployees   *int64          `json:"fullTimeEmployees,omitempty"`
	MaxAge              int64           `json:"maxAge"`
	CompanyOfficers     []CompanyOfficer `json:"companyOfficers,omitempty"`
}

// RecommendationTrend provides analyst recommendation trends.
type RecommendationTrend struct {
	Trend  []RecommendationTrendItem `json:"trend,omitempty"`
	MaxAge int64                      `json:"maxAge"`
}

// RecommendationTrendItem represents a single period's recommendation counts.
type RecommendationTrendItem struct {
	Period      string `json:"period,omitempty"`
	StrongBuy   int64  `json:"strongBuy,omitempty"`
	Buy         int64  `json:"buy,omitempty"`
	Hold        int64  `json:"hold,omitempty"`
	Sell        int64  `json:"sell,omitempty"`
	StrongSell  int64  `json:"strongSell,omitempty"`
}

// EarningsChartQuarterly represents a single quarter in the earnings chart.
type EarningsChartQuarterly struct {
	Date     string   `json:"date,omitempty"`
	Actual   *float64 `json:"actual,omitempty"`
	Estimate *float64 `json:"estimate,omitempty"`
}

// FinancialsChartYearly represents a single year in the financials chart.
type FinancialsChartYearly struct {
	Date     string  `json:"date,omitempty"`
	Revenue  *int64  `json:"revenue,omitempty"`
	Earnings *int64  `json:"earnings,omitempty"`
}

// FinancialsChartQuarterly represents a single quarter in the financials chart.
type FinancialsChartQuarterly struct {
	Date     string  `json:"date,omitempty"`
	Revenue  *int64  `json:"revenue,omitempty"`
	Earnings *int64  `json:"earnings,omitempty"`
}

// Earnings provides earnings data within quote summary.
type Earnings struct {
	MaxAge            int64                  `json:"maxAge"`
	FinancialCurrency string                 `json:"financialCurrency,omitempty"`
	EarningsChart    *EarningsChart          `json:"earningsChart,omitempty"`
	FinancialsChart  *FinancialsChart        `json:"financialsChart,omitempty"`
}

// EarningsChart provides quarterly earnings chart data.
type EarningsChart struct {
	Quarterly    []EarningsChartQuarterly `json:"quarterly,omitempty"`
	EarningsDate []types.YahooDate        `json:"earningsDate,omitempty"`
}

// FinancialsChart provides yearly and quarterly financial chart data.
type FinancialsChart struct {
	Yearly    []FinancialsChartYearly    `json:"yearly,omitempty"`
	Quarterly []FinancialsChartQuarterly `json:"quarterly,omitempty"`
}

// Stub sub-module structs

type BalanceSheetHistory struct {
	MaxAge                int64 `json:"maxAge"`
	BalanceSheetStatements []struct {
		MaxAge int64 `json:"maxAge"`
	} `json:"balanceSheetStatements,omitempty"`
}

type BalanceSheetHistoryQuarterly struct {
	MaxAge                int64 `json:"maxAge"`
	BalanceSheetStatements []struct {
		MaxAge int64 `json:"maxAge"`
	} `json:"balanceSheetStatements,omitempty"`
}

type CalendarEvents struct{ MaxAge int64 `json:"maxAge"` }
type CashflowStatementHistory struct{ MaxAge int64 `json:"maxAge"` }
type CashflowStatementHistoryQuarterly struct{ MaxAge int64 `json:"maxAge"` }
type EarningsHistory struct{ MaxAge int64 `json:"maxAge"` }
type EarningsTrend struct{ MaxAge int64 `json:"maxAge"` }
type FundOwnership struct{ MaxAge int64 `json:"maxAge"` }
type FundPerformance struct{ MaxAge int64 `json:"maxAge"` }
type FundProfile struct{ MaxAge int64 `json:"maxAge"` }
type IncomeStatementHistory struct{ MaxAge int64 `json:"maxAge"` }
type IncomeStatementHistoryQuarterly struct{ MaxAge int64 `json:"maxAge"` }
type IndexTrend struct{ MaxAge int64 `json:"maxAge"` }
type IndustryTrend struct{ MaxAge int64 `json:"maxAge"` }
type InsiderHolders struct{ MaxAge int64 `json:"maxAge"` }
type InsiderTransactions struct{ MaxAge int64 `json:"maxAge"` }
type InstitutionOwnership struct{ MaxAge int64 `json:"maxAge"` }
type MajorDirectHolders struct{ MaxAge int64 `json:"maxAge"` }
type MajorHoldersBreakdown struct{ MaxAge int64 `json:"maxAge"` }
type NetSharePurchaseActivity struct{ MaxAge int64 `json:"maxAge"` }

type QuoteTypeSummary struct {
	MaxAge    int64   `json:"maxAge"`
	QuoteType string  `json:"quoteType,omitempty"`
	Exchange  string  `json:"exchange,omitempty"`
	Symbol    string  `json:"symbol,omitempty"`
	ShortName string  `json:"shortName,omitempty"`
	LongName  string  `json:"longName,omitempty"`
	MarketCap *int64  `json:"marketCap,omitempty"`
	Currency  string  `json:"currency,omitempty"`
}

type SECFilings struct{ MaxAge int64 `json:"maxAge"` }
type SectorTrend struct{ MaxAge int64 `json:"maxAge"` }
type TopHoldings struct{ MaxAge int64 `json:"maxAge"` }
type UpgradeDowngradeHistory struct{ MaxAge int64 `json:"maxAge"` }

// quoteSummaryResponse is the raw API response envelope.
type quoteSummaryResponse struct {
	QuoteSummary struct {
		Result []Result      `json:"result"`
		Error  interface{}   `json:"error"`
	} `json:"quoteSummary"`
}

// resolveModules returns the effective module list based on opts.
func resolveModules(opts *Options) []string {
	if opts == nil || len(opts.Modules) == 0 {
		return []string{"price", "summaryDetail"}
	}
	for _, m := range opts.Modules {
		if m == "all" {
			return allModules
		}
	}
	return opts.Modules
}

type quoteSummaryParseError struct {
	Msg string
}

func (e *quoteSummaryParseError) Error() string {
	return fmt.Sprintf("quoteSummary parse error: %s", e.Msg)
}
