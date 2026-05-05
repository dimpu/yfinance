package yahoofinance

// timeseriesFinancialsKeys contains the field names for the financials module.
// These are combined with periodType (quarterly/annual/trailing) to form query keys.
var timeseriesFinancialsKeys = []string{
	"TotalRevenue",
	"OperatingRevenue",
	"CostOfRevenue",
	"GrossProfit",
	"ResearchAndDevelopment",
	"SellingGeneralAndAdministration",
	"OperatingExpense",
	"OperatingIncome",
	"InterestExpenseNonOperating",
	"NetNonOperatingInterestIncomeExpense",
	"TotalExpenses",
	"PretaxIncome",
	"TaxProvision",
	"NetIncomeContinuousOperations",
	"NetIncome",
	"BasicAverageShares",
	"DilutedAverageShares",
	"BasicEPS",
	"DilutedEPS",
	"EBIT",
	"EBITDA",
	"NormalizedEBITDA",
	"NetIncomeFromContinuingOperationNetMinorityInterest",
	"TotalUnusualItems",
	"TotalUnusualItemsExcludingGoodwill",
	"NormalizedIncome",
	"ReconciledCostOfRevenue",
	"ReconciledDepreciation",
	"TaxRateForCalcs",
	"RentExpenseSupplemental",
	"InterestIncome",
	"InterestExpense",
	"OtherIncomeExpense",
	"OtherNonOperatingIncomeExpense",
	"SpecialIncomeCharges",
	"GainOnSaleOfPPE",
	"GainOnSaleOfBusiness",
	"OtherSpecialCharges",
	"ImpairmentOfCapitalAssets",
	"RestructuringCharges",
	"SecuritiesAmortization",
	"EarningsFromEquityInterestNetOfTax",
	"TaxesRefundsPaid",
	"TaxEffectOfUnusualItems",
	"NormalizedBasicEPS",
	"NormalizedDilutedEPS",
}

// timeseriesBalanceSheetKeys contains the field names for the balance-sheet module.
var timeseriesBalanceSheetKeys = []string{
	"NetDebt",
	"TotalDebt",
	"CashAndCashEquivalents",
	"TotalAssets",
	"TotalLiabilitiesNetMinorityInterest",
	"TotalEquityGrossMinorityInterest",
	"StockholdersEquity",
	"CommonStockEquity",
	"LongTermDebt",
	"CurrentDebt",
	"CurrentAssets",
	"CurrentLiabilities",
	"Inventory",
	"AccountsReceivable",
	"AccountsPayable",
	"RetainedEarnings",
	"AdditionalPaidInCapital",
	"NetPPE",
	"Goodwill",
	"OtherIntangibleAssets",
	"WorkingCapital",
	"InvestedCapital",
	"TotalLiabilities",
	"TotalNonCurrentAssets",
	"TotalNonCurrentLiabilities",
	"OtherCurrentAssets",
	"OtherCurrentLiabilities",
	"OtherNonCurrentAssets",
	"OtherNonCurrentLiabilities",
	"TreasuryStock",
	"CapitalStock",
	"PropertyPlantEquipment",
	"AccumulatedDepreciation",
	"GrossPPE",
	"CashEquivalents",
	"CashFinancial",
	"Receivables",
	"OtherReceivables",
	"InventoryRawMaterials",
	"InventoryWorkInProcess",
	"InventoryFinishedGoods",
	"DeferredTaxAssets",
	"DeferredTaxLiabilities",
	"LongTermProvisions",
	"ShortLongTermDebt",
	"LongTermCapitalLeaseObligation",
	"CurrentCapitalLeaseObligation",
	"PensionFundObligations",
	"SecuritiesReceivable",
	"OtherShortTermInvestments",
	"TradingSecurities",
	"AvtloAssets",
}

// timeseriesCashFlowKeys contains the field names for the cash-flow module.
var timeseriesCashFlowKeys = []string{
	"FreeCashFlow",
	"OperatingCashFlow",
	"CapitalExpenditure",
	"RepurchaseOfCapitalStock",
	"RepaymentOfDebt",
	"IssuanceOfDebt",
	"IssuanceOfCapitalStock",
	"CashDividendsPaid",
	"ChangeInWorkingCapital",
	"ChangeInReceivables",
	"ChangeInInventory",
	"DepreciationAndAmortization",
	"StockBasedCompensation",
	"NetIncomeFromContinuingOperations",
	"DeferredTax",
	"OtherNonCashItems",
	"EndCashPosition",
	"BeginningCashPosition",
	"ChangesInCash",
	"EffectOfExchangeRateChanges",
	"CashFlowFromDiscontinuedOperation",
	"CashFlowFromFinancing",
	"CashFlowFromInvesting",
	"CashFlowFromOperating",
	"NetOtherInvestingChanges",
	"NetOtherFinancingCharges",
	"NetBusinessPurchaseAndSale",
	"PensionAndEmployeeBenefitExpense",
	"DeferredIncomeTax",
	"OtherWorkingCapital",
	"OtherCashAdjustmentInside",
	"OtherCashAdjustmentOutside",
	"InterestPaidCFF",
	"IncomeTaxPaidCFF",
	"IncomeTaxPaidSupplemental",
	"InterestPaidSupplemental",
	"NetIntangiblesPurchaseAndSale",
	"NetPP EPurchaseAndSale",
	"NetInvestmentPurchaseAndSale",
	"PurchaseOfBusiness",
	"PurchaseOfInvestment",
	"SaleOfInvestment",
	"RepurchaseOfCapitalStock",
	"PreferredStockIssuance",
	"PreferredStockRepurchase",
	"CommonStockIssuance",
	"CommonStockRepurchase",
	"NetLongTermDebtIssuance",
	"LongTermDebtPayments",
	"NetShortTermDebtIssuance",
	"ShortTermDebtPayments",
}

// buildTimeseriesQueryKeys builds the type query parameter by combining
// periodType with the field names for the given module.
func buildTimeseriesQueryKeys(periodType string, module string) string {
	var keys []string
	switch module {
	case "financials":
		keys = timeseriesFinancialsKeys
	case "balance-sheet":
		keys = timeseriesBalanceSheetKeys
	case "cash-flow":
		keys = timeseriesCashFlowKeys
	case "all":
		keys = append(timeseriesFinancialsKeys, timeseriesBalanceSheetKeys...)
		keys = append(keys, timeseriesCashFlowKeys...)
	default:
		return ""
	}

	result := make([]string, len(keys))
	for i, k := range keys {
		result[i] = periodType + k
	}
	return joinStrings(result, ",")
}

// joinStrings joins a slice of strings with a separator.
// This is a simple helper to avoid importing strings for just one use.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
