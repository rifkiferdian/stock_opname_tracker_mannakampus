package models

type Product struct {
	ID                          int
	RowNumber                   int
	StoreID                     int
	StoreName                   string
	ProductCode                 string
	Barcode                     string
	BarcodeBox                  string
	BarcodeCarton               string
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
	PcsPerBox                   int
	BoxPerCarton                int
	PcsPerCarton                int
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
	StoreID      int
	StoreName    string
}

type ProductStats struct {
	TotalProducts    int
	ActiveProducts   int
	InactiveProducts int
	LinkedSuppliers  int
}

type ProductDetail struct {
	Product
	CurrentStock                 float64
	CurrentStockDisplay          string
	CurrentStockBreakdownDisplay string
	CurrentStockBreakdownParts   []string
	OnOrderQty                   float64
	OnOrderQtyDisplay            string
	OnOrderCartonDisplay         string
	OnOrderBreakdownDisplay      string
	OnOrderBreakdownParts        []string
	AvailabilityPercent          int
	SupplierNetworkCount         int
	StockHistoryCount            int
	LatestSessionDate            string
	LatestSessionDisplay         string
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
	QtyStoreCarton        int
	QtyStoreBox           int
	QtyStorePcs           int
	QtyStore              float64
	QtyStoreDisplay       string
	QtyStoreBreakdown     string
	QtyWarehouseCarton    int
	QtyWarehouseBox       int
	QtyWarehousePcs       int
	QtyWarehouse          float64
	QtyWarehouseDisplay   string
	QtyWarehouseBreakdown string
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
	StoreID             int
	ProductCode         string
	Barcode             string
	BarcodeBox          string
	BarcodeCarton       string
	ProductName         string
	CategoryID          int
	UnitID              int
	Brand               string
	MinStock            float64
	MaxStock            float64
	ReorderPoint        float64
	DefaultLeadTimeDays int
	PackSize            float64
	PcsPerBox           int
	BoxPerCarton        int
	PcsPerCarton        int
	IsActive            bool
	SupplierID          int
	LastPrice           float64
}

type ProductUpdateInput struct {
	ID                  int
	StoreID             int
	ProductCode         string
	Barcode             string
	BarcodeBox          string
	BarcodeCarton       string
	ProductName         string
	CategoryID          int
	UnitID              int
	Brand               string
	MinStock            float64
	MaxStock            float64
	ReorderPoint        float64
	DefaultLeadTimeDays int
	PackSize            float64
	PcsPerBox           int
	BoxPerCarton        int
	PcsPerCarton        int
	IsActive            bool
	SupplierID          int
	LastPrice           float64
}
