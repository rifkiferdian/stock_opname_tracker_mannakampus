package models

type StockOpnameReportFilter struct {
	Status string
	Page   int
	Limit  int
}

type StockOpnameReportPage struct {
	Filter                 StockOpnameReportFilter
	StatusOptions          []StockOpnameReportStatusOption
	CurrentStatusLabel     string
	CurrentDateLabel       string
	HistoryDateLabels      []string
	SummaryCards           []StockOpnameReportSummaryCard
	ReportRows             []StockOpnameReportRow
	TrendBars              []StockOpnameReportTrendBar
	AuditFindings          []StockOpnameReportAuditFinding
	Pagination             Pagination
	TotalRows              int
	LatestSessionCount     int
	ApplyAllSubmittedCount int
	ReviewListURL          string
	ExportURL              string
}

type StockOpnameReportStatusOption struct {
	Value string
	Label string
}

type StockOpnameReportSummaryCard struct {
	Label string
	Value string
	Note  string
	Tone  string
}

type StockOpnameReportHistoryPoint struct {
	DateLabel        string
	ShopCount        string
	ShopBreakdown    string
	ShopCarton       int
	ShopBox          int
	ShopPcs          int
	WHSCount         string
	WHSBreakdown     string
	WHSCarton        int
	WHSBox           int
	WHSPcs           int
	POCount          string
	POBreakdown      string
	POCarton         int
	POBox            int
	POPcs            int
	SuggestCarton    int
	SuggestBox       int
	SuggestPcs       int
	SuggestBreakdown string
}

type StockOpnameReportRow struct {
	ProductID               int
	Name                    string
	SKU                     string
	Barcode                 string
	Brand                   string
	CategoryName            string
	LeadTimeLabel           string
	AvatarText              string
	AvatarClass             string
	CurrentShop             string
	CurrentShopBreakdown    string
	CurrentShopCarton       int
	CurrentShopBox          int
	CurrentShopPcs          int
	CurrentWH               string
	CurrentWHBreakdown      string
	CurrentWHCarton         int
	CurrentWHBox            int
	CurrentWHPcs            int
	CurrentPO               string
	CurrentPOBreakdown      string
	CurrentPOCarton         int
	CurrentPOBox            int
	CurrentPOPcs            int
	CurrentStatus           string
	StatusTone              string
	CurrentSessionID        int
	CurrentItemID           int
	CurrentSuggestQty       string
	CurrentSuggestBreakdown string
	CurrentSuggestCartonRaw int
	CurrentSuggestCarton    int
	CurrentSuggestBox       int
	CurrentSuggestPcs       int
	CurrentCheckerNote      string
	CurrentBuyerNotes       string
	CurrentApproveSeed      string
	CanInlineApprove        bool
	History                 []StockOpnameReportHistoryPoint
	ActionLabel             string
	ActionURL               string
	LatestSessionID         int
	LatestSessionDate       string
	POTrendSeriesJSON       string
	POTrendLabelsJSON       string
	POTrendTotalLabel       string
}

type StockOpnameReportTrendBar struct {
	Label     string
	Height    string
	IsCurrent bool
	Value     string
}

type StockOpnameReportAuditFinding struct {
	Title     string
	SKU       string
	Icon      string
	ActionURL string
}
