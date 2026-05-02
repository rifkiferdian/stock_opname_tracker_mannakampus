package controllers

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	status := c.Query("status")
	reportService := buildStockOpnameReportService()

	if c.Query("export") == "csv" {
		exportPage, err := reportService.GetDetailPage(id, models.StockOpnameReportFilter{
			Status: status,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		renderStockOpnameReportCSV(c, supplier, exportPage)
		return
	}

	reportPage, err := reportService.GetDetailPage(id, models.StockOpnameReportFilter{
		Status: status,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	reportPage.ExportURL = buildStockOpnameReportExportURL(id, reportPage.Filter)

	Render(c, "stock_opname_report.html", gin.H{
		"Title":    "Laporan Stock Opname",
		"Page":     "report_stock_opname_detail",
		"Supplier": supplier,
		"Report":   reportPage,
		"Success":  c.Query("success"),
		"Error":    c.Query("error"),
	})
}

func buildStockOpnameReportService() *services.StockOpnameReportService {
	repo := &repositories.StockOpnameReportRepository{DB: config.DB}
	return &services.StockOpnameReportService{Repo: repo}
}

func renderStockOpnameReportCSV(c *gin.Context, supplier models.Supplier, page models.StockOpnameReportPage) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	headers := []string{
		"Product Code",
		"Product Name",
		"Barcode",
		"Category",
		"Lead Time",
		"Current Date",
		"Shop Qty",
		"Warehouse Qty",
		"PO Qty",
		"Latest Status",
	}
	for _, label := range page.HistoryDateLabels {
		headers = append(headers,
			fmt.Sprintf("%s Date", label),
			fmt.Sprintf("%s Shop Qty", label),
			fmt.Sprintf("%s Warehouse Qty", label),
			fmt.Sprintf("%s PO Qty", label),
		)
	}

	if err := writer.Write(headers); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	for _, row := range page.ReportRows {
		record := []string{
			row.SKU,
			row.Name,
			row.Barcode,
			row.CategoryName,
			row.LeadTimeLabel,
			page.CurrentDateLabel,
			row.CurrentShop,
			row.CurrentWH,
			row.CurrentPO,
			row.CurrentStatus,
		}

		for _, history := range row.History {
			record = append(record, history.DateLabel, history.ShopCount, history.WHSCount, history.POCount)
		}

		if err := writer.Write(record); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("stock-opname-report-%s.csv", supplier.SupplierCode)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buffer.Bytes())
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

func buildStockOpnameReportExportURL(supplierID int, filter models.StockOpnameReportFilter) string {
	values := url.Values{}
	if filter.Status != "" {
		values.Set("status", filter.Status)
	}
	values.Set("export", "csv")
	return fmt.Sprintf("/reports/stock-opname/%d?%s", supplierID, values.Encode())
}
