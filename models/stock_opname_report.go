package models

type StockOpnameReportFilter struct {
	CategoryID int
	Page       int
	Limit      int
}

type StockOpnameReportPage struct {
	Filter               StockOpnameReportFilter
	Categories           []ProductCategory
	CurrentDateLabel     string
	HistoryDateLabels    []string
	SummaryCards         []StockOpnameReportSummaryCard
	ReportRows           []StockOpnameReportRow
	TrendBars            []StockOpnameReportTrendBar
	AuditFindings        []StockOpnameReportAuditFinding
	Pagination           Pagination
	TotalRows            int
	LatestSessionCount   int
	ReviewListURL        string
	ExportURL            string
	CurrentCategoryLabel string
}

type StockOpnameReportSummaryCard struct {
	Label string
	Value string
	Note  string
	Tone  string
}

type StockOpnameReportHistoryPoint struct {
	DateLabel string
	ShopCount string
	WHSCount  string
	POCount   string
}

type StockOpnameReportRow struct {
	ProductID          int
	Name               string
	SKU                string
	Barcode            string
	Brand              string
	CategoryName       string
	LeadTimeLabel      string
	AvatarText         string
	AvatarClass        string
	CurrentShop        string
	CurrentWH          string
	CurrentPO          string
	CurrentStatus      string
	StatusTone         string
	CurrentSessionID   int
	CurrentItemID      int
	CurrentSuggestQty  string
	CurrentCheckerNote string
	CurrentBuyerNotes  string
	CurrentApproveSeed string
	CanInlineApprove   bool
	History            []StockOpnameReportHistoryPoint
	ActionLabel        string
	ActionURL          string
	LatestSessionID    int
	LatestSessionDate  string
	POTrendSeriesJSON  string
	POTrendLabelsJSON  string
	POTrendTotalLabel  string
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
