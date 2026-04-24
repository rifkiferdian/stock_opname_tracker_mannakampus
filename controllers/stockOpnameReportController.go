package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type stockOpnameSummaryCard struct {
	Label string
	Value string
	Note  string
	Tone  string
}

type stockOpnameHistoryPoint struct {
	ShopCount string
	WHSCount  string
	Approx    string
}

type stockOpnameReportRow struct {
	Name           string
	SKU            string
	IconClass      string
	ThumbnailTone  string
	CurrentShop    string
	CurrentWH      string
	CurrentSuggest string
	StatusTone     string
	History        []stockOpnameHistoryPoint
	ActionLabel    string
}

type stockOpnameTrendBar struct {
	Label     string
	Height    string
	IsCurrent bool
	Value     string
}

type stockOpnameAuditFinding struct {
	Title string
	SKU   string
	Icon  string
}

func StockOpnameReportIndex(c *gin.Context) {
	supplierService := buildSupplierService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	renderStockOpnameReportSupplierPage(c, supplierService, "", models.SupplierListFilter{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Type:   c.Query("type"),
		Sort:   c.DefaultQuery("sort", "recent"),
		Page:   page,
		Limit:  10,
	})
}

func StockOpnameReportDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	supplierService := buildSupplierService()
	supplier, err := supplierService.GetSupplierByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Supplier report not found",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "stock_opname_report.html", gin.H{
		"Title":         "Laporan Stock Opname",
		"Page":          "report_stock_opname_detail",
		"Supplier":      supplier,
		"SummaryCards":  stockOpnameSummaryCards(),
		"ReportDates":   stockOpnameReportDates(),
		"ReportRows":    stockOpnameReportRows(),
		"TrendBars":     stockOpnameTrendBars(),
		"AuditFindings": stockOpnameAuditFindings(),
	})
}

func renderStockOpnameReportSupplierPage(c *gin.Context, supplierService *services.SupplierService, message string, filter models.SupplierListFilter) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	suppliers, totalItems, err := supplierService.GetSuppliers(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stats, err := supplierService.GetSupplierStats()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	types, err := supplierService.GetSupplierTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "stock_opname_report_supplier_list.html", gin.H{
		"Title":      "Laporan Stock Opname",
		"Page":       "report_stock_opname",
		"suppliers":  suppliers,
		"Stats":      stats,
		"Types":      types,
		"Filters":    filter,
		"Pagination": buildStockOpnameReportPagination(filter, totalItems),
		"Error":      message,
	})
}

func buildStockOpnameReportPagination(filter models.SupplierListFilter, totalItems int) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: filter.Page,
		PageSize:    filter.Limit,
		TotalItems:  totalItems,
	}

	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}
	if totalItems == 0 {
		return pagination
	}

	pagination.TotalPages = (totalItems + pagination.PageSize - 1) / pagination.PageSize
	if pagination.CurrentPage > pagination.TotalPages {
		pagination.CurrentPage = pagination.TotalPages
	}

	pagination.StartItem = ((pagination.CurrentPage - 1) * pagination.PageSize) + 1
	pagination.EndItem = pagination.StartItem + pagination.PageSize - 1
	if pagination.EndItem > totalItems {
		pagination.EndItem = totalItems
	}

	pagination.HasPrev = pagination.CurrentPage > 1
	pagination.HasNext = pagination.CurrentPage < pagination.TotalPages
	if pagination.HasPrev {
		pagination.PrevURL = buildStockOpnameReportPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockOpnameReportPageURL(filter, pagination.CurrentPage+1)
	}

	startPage := pagination.CurrentPage - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 4
	if endPage > pagination.TotalPages {
		endPage = pagination.TotalPages
	}
	if endPage-startPage < 4 {
		startPage = endPage - 4
		if startPage < 1 {
			startPage = 1
		}
	}

	for page := startPage; page <= endPage; page++ {
		pagination.Pages = append(pagination.Pages, models.PaginationLink{
			Number: page,
			URL:    buildStockOpnameReportPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockOpnameReportPageURL(filter models.SupplierListFilter, page int) string {
	values := url.Values{}
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	if filter.Status != "" {
		values.Set("status", filter.Status)
	}
	if filter.Type != "" {
		values.Set("type", filter.Type)
	}
	if filter.Sort != "" && filter.Sort != "recent" {
		values.Set("sort", filter.Sort)
	} else if filter.Sort == "recent" {
		values.Set("sort", "recent")
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}

	encoded := values.Encode()
	if encoded == "" {
		return "/reports/stock-opname"
	}
	return "/reports/stock-opname?" + encoded
}

func stockOpnameSummaryCards() []stockOpnameSummaryCard {
	return []stockOpnameSummaryCard{
		{Label: "Total item dalam review", Value: "142", Note: "+12 pagi ini", Tone: "info"},
		{Label: "Tingkat selisih", Value: "4.2%", Note: "Di atas target", Tone: "danger"},
		{Label: "Menunggu persetujuan", Value: "28", Note: "Jatuh tempo hari ini", Tone: "info"},
	}
}

func stockOpnameReportRows() []stockOpnameReportRow {
	return []stockOpnameReportRow{
		{
			Name:           "Ultra-Lite Trainer",
			SKU:            "SKU: FTW-9902-WH",
			IconClass:      "bx bx-body",
			ThumbnailTone:  "neutral",
			CurrentShop:    "12",
			CurrentWH:      "10",
			CurrentSuggest: "Periksa Input",
			StatusTone:     "input",
			ActionLabel:    "Tinjau",
			History: []stockOpnameHistoryPoint{
				{ShopCount: "12", WHSCount: "10", Approx: "422"},
				{ShopCount: "08", WHSCount: "325", Approx: "403"},
				{ShopCount: "05", WHSCount: "512", Approx: "537"},
				{ShopCount: "15", WHSCount: "480", Approx: "495"},
				{ShopCount: "11", WHSCount: "468", Approx: "479"},
				{ShopCount: "09", WHSCount: "446", Approx: "455"},
				{ShopCount: "14", WHSCount: "438", Approx: "452"},
				{ShopCount: "10", WHSCount: "410", Approx: "420"},
				{ShopCount: "06", WHSCount: "398", Approx: "404"},
				{ShopCount: "07", WHSCount: "385", Approx: "392"},
				{ShopCount: "13", WHSCount: "372", Approx: "385"},
				{ShopCount: "08", WHSCount: "360", Approx: "368"},
			},
		},
		{
			Name:           "AeroStride Pro 7",
			SKU:            "SKU: FTW-8812-RD",
			IconClass:      "bx bx-run",
			ThumbnailTone:  "accent",
			CurrentShop:    "2",
			CurrentWH:      "115",
			CurrentSuggest: "Stok Menipis",
			StatusTone:     "low",
			ActionLabel:    "Eskalasi",
			History: []stockOpnameHistoryPoint{
				{ShopCount: "02", WHSCount: "115", Approx: "117"},
				{ShopCount: "05", WHSCount: "140", Approx: "145"},
				{ShopCount: "01", WHSCount: "190", Approx: "191"},
				{ShopCount: "08", WHSCount: "220", Approx: "228"},
				{ShopCount: "04", WHSCount: "205", Approx: "209"},
				{ShopCount: "03", WHSCount: "198", Approx: "201"},
				{ShopCount: "06", WHSCount: "184", Approx: "190"},
				{ShopCount: "02", WHSCount: "176", Approx: "178"},
				{ShopCount: "03", WHSCount: "168", Approx: "171"},
				{ShopCount: "02", WHSCount: "159", Approx: "161"},
				{ShopCount: "01", WHSCount: "152", Approx: "153"},
				{ShopCount: "03", WHSCount: "147", Approx: "150"},
			},
		},
		{
			Name:           "SonicMaster Studio",
			SKU:            "SKU: EL-4431-BLK",
			IconClass:      "bx bx-headphone",
			ThumbnailTone:  "dark",
			CurrentShop:    "5",
			CurrentWH:      "200",
			CurrentSuggest: "Stabil",
			StatusTone:     "stable",
			ActionLabel:    "Periksa",
			History: []stockOpnameHistoryPoint{
				{ShopCount: "05", WHSCount: "200", Approx: "205"},
				{ShopCount: "03", WHSCount: "218", Approx: "221"},
				{ShopCount: "12", WHSCount: "195", Approx: "207"},
				{ShopCount: "02", WHSCount: "240", Approx: "242"},
				{ShopCount: "04", WHSCount: "231", Approx: "235"},
				{ShopCount: "06", WHSCount: "224", Approx: "230"},
				{ShopCount: "05", WHSCount: "215", Approx: "220"},
				{ShopCount: "03", WHSCount: "209", Approx: "212"},
				{ShopCount: "04", WHSCount: "202", Approx: "206"},
				{ShopCount: "02", WHSCount: "196", Approx: "198"},
				{ShopCount: "05", WHSCount: "188", Approx: "193"},
				{ShopCount: "03", WHSCount: "182", Approx: "185"},
			},
		},
	}
}

func stockOpnameReportDates() []string {
	return []string{
		"20 Oct 2023",
		"05 Oct 2023",
		"20 Sep 2023",
		"05 Sep 2023",
		"20 Aug 2023",
		"05 Aug 2023",
		"20 Jul 2023",
		"05 Jul 2023",
		"20 Jun 2023",
		"05 Jun 2023",
		"20 May 2023",
		"05 May 2023",
	}
}

func stockOpnameTrendBars() []stockOpnameTrendBar {
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	approvalCounts := []int{28, 32, 30, 35, 38, 36, 41, 39, 44, 47, 45, 52}
	maxApproval := 52
	now := time.Now()

	bars := make([]stockOpnameTrendBar, 0, 12)
	for index, count := range approvalCounts {
		monthTime := now.AddDate(0, -(len(approvalCounts) - 1 - index), 0)
		height := 28 + (count * 56 / maxApproval)

		bars = append(bars, stockOpnameTrendBar{
			Label:     fmt.Sprintf("%s %d", monthNames[int(monthTime.Month())-1], monthTime.Year()),
			Height:    fmt.Sprintf("%d%%", height),
			IsCurrent: index == len(approvalCounts)-1,
			Value:     strconv.Itoa(count),
		})
	}

	return bars
}

func stockOpnameAuditFindings() []stockOpnameAuditFinding {
	return []stockOpnameAuditFinding{
		{Title: "Input Tidak Konsisten", SKU: "SKU: FTW-8812-RD", Icon: "bx bx-error"},
		{Title: "Kecocokan Presisi Tinggi", SKU: "SKU: EL-4431-BLK", Icon: "bx bx-check-shield"},
	}
}
