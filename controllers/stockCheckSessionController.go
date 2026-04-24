package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const stockCheckSessionDetailItemLimit = 100

func StockCheckSessionIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	storeID, _ := strconv.Atoi(c.DefaultQuery("store_id", "0"))
	supplierID, _ := strconv.Atoi(c.DefaultQuery("supplier_id", "0"))

	renderStockCheckSessionPage(c, buildStockCheckSessionService(), "", "", models.StockCheckSession{}, models.StockCheckSessionListFilter{
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		StoreID:    storeID,
		SupplierID: supplierID,
		Status:     c.Query("status"),
		Page:       page,
		Limit:      10,
	})
}

func StockCheckSessionDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	renderStockCheckSessionDetailPage(c, buildStockCheckSessionService(), id, c.Query("success"), "", models.StockCheckSessionReviewItemEditForm{})
}

func StockCheckSessionReviewItemUpdate(c *gin.Context) {
	type stockCheckSessionReviewItemForm struct {
		ItemID         int    `form:"item_id" binding:"required"`
		ApprovedBuyQty string `form:"approved_buy_qty" binding:"required"`
		BuyerNotes     string `form:"buyer_notes"`
	}

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	var form stockCheckSessionReviewItemForm
	service := buildStockCheckSessionService()

	if err := c.ShouldBind(&form); err != nil {
		renderStockCheckSessionDetailPage(c, service, sessionID, "", "Form edit item tidak lengkap", models.StockCheckSessionReviewItemEditForm{
			ItemID:         form.ItemID,
			ApprovedBuyQty: form.ApprovedBuyQty,
			BuyerNotes:     form.BuyerNotes,
		})
		return
	}

	approvedBuyQty, err := strconv.ParseFloat(strings.TrimSpace(form.ApprovedBuyQty), 64)
	if err != nil {
		renderStockCheckSessionDetailPage(c, service, sessionID, "", "Final approve harus berupa angka yang valid", models.StockCheckSessionReviewItemEditForm{
			ItemID:         form.ItemID,
			ApprovedBuyQty: form.ApprovedBuyQty,
			BuyerNotes:     form.BuyerNotes,
		})
		return
	}

	err = service.UpdateReviewItem(models.StockCheckSessionReviewItemUpdateInput{
		SessionID:      sessionID,
		ItemID:         form.ItemID,
		ApprovedBuyQty: approvedBuyQty,
		BuyerNotes:     form.BuyerNotes,
		ReviewedBy:     extractCurrentUserID(c),
		UpdatedBy:      extractCurrentUserID(c),
	})
	if err != nil {
		renderStockCheckSessionDetailPage(c, service, sessionID, "", err.Error(), models.StockCheckSessionReviewItemEditForm{
			ItemID:         form.ItemID,
			ApprovedBuyQty: form.ApprovedBuyQty,
			BuyerNotes:     form.BuyerNotes,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, buildStockCheckSessionDetailPageURL(sessionID, parsePositiveInt(c.Query("page"), 1), "Item review berhasil diperbarui"))
}

func StockCheckSessionStore(c *gin.Context) {
	type stockCheckSessionForm struct {
		SessionDate    string `form:"session_date" binding:"required"`
		StoreID        int    `form:"store_id" binding:"required"`
		SupplierID     int    `form:"supplier_id" binding:"required"`
		InitiationType string `form:"initiation_type" binding:"required"`
		Status         string `form:"status" binding:"required"`
		Notes          string `form:"notes"`
	}

	var form stockCheckSessionForm
	service := buildStockCheckSessionService()
	filter := buildStockCheckSessionFilter(c)

	if err := c.ShouldBind(&form); err != nil {
		renderStockCheckSessionPage(c, service, "Form stock check session tidak lengkap", "create", models.StockCheckSession{
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}, filter)
		return
	}

	err := service.CreateSession(models.StockCheckSessionCreateInput{
		SessionDate:    form.SessionDate,
		StoreID:        form.StoreID,
		SupplierID:     form.SupplierID,
		InitiationType: form.InitiationType,
		Status:         form.Status,
		Notes:          form.Notes,
		CreatedBy:      extractCurrentUserID(c),
	})
	if err != nil {
		renderStockCheckSessionPage(c, service, err.Error(), "create", models.StockCheckSession{
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}, filter)
		return
	}

	c.Redirect(http.StatusSeeOther, "/stock-check-sessions")
}

func StockCheckSessionUpdate(c *gin.Context) {
	type stockCheckSessionForm struct {
		ID             int    `form:"id" binding:"required"`
		SessionNumber  string `form:"session_number_display"`
		SessionDate    string `form:"session_date" binding:"required"`
		StoreID        int    `form:"store_id" binding:"required"`
		SupplierID     int    `form:"supplier_id" binding:"required"`
		InitiationType string `form:"initiation_type" binding:"required"`
		Status         string `form:"status" binding:"required"`
		Notes          string `form:"notes"`
	}

	var form stockCheckSessionForm
	service := buildStockCheckSessionService()
	filter := buildStockCheckSessionFilter(c)

	if err := c.ShouldBind(&form); err != nil {
		renderStockCheckSessionPage(c, service, "Form edit stock check session tidak lengkap", "edit", models.StockCheckSession{
			ID:             form.ID,
			SessionNumber:  form.SessionNumber,
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}, filter)
		return
	}

	err := service.UpdateSession(models.StockCheckSessionUpdateInput{
		ID:             form.ID,
		SessionDate:    form.SessionDate,
		StoreID:        form.StoreID,
		SupplierID:     form.SupplierID,
		InitiationType: form.InitiationType,
		Status:         form.Status,
		Notes:          form.Notes,
	})
	if err != nil {
		renderStockCheckSessionPage(c, service, err.Error(), "edit", models.StockCheckSession{
			ID:             form.ID,
			SessionNumber:  form.SessionNumber,
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}, filter)
		return
	}

	c.Redirect(http.StatusSeeOther, "/stock-check-sessions")
}

func StockCheckSessionDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	service := buildStockCheckSessionService()
	if err := service.DeleteSession(id); err != nil {
		renderStockCheckSessionPage(c, service, err.Error(), "", models.StockCheckSession{}, buildStockCheckSessionFilter(c))
		return
	}

	c.Redirect(http.StatusSeeOther, "/stock-check-sessions")
}

func buildStockCheckSessionService() *services.StockCheckSessionService {
	repo := &repositories.StockCheckSessionRepository{DB: config.DB}
	return &services.StockCheckSessionService{Repo: repo}
}

func renderStockCheckSessionPage(c *gin.Context, service *services.StockCheckSessionService, message string, formMode string, formSession models.StockCheckSession, filter models.StockCheckSessionListFilter) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	sessions, totalItems, err := service.GetSessions(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stores, err := service.GetStoreOptions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	suppliers, err := service.GetSupplierOptions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	pagination := buildStockCheckSessionPagination(filter, totalItems)

	Render(c, "stock_check_sessions.html", gin.H{
		"Title":       "Stock Check Sessions",
		"Page":        "stock_check_sessions",
		"Sessions":    sessions,
		"Stores":      stores,
		"Suppliers":   suppliers,
		"Filters":     filter,
		"Pagination":  pagination,
		"Error":       message,
		"FormMode":    formMode,
		"FormSession": formSession,
	})
}

func renderStockCheckSessionDetailPage(c *gin.Context, service *services.StockCheckSessionService, id int, successMessage string, errorMessage string, reviewForm models.StockCheckSessionReviewItemEditForm) {
	currentPage := parsePositiveInt(c.Query("page"), 1)
	pageData, err := service.GetSessionDetailPage(id, currentPage, stockCheckSessionDetailItemLimit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Stock check session tidak ditemukan",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	pageData.Pagination = buildStockCheckSessionDetailPagination(id, pageData.Pagination)

	Render(c, "stock_check_session_detail.html", gin.H{
		"Title":       pageData.Session.SessionNumber,
		"Page":        "stock_check_sessions",
		"Session":     pageData.Session,
		"Items":       pageData.Items,
		"Overview":    pageData.OverviewCards,
		"Pagination":  pageData.Pagination,
		"Success":     successMessage,
		"Error":       errorMessage,
		"ReviewForm":  reviewForm,
		"CurrentPath": c.Request.URL.Path,
	})
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func buildStockCheckSessionDetailPagination(sessionID int, pagination models.Pagination) models.Pagination {
	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = stockCheckSessionDetailItemLimit
	}
	if pagination.TotalItems == 0 {
		return pagination
	}

	if pagination.TotalPages <= 0 {
		pagination.TotalPages = (pagination.TotalItems + pagination.PageSize - 1) / pagination.PageSize
	}
	if pagination.CurrentPage > pagination.TotalPages {
		pagination.CurrentPage = pagination.TotalPages
	}

	pagination.StartItem = ((pagination.CurrentPage - 1) * pagination.PageSize) + 1
	pagination.EndItem = pagination.StartItem + pagination.PageSize - 1
	if pagination.EndItem > pagination.TotalItems {
		pagination.EndItem = pagination.TotalItems
	}

	pagination.HasPrev = pagination.CurrentPage > 1
	pagination.HasNext = pagination.CurrentPage < pagination.TotalPages
	if pagination.HasPrev {
		pagination.PrevURL = buildStockCheckSessionDetailPageURL(sessionID, pagination.CurrentPage-1, "")
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockCheckSessionDetailPageURL(sessionID, pagination.CurrentPage+1, "")
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

	pagination.Pages = nil
	for page := startPage; page <= endPage; page++ {
		pagination.Pages = append(pagination.Pages, models.PaginationLink{
			Number: page,
			URL:    buildStockCheckSessionDetailPageURL(sessionID, page, ""),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockCheckSessionDetailPageURL(sessionID int, page int, successMessage string) string {
	values := url.Values{}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	if successMessage != "" {
		values.Set("success", successMessage)
	}

	baseURL := fmt.Sprintf("/stock-check-sessions/%d", sessionID)
	encoded := values.Encode()
	if encoded == "" {
		return baseURL
	}
	return baseURL + "?" + encoded
}

func buildStockCheckSessionFilter(c *gin.Context) models.StockCheckSessionListFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	storeID, _ := strconv.Atoi(c.DefaultQuery("store_id", "0"))
	supplierID, _ := strconv.Atoi(c.DefaultQuery("supplier_id", "0"))

	return models.StockCheckSessionListFilter{
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		StoreID:    storeID,
		SupplierID: supplierID,
		Status:     c.Query("status"),
		Page:       page,
		Limit:      10,
	}
}

func extractCurrentUserID(c *gin.Context) int {
	sess := sessions.Default(c)

	if raw := sess.Get("user_id"); raw != nil {
		switch id := raw.(type) {
		case int:
			return id
		case int64:
			return int(id)
		case float64:
			return int(id)
		}
	}

	if raw := sess.Get("user"); raw != nil {
		switch user := raw.(type) {
		case models.SessionUser:
			return user.UserID
		case map[string]interface{}:
			if id, ok := user["user_id"].(float64); ok {
				return int(id)
			}
		case gin.H:
			if id, ok := user["user_id"].(float64); ok {
				return int(id)
			}
		}
	}

	return 0
}

func buildStockCheckSessionPagination(filter models.StockCheckSessionListFilter, totalItems int) models.Pagination {
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
		pagination.PrevURL = buildStockCheckSessionPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockCheckSessionPageURL(filter, pagination.CurrentPage+1)
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
			URL:    buildStockCheckSessionPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockCheckSessionPageURL(filter models.StockCheckSessionListFilter, page int) string {
	values := url.Values{}
	if filter.DateFrom != "" {
		values.Set("date_from", filter.DateFrom)
	}
	if filter.DateTo != "" {
		values.Set("date_to", filter.DateTo)
	}
	if filter.StoreID > 0 {
		values.Set("store_id", strconv.Itoa(filter.StoreID))
	}
	if filter.SupplierID > 0 {
		values.Set("supplier_id", strconv.Itoa(filter.SupplierID))
	}
	if filter.Status != "" {
		values.Set("status", filter.Status)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}

	encoded := values.Encode()
	if encoded == "" {
		return "/stock-check-sessions"
	}
	return "/stock-check-sessions?" + encoded
}
