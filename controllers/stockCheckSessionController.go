package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
