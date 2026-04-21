package models

type Product struct {
	ID                          int
	ProductCode                 string
	Barcode                     string
	ProductName                 string
	CategoryID                  int
	CategoryName                string
	UnitID                      int
	UnitName                    string
	Brand                       string
	MinStock                    float64
	MinStockDisplay             string
	MaxStock                    float64
	MaxStockDisplay             string
	ReorderPoint                float64
	ReorderPointDisplay         string
	DefaultLeadTimeDays         int
	PackSize                    float64
	PackSizeDisplay             string
	IsActive                    bool
	StatusLabel                 string
	PrimarySupplierID           int
	PrimarySupplierName         string
	PrimarySupplierPrice        float64
	PrimarySupplierPriceDisplay string
	CreatedAt                   string
	CreatedAtDisplay            string
	UpdatedAt                   string
	UpdatedAtDisplay            string
}

type ProductCategory struct {
	ID           int
	CategoryCode string
	CategoryName string
}

type ProductSupplierOption struct {
	ID           int
	SupplierName string
}

type ProductStats struct {
	TotalProducts    int
	ActiveProducts   int
	InactiveProducts int
	LinkedSuppliers  int
}

type ProductDetail struct {
	Product
	CurrentStock         float64
	CurrentStockDisplay  string
	OnOrderQty           float64
	OnOrderQtyDisplay    string
	AvailabilityPercent  int
	SupplierNetworkCount int
	StockHistoryCount    int
	LatestSessionDate    string
	LatestSessionDisplay string
}

type ProductSupplierNetwork struct {
	SupplierID       int
	SupplierCode     string
	SupplierName     string
	SupplierType     string
	Address          string
	LastPrice        float64
	LastPriceDisplay string
	MOQ              float64
	MOQDisplay       string
	PackSize         float64
	PackSizeDisplay  string
	LeadTimeDays     int
	PriorityNo       int
	IsPrimary        bool
	IsActive         bool
	PriorityLabel    string
	StatusLabel      string
}

type ProductStockHistory struct {
	ItemID                int
	SessionID             int
	SessionNumber         string
	SessionDate           string
	SessionDateDisplay    string
	StoreName             string
	CheckerName           string
	QtyStore              float64
	QtyStoreDisplay       string
	QtyWarehouse          float64
	QtyWarehouseDisplay   string
	Discrepancy           float64
	DiscrepancyDisplay    string
	SuggestBuyQty         float64
	SuggestBuyQtyDisplay  string
	ApprovedBuyQty        float64
	ApprovedBuyQtyDisplay string
	CheckerNotes          string
	Status                string
	StatusLabel           string
	StatusBadgeClass      string
}

type ProductListFilter struct {
	Search     string
	CategoryID int
	Status     string
	Brand      string
	Sort       string
	Page       int
	Limit      int
}

type ProductCreateInput struct {
	ProductCode         string
	Barcode             string
	ProductName         string
	CategoryID          int
	UnitID              int
	Brand               string
	MinStock            float64
	MaxStock            float64
	ReorderPoint        float64
	DefaultLeadTimeDays int
	PackSize            float64
	IsActive            bool
	SupplierID          int
	LastPrice           float64
}

type ProductUpdateInput struct {
	ID                  int
	ProductCode         string
	Barcode             string
	ProductName         string
	CategoryID          int
	UnitID              int
	Brand               string
	MinStock            float64
	MaxStock            float64
	ReorderPoint        float64
	DefaultLeadTimeDays int
	PackSize            float64
	IsActive            bool
	SupplierID          int
	LastPrice           float64
}
