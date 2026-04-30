package models

type BuyerDashboardPage struct {
	GreetingName      string
	DateLabel         string
	WeekRangeLabel    string
	MetricCards       []BuyerDashboardMetricCard
	HeroHighlights    []BuyerDashboardHighlight
	Stages            []BuyerDashboardStage
	POTrendLabelsJSON string
	POTrendSeriesJSON string
	Priorities        []BuyerDashboardPriorityItem
	WeeklySuppliers   []BuyerDashboardWeeklySupplier
	Suppliers         []BuyerDashboardSupplierQueue
	Sessions          []BuyerDashboardSessionQueue
}

type BuyerDashboardMetricCard struct {
	Label   string
	Value   string
	Caption string
	Icon    string
	Tone    string
}

type BuyerDashboardHighlight struct {
	Label string
	Value string
}

type BuyerDashboardStage struct {
	Label           string
	Count           int
	Note            string
	ProgressPercent int
	Tone            string
}

type BuyerDashboardMetricsSnapshot struct {
	ReviewSessionCount    int
	PendingSKUCount       int
	ApprovedSKUCount      int
	SupplierFollowUpCount int
	TodaySessionCount     int
	ActiveProductCount    int
	ActiveSupplierCount   int
	PendingValue          float64
	PendingValueDisplay   string
	ApprovedValue         float64
	ApprovedValueDisplay  string
}

type BuyerDashboardPipelineSnapshot struct {
	DraftCount    int
	WaitingCount  int
	ApprovedCount int
	RejectedCount int
}

type BuyerDashboardPriorityItem struct {
	SessionID             int
	SessionNumber         string
	SessionDateDisplay    string
	StoreName             string
	SupplierName          string
	ProductCode           string
	ProductName           string
	CheckerNotes          string
	ConditionLabel        string
	ConditionBadgeClass   string
	StatusLabel           string
	StatusBadgeClass      string
	PhysicalQtyDisplay    string
	SystemQtyDisplay      string
	SuggestBuyQtyDisplay  string
	EstimatedValueDisplay string
	LeadTimeDays          int
	DetailURL             string
}

type BuyerDashboardWeeklySupplier struct {
	SupplierID               int
	SupplierCode             string
	SupplierName             string
	StoreNames               string
	SessionCount             int
	LatestSessionDateDisplay string
	PendingItems             int
	ApprovedItems            int
	SupplierURL              string
	SessionListURL           string
}

type BuyerDashboardSupplierQueue struct {
	SupplierID            int
	SupplierCode          string
	SupplierName          string
	PendingItems          int
	OpenSessions          int
	CriticalItems         int
	EstimatedValueDisplay string
	AverageLeadTimeDays   int
	PaymentTermDays       int
	SupplierURL           string
	SessionListURL        string
}

type BuyerDashboardSessionQueue struct {
	SessionID             int
	SessionNumber         string
	SessionDateDisplay    string
	StoreName             string
	SupplierName          string
	StatusLabel           string
	StatusTextClass       string
	PendingItems          int
	ApprovedItems         int
	RejectedItems         int
	SuggestedValueDisplay string
	DetailURL             string
}
