package models

type Unit struct {
	ID               int
	UnitCode         string
	UnitName         string
	Description      string
	CreatedAt        string
	CreatedAtDisplay string
	UpdatedAt        string
	UpdatedAtDisplay string
}

type UnitCreateInput struct {
	UnitCode    string
	UnitName    string
	Description string
}

type UnitListFilter struct {
	Search string
	Sort   string
}

type UnitUpdateInput struct {
	ID          int
	UnitCode    string
	UnitName    string
	Description string
}
