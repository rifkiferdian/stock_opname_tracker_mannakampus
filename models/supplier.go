package models

type Supplier struct {
	ID                int
	StoreID           int
	StoreName         string
	SupplierGroupID   int
	SupplierGroupName string
	SupplierCode      string
	SupplierName      string
	SupplierType      string
	Address           string
	Phone             string
	Email             string
	PICName           string
	PaymentTermDays   int
	IsActive          bool
	StatusLabel       string
	ProductCount      int
	CreatedAt         string
	CreatedAtDisplay  string
	UpdatedAt         string
	UpdatedAtDisplay  string
	LastSODate        string
	LastSODateDisplay string
}

type SupplierGroup struct {
	ID               int
	StoreID          int
	StoreName        string
	GroupCode        string
	GroupName        string
	Description      string
	IsActive         bool
	StatusLabel      string
	SupplierCount    int
	CreatedAt        string
	CreatedAtDisplay string
	UpdatedAt        string
	UpdatedAtDisplay string
}

type SupplierProduct struct {
	ProductID        int
	ProductCode      string
	Barcode          string
	ProductName      string
	CategoryName     string
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
	StatusLabel      string
	UpdatedAt        string
	UpdatedAtDisplay string
}

type SupplierProductOption struct {
	ID           int
	ProductCode  string
	Barcode      string
	ProductName  string
	CategoryName string
	UnitName     string
}

type SupplierStats struct {
	TotalSuppliers    int
	ActiveSuppliers   int
	InactiveSuppliers int
	LinkedProducts    int
}

type SupplierGroupStats struct {
	TotalGroups     int
	ActiveGroups    int
	InactiveGroups  int
	LinkedSuppliers int
}

type SupplierListFilter struct {
	Search     string
	SearchMode string
	Status     string
	Type       string
	Sort       string
	Page       int
	Limit      int
}

type SupplierGroupListFilter struct {
	Search string
	Status string
	Sort   string
	Page   int
	Limit  int
}

type PaginationLink struct {
	Number int
	URL    string
	Active bool
}

type Pagination struct {
	CurrentPage int
	PageSize    int
	TotalItems  int
	TotalPages  int
	StartItem   int
	EndItem     int
	HasPrev     bool
	HasNext     bool
	PrevURL     string
	NextURL     string
	Pages       []PaginationLink
}

type SupplierCreateInput struct {
	StoreID         int
	SupplierGroupID int
	SupplierCode    string
	SupplierName    string
	SupplierType    string
	Address         string
	Phone           string
	Email           string
	PICName         string
	PaymentTermDays int
	IsActive        bool
}

type SupplierGroupCreateInput struct {
	StoreID     int
	GroupCode   string
	GroupName   string
	Description string
	IsActive    bool
}

type SupplierUpdateInput struct {
	ID              int
	StoreID         int
	SupplierGroupID int
	SupplierCode    string
	SupplierName    string
	SupplierType    string
	Address         string
	Phone           string
	Email           string
	PICName         string
	PaymentTermDays int
	IsActive        bool
}

type SupplierProductCreateInput struct {
	SupplierID   int
	ProductID    int
	LastPrice    float64
	MOQ          float64
	PackSize     float64
	LeadTimeDays int
	IsPrimary    bool
	IsActive     bool
}

type SupplierProductDeleteInput struct {
	SupplierID int
	ProductID  int
}

type SupplierGroupUpdateInput struct {
	ID          int
	StoreID     int
	GroupCode   string
	GroupName   string
	Description string
	IsActive    bool
}
