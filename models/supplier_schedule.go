package models

type SupplierSchedule struct {
	ID               int
	StoreID          int
	StoreName        string
	SupplierID       int
	SupplierCode     string
	SupplierName     string
	DayOfWeek        int
	DayName          string
	SOTime           string
	SequenceNo       int
	IsActive         bool
	StatusLabel      string
	Notes            string
	UpdatedAt        string
	UpdatedAtDisplay string
}

type SupplierScheduleListFilter struct {
	Search    string
	StoreID   int
	DayOfWeek int
	Status    string
}

type SupplierScheduleCreateInput struct {
	StoreID    int
	SupplierID int
	DayOfWeek  int
	SOTime     string
	SequenceNo int
	IsActive   bool
	Notes      string
}

type SupplierScheduleUpdateInput struct {
	ID         int
	StoreID    int
	SupplierID int
	DayOfWeek  int
	SOTime     string
	SequenceNo int
	IsActive   bool
	Notes      string
}

type SupplierScheduleStats struct {
	TotalSuppliers       int
	ScheduledSuppliers   int
	UnscheduledSuppliers int
	TotalSchedules       int
}
