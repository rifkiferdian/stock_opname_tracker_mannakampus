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

type StockCheckSessionUpdateInput struct {
	ID             int
	SessionDate    string
	StoreID        int
	SupplierID     int
	InitiationType string
	Status         string
	Notes          string
}
