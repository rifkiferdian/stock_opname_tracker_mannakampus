package controllers

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gobase-app/models"

	"github.com/gin-gonic/gin"
)

func StockCheckPORecapIndex(c *gin.Context) {
	currentUserID := extractCurrentUserID(c)
	if currentUserID <= 0 {
		c.Redirect(http.StatusFound, "/logout")
		return
	}

	service := buildStockCheckSessionService()
	dateFrom := sanitizeQueryDate(c.Query("date_from"))
	dateTo := sanitizeQueryDate(c.Query("date_to"))
	todayKey := time.Now().Format("2006-01-02")

	filter := models.StockCheckSessionListFilter{
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		SupplierName: c.Query("supplier_name"),
		Status:       "closed",
		Page:         parsePositiveInt(c.Query("page"), 1),
		Limit:        15,
	}
	if filter.DateFrom != "" && filter.DateTo != "" && filter.DateFrom > filter.DateTo {
		filter.DateFrom, filter.DateTo = filter.DateTo, filter.DateFrom
	}

	queryFilter := filter
	queryFilter.Page = 1
	queryFilter.Limit = 10000

	sessions, _, err := service.GetSessions(queryFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	filtered := make([]models.StockCheckSession, 0, len(sessions))
	if currentUserHasRole(c, "super-admin") {
		filtered = sessions
	} else {
		stores, err := service.GetStoreOptionsByUserID(currentUserID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		allowedStoreSet := buildStoreAccessSet(stores)
		for _, session := range sessions {
			if !isStoreAccessible(allowedStoreSet, session.StoreID) {
				continue
			}
			filtered = append(filtered, session)
		}
	}

	totalItems := len(filtered)
	pagination := buildStockCheckPORecapPagination(filter, totalItems)
	pagedSessions := paginateStockCheckPORecapSessions(filtered, pagination)

	Render(c, "stock_check_po_recap.html", gin.H{
		"Title":       "Rekap PO Supplier",
		"Page":        "stock_check_po_recap",
		"CurrentRole": extractCurrentUserRole(c),
		"TodayKey":    todayKey,
		"Filters":     filter,
		"Sessions":    pagedSessions,
		"Pagination":  pagination,
	})
}

func paginateStockCheckPORecapSessions(items []models.StockCheckSession, pagination models.Pagination) []models.StockCheckSession {
	if len(items) == 0 || pagination.PageSize <= 0 || pagination.CurrentPage <= 0 {
		return items
	}

	startIndex := (pagination.CurrentPage - 1) * pagination.PageSize
	if startIndex >= len(items) {
		return []models.StockCheckSession{}
	}
	endIndex := startIndex + pagination.PageSize
	if endIndex > len(items) {
		endIndex = len(items)
	}
	return items[startIndex:endIndex]
}

func buildStockCheckPORecapPagination(filter models.StockCheckSessionListFilter, totalItems int) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: filter.Page,
		PageSize:    filter.Limit,
		TotalItems:  totalItems,
	}

	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 15
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
		pagination.PrevURL = buildStockCheckPORecapPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockCheckPORecapPageURL(filter, pagination.CurrentPage+1)
	}

	startPage := pagination.CurrentPage - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 2
	if endPage > pagination.TotalPages {
		endPage = pagination.TotalPages
	}
	if endPage-startPage < 2 {
		startPage = endPage - 2
		if startPage < 1 {
			startPage = 1
		}
	}

	for page := startPage; page <= endPage; page++ {
		pagination.Pages = append(pagination.Pages, models.PaginationLink{
			Number: page,
			URL:    buildStockCheckPORecapPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockCheckPORecapPageURL(filter models.StockCheckSessionListFilter, page int) string {
	values := url.Values{}
	if filter.DateFrom != "" {
		values.Set("date_from", filter.DateFrom)
	}
	if filter.DateTo != "" {
		values.Set("date_to", filter.DateTo)
	}
	if filter.SupplierName != "" {
		values.Set("supplier_name", filter.SupplierName)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}

	encoded := values.Encode()
	if encoded == "" {
		return "/stock-checker/po-recap"
	}
	return "/stock-checker/po-recap?" + encoded
}
