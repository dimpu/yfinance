package yahoofinance

// InsightsOptions controls optional parameters for the Insights endpoint.
type InsightsOptions struct {
	Lang         string `json:"lang,omitempty"`
	Region       string `json:"region,omitempty"`
	ReportsCount int    `json:"reportsCount,omitempty"`
}

// InsightsResult holds the response from the Insights endpoint.
type InsightsResult struct {
	Symbol         string                    `json:"symbol"`
	InstrumentInfo *InsightsInstrumentInfo   `json:"instrumentInfo,omitempty"`
	CompanySnapshot *InsightsCompanySnapshot `json:"companySnapshot,omitempty"`
	Recommendation *InsightsRecommendation   `json:"recommendation,omitempty"`
	Events         []InsightsEvent           `json:"events,omitempty"`
	Reports        []InsightsReport          `json:"reports,omitempty"`
	SigDevs        []InsightsSigDev          `json:"sigDevs,omitempty"`
}

// InsightsInstrumentInfo contains technical and valuation data.
type InsightsInstrumentInfo struct {
	KeyTechnicals   *InsightsKeyTechnicals   `json:"keyTechnicals,omitempty"`
	TechnicalEvents *InsightsTechnicalEvents `json:"technicalEvents,omitempty"`
	Valuation       *InsightsValuation       `json:"valuation,omitempty"`
}

// InsightsKeyTechnicals contains key technical levels.
type InsightsKeyTechnicals struct {
	Provider   string   `json:"provider"`
	Support    *float64 `json:"support,omitempty"`
	Resistance *float64 `json:"resistance,omitempty"`
	StopLoss   *float64 `json:"stopLoss,omitempty"`
}

// InsightsTechnicalEvents contains technical event analysis.
type InsightsTechnicalEvents struct {
	Provider                string           `json:"provider"`
	Sector                  string           `json:"sector,omitempty"`
	ShortTermOutlook        *InsightsOutlook `json:"shortTermOutlook,omitempty"`
	IntermediateTermOutlook *InsightsOutlook `json:"intermediateTermOutlook,omitempty"`
	LongTermOutlook         *InsightsOutlook `json:"longTermOutlook,omitempty"`
}

// InsightsValuation contains valuation metrics.
type InsightsValuation struct {
	Color         string   `json:"color,omitempty"`
	Description   string   `json:"description,omitempty"`
	Discount      *float64 `json:"discount,omitempty"`
	Provider      string   `json:"provider"`
	RelativeValue *float64 `json:"relativeValue,omitempty"`
}

// InsightsCompanySnapshot contains company and sector info.
type InsightsCompanySnapshot struct {
	SectorInfo string              `json:"sectorInfo,omitempty"`
	Company    *InsightsCompany    `json:"company,omitempty"`
	Sector     *InsightsSectorInfo `json:"sector,omitempty"`
}

// InsightsCompany contains company metrics.
type InsightsCompany struct {
	Innovativeness    *float64 `json:"innovativeness,omitempty"`
	Hiring            *float64 `json:"hiring,omitempty"`
	Sustainability    *float64 `json:"sustainability,omitempty"`
	InsiderSentiments *float64 `json:"insiderSentiments,omitempty"`
	EarningsReports   *float64 `json:"earningsReports,omitempty"`
	Dividends         *float64 `json:"dividends,omitempty"`
}

// InsightsSectorInfo contains sector metrics.
type InsightsSectorInfo struct {
	Innovativeness    float64  `json:"innovativeness"`
	Hiring            float64  `json:"hiring"`
	Sustainability    *float64 `json:"sustainability,omitempty"`
	InsiderSentiments float64  `json:"insiderSentiments"`
	EarningsReports   *float64 `json:"earningsReports,omitempty"`
	Dividends         *float64 `json:"dividends,omitempty"`
}

// InsightsRecommendation contains analyst recommendation.
type InsightsRecommendation struct {
	TargetPrice *float64 `json:"targetPrice,omitempty"`
	Provider    string   `json:"provider"`
	Rating      string   `json:"rating"` // "BUY", "SELL", "HOLD"
}

// InsightsOutlook contains outlook analysis.
type InsightsOutlook struct {
	StateDescription string   `json:"stateDescription"`
	Direction        string   `json:"direction"` // "Bearish", "Bullish", "Neutral"
	Score            *float64 `json:"score,omitempty"`
	ScoreDescription string   `json:"scoreDescription,omitempty"`
	SectorDirection  string   `json:"sectorDirection,omitempty"`
	SectorScore      *float64 `json:"sectorScore,omitempty"`
	IndexDirection   string   `json:"indexDirection,omitempty"`
	IndexScore       *float64 `json:"indexScore,omitempty"`
}

// InsightsEvent represents an event (structure TBD).
type InsightsEvent struct{}

// InsightsReport represents a research report.
type InsightsReport struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Provider    string `json:"provider"`
	PublishedAt int64  `json:"publishedAt"`
}

// InsightsSigDev represents a significant development.
type InsightsSigDev struct {
	Headline string `json:"headline"`
	Date     int64  `json:"date"`
}
