package insights

// Options controls optional parameters for the Insights endpoint.
type Options struct {
	Lang         string `json:"lang,omitempty"`
	Region       string `json:"region,omitempty"`
	ReportsCount int    `json:"reportsCount,omitempty"`
}

// Result holds the response from the Insights endpoint.
type Result struct {
	Symbol          string             `json:"symbol"`
	InstrumentInfo  *InstrumentInfo    `json:"instrumentInfo,omitempty"`
	CompanySnapshot *CompanySnapshot   `json:"companySnapshot,omitempty"`
	Recommendation  *Recommendation    `json:"recommendation,omitempty"`
	Events          []Event            `json:"events,omitempty"`
	Reports         []Report           `json:"reports,omitempty"`
	SigDevs         []SigDev           `json:"sigDevs,omitempty"`
}

// InstrumentInfo contains technical and valuation data.
type InstrumentInfo struct {
	KeyTechnicals   *KeyTechnicals   `json:"keyTechnicals,omitempty"`
	TechnicalEvents *TechnicalEvents `json:"technicalEvents,omitempty"`
	Valuation       *Valuation       `json:"valuation,omitempty"`
}

// KeyTechnicals contains key technical levels.
type KeyTechnicals struct {
	Provider   string   `json:"provider"`
	Support    *float64 `json:"support,omitempty"`
	Resistance *float64 `json:"resistance,omitempty"`
	StopLoss   *float64 `json:"stopLoss,omitempty"`
}

// TechnicalEvents contains technical event analysis.
type TechnicalEvents struct {
	Provider                string    `json:"provider"`
	Sector                  string    `json:"sector,omitempty"`
	ShortTermOutlook        *Outlook  `json:"shortTermOutlook,omitempty"`
	IntermediateTermOutlook *Outlook  `json:"intermediateTermOutlook,omitempty"`
	LongTermOutlook         *Outlook  `json:"longTermOutlook,omitempty"`
}

// Valuation contains valuation metrics.
type Valuation struct {
	Color         string   `json:"color,omitempty"`
	Description   string   `json:"description,omitempty"`
	Discount      *float64 `json:"discount,omitempty"`
	Provider      string   `json:"provider"`
	RelativeValue *float64 `json:"relativeValue,omitempty"`
}

// CompanySnapshot contains company and sector info.
type CompanySnapshot struct {
	SectorInfo string       `json:"sectorInfo,omitempty"`
	Company    *Company     `json:"company,omitempty"`
	Sector     *SectorInfo  `json:"sector,omitempty"`
}

// Company contains company metrics.
type Company struct {
	Innovativeness    *float64 `json:"innovativeness,omitempty"`
	Hiring            *float64 `json:"hiring,omitempty"`
	Sustainability    *float64 `json:"sustainability,omitempty"`
	InsiderSentiments *float64 `json:"insiderSentiments,omitempty"`
	EarningsReports   *float64 `json:"earningsReports,omitempty"`
	Dividends         *float64 `json:"dividends,omitempty"`
}

// SectorInfo contains sector metrics.
type SectorInfo struct {
	Innovativeness    float64  `json:"innovativeness"`
	Hiring            float64  `json:"hiring"`
	Sustainability    *float64 `json:"sustainability,omitempty"`
	InsiderSentiments float64  `json:"insiderSentiments"`
	EarningsReports   *float64 `json:"earningsReports,omitempty"`
	Dividends         *float64 `json:"dividends,omitempty"`
}

// Recommendation contains analyst recommendation.
type Recommendation struct {
	TargetPrice *float64 `json:"targetPrice,omitempty"`
	Provider    string   `json:"provider"`
	Rating      string   `json:"rating"`
}

// Outlook contains outlook analysis.
type Outlook struct {
	StateDescription string   `json:"stateDescription"`
	Direction        string   `json:"direction"`
	Score            *float64 `json:"score,omitempty"`
	ScoreDescription string   `json:"scoreDescription,omitempty"`
	SectorDirection  string   `json:"sectorDirection,omitempty"`
	SectorScore      *float64 `json:"sectorScore,omitempty"`
	IndexDirection   string   `json:"indexDirection,omitempty"`
	IndexScore       *float64 `json:"indexScore,omitempty"`
}

// Event represents an event.
type Event struct{}

// Report represents a research report.
type Report struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Provider    string `json:"provider"`
	PublishedAt int64  `json:"publishedAt"`
}

// SigDev represents a significant development.
type SigDev struct {
	Headline string `json:"headline"`
	Date     int64  `json:"date"`
}
