package models

type StockCheckSession struct {
	ID                       int
	SessionNumber            string
	SessionDate              string
	SessionDateDisplay       string
	SessionDateMonthShort    string
	SessionDateDay           string
	SessionDateYear          string
	StoreID                  int
	StoreName                string
	SupplierID               int
	SupplierCode             string
	SupplierName             string
	InitiationType           string
	InitiationTypeLabel      string
	InitiationTypeBadgeClass string
	Status                   string
	StatusLabel              string
	StatusTextClass          string
	StatusDotClass           string
	CreatedBy                int
	CreatedByName            string
	CreatedByInitials        string
	CreatedByAvatarClass     string
	Notes                    string
	CreatedAt                string
	CreatedAtDisplay         string
}

type StockCheckSessionListFilter struct {
	DateFrom   string
	DateTo     string
	StoreID    int
	SupplierID int
	Status     string
	Page       int
	Limit      int
}

type StockCheckSessionCreateInput struct {
	SessionNumber  string
	SessionDate    string
	StoreID        int
	SupplierID     int
	InitiationType string
	Status         string
	Notes          string
	CreatedBy      int
}

type StockCheckSessionCheckerScanInput struct {
	SessionID int
	Location  string
	Barcode   string
	Qty       float64
	UpdatedBy int
}

type StockCheckSessionUpdateInput struct {
	ID             int
	SessionDate    string
	StoreID        int
	SupplierID     int
	InitiationType string
	Status         string
	Notes          string
}

type StockCheckSessionReviewItemUpdateInput struct {
	SessionID      int
	ItemID         int
	ApprovedBuyQty float64
	BuyerNotes     string
	Status         string
	ReviewedBy     int
	UpdatedBy      int
}

type StockCheckSessionReviewItemEditForm struct {
	ItemID         int
	ApprovedBuyQty string
	BuyerNotes     string
}

type StockCheckSessionCheckerScanForm struct {
	Location string
	Barcode  string
	Qty      string
}

type StockCheckSessionDetailPage struct {
	Session       StockCheckSessionDetail
	Items         []StockCheckSessionReviewItem
	OverviewCards []StockCheckSessionOverviewCard
	Pagination    Pagination
}

type StockCheckSessionCheckerInputPage struct {
	Session StockCheckSession
	Items   []StockCheckSessionCheckerInputItem
}

type StockCheckSessionDetail struct {
	StockCheckSession
	StageLabel                    string
	StatusBadgeClass              string
	SuggestedPurchaseValue        float64
	SuggestedPurchaseValueDisplay string
	FinalApprovedValue            float64
	FinalApprovedValueDisplay     string
	ApprovalYieldPercent          float64
	ApprovalYieldDisplay          string
	ItemCount                     int
	ApprovedItems                 int
	RejectedItems                 int
	OnHoldItems                   int
	TotalSuggestedQty             float64
	TotalSuggestedQtyDisplay      string
	TotalApprovedQty              float64
	TotalApprovedQtyDisplay       string
	DistinctSupplierCount         int
}

type StockCheckSessionReviewItem struct {
	ID                       int
	ProductID                int
	ProductCode              string
	ProductName              string
	Brand                    string
	UnitName                 string
	ProductInitials          string
	ProductAvatarClass       string
	QtyStore                 float64
	QtyStoreDisplay          string
	QtyWarehouse             float64
	QtyWarehouseDisplay      string
	TotalQty                 float64
	TotalQtyDisplay          string
	SystemTotalQty           float64
	SystemTotalQtyDisplay    string
	SuggestBuyQty            float64
	SuggestBuyQtyDisplay     string
	ApprovedBuyQty           float64
	ApprovedBuyQtyDisplay    string
	SelectedSupplierName     string
	CheckerNotes             string
	BuyerNotes               string
	ConditionStatus          string
	ConditionLabel           string
	ConditionBadgeClass      string
	BuyerNoteAccentClass     string
	Status                   string
	StatusLabel              string
	StatusBadgeClass         string
	SuggestLineValue         float64
	SuggestLineValueDisplay  string
	ApprovedLineValue        float64
	ApprovedLineValueDisplay string
}

type StockCheckSessionCheckerInputItem struct {
	ID                  int
	ProductID           int
	ProductCode         string
	Barcode             string
	ProductName         string
	CategoryName        string
	UnitName            string
	QtyStore            float64
	QtyStoreDisplay     string
	QtyWarehouse        float64
	QtyWarehouseDisplay string
	TotalQty            float64
	TotalQtyDisplay     string
	Status              string
	StatusLabel         string
	HasBarcode          bool
}

type StockCheckSessionOverviewCard struct {
	Title         string
	Description   string
	Icon          string
	IconWrapClass string
}
