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

	"github.com/gin-gonic/gin"
)

func SupplierIndex(c *gin.Context) {
	supplierService := buildSupplierService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	renderSupplierPage(c, supplierService, "", models.SupplierListFilter{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Sort:   c.DefaultQuery("sort", "recent"),
		Page:   page,
		Limit:  150,
	})
}

func SupplierDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	supplierService := buildSupplierService()
	renderSupplierDetailPage(c, supplierService, id, c.Query("error"), c.Query("success"), models.SupplierProductCreateInput{})
}

func SupplierProductStore(c *gin.Context) {
	supplierID, err := strconv.Atoi(c.Param("id"))
	if err != nil || supplierID <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	type supplierProductForm struct {
		ProductID    int     `form:"product_id" binding:"required"`
		LastPrice    float64 `form:"last_price"`
		MOQ          float64 `form:"moq"`
		PackSize     float64 `form:"pack_size"`
		LeadTimeDays int     `form:"lead_time_days"`
		IsPrimary    int     `form:"is_primary"`
	}

	var form supplierProductForm
	supplierService := buildSupplierService()

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierDetailPage(c, supplierService, supplierID, "Form item supplier tidak lengkap", "", models.SupplierProductCreateInput{
			SupplierID:   supplierID,
			ProductID:    form.ProductID,
			LastPrice:    form.LastPrice,
			MOQ:          form.MOQ,
			PackSize:     form.PackSize,
			LeadTimeDays: form.LeadTimeDays,
			IsPrimary:    form.IsPrimary == 1,
		})
		return
	}

	input := models.SupplierProductCreateInput{
		SupplierID:   supplierID,
		ProductID:    form.ProductID,
		LastPrice:    form.LastPrice,
		MOQ:          form.MOQ,
		PackSize:     form.PackSize,
		LeadTimeDays: form.LeadTimeDays,
		IsPrimary:    form.IsPrimary == 1,
		IsActive:     c.DefaultPostForm("is_active", "1") == "1",
	}

	if err := supplierService.CreateSupplierProduct(input); err != nil {
		renderSupplierDetailPage(c, supplierService, supplierID, err.Error(), "", input)
		return
	}

	c.Redirect(http.StatusSeeOther, buildSupplierDetailURL(supplierID, "Item berhasil ditautkan ke supplier"))
}

func SupplierProductUpdate(c *gin.Context) {
	supplierID, err := strconv.Atoi(c.Param("id"))
	if err != nil || supplierID <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil || productID <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	type supplierProductEditForm struct {
		LastPrice    float64 `form:"last_price"`
		MOQ          float64 `form:"moq"`
		PackSize     float64 `form:"pack_size"`
		LeadTimeDays int     `form:"lead_time_days"`
		IsPrimary    int     `form:"is_primary"`
		IsActive     int     `form:"is_active"`
	}

	var form supplierProductEditForm
	supplierService := buildSupplierService()

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierDetailPage(c, supplierService, supplierID, "Form edit item supplier tidak lengkap", "", models.SupplierProductCreateInput{})
		return
	}

	err = supplierService.UpdateSupplierProduct(models.SupplierProductCreateInput{
		SupplierID:   supplierID,
		ProductID:    productID,
		LastPrice:    form.LastPrice,
		MOQ:          form.MOQ,
		PackSize:     form.PackSize,
		LeadTimeDays: form.LeadTimeDays,
		IsPrimary:    form.IsPrimary == 1,
		IsActive:     c.DefaultPostForm("is_active", "1") == "1",
	})
	if err != nil {
		renderSupplierDetailPage(c, supplierService, supplierID, err.Error(), "", models.SupplierProductCreateInput{})
		return
	}

	c.Redirect(http.StatusSeeOther, buildSupplierDetailURL(supplierID, "Item supplier berhasil diperbarui"))
}

func SupplierProductDelete(c *gin.Context) {
	supplierID, err := strconv.Atoi(c.Param("id"))
	if err != nil || supplierID <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil || productID <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	supplierService := buildSupplierService()
	if err := supplierService.DeleteSupplierProduct(models.SupplierProductDeleteInput{
		SupplierID: supplierID,
		ProductID:  productID,
	}); err != nil {
		renderSupplierDetailPage(c, supplierService, supplierID, err.Error(), "", models.SupplierProductCreateInput{})
		return
	}

	c.Redirect(http.StatusSeeOther, buildSupplierDetailURL(supplierID, "Item supplier berhasil dihapus"))
}

func SupplierStore(c *gin.Context) {
	type supplierForm struct {
		StoreID         int    `form:"store_id" binding:"required"`
		SupplierGroupID int    `form:"supplier_group_id"`
		SupplierCode    string `form:"supplier_code" binding:"required"`
		SupplierName    string `form:"supplier_name" binding:"required"`
		SupplierType    string `form:"supplier_type"`
		Address         string `form:"address"`
		Phone           string `form:"phone"`
		Email           string `form:"email"`
		PICName         string `form:"pic_name"`
		PaymentTermDays int    `form:"payment_term_days"`
		IsActive        int    `form:"is_active"`
	}

	var form supplierForm
	supplierService := buildSupplierService()

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierPage(c, supplierService, "Form supplier tidak lengkap", models.SupplierListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	hasStoreAccess, err := currentUserHasStoreAccess(c, form.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		renderSupplierPage(c, supplierService, "Store tidak tersedia untuk user login", models.SupplierListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	err = supplierService.CreateSupplier(models.SupplierCreateInput{
		StoreID:         form.StoreID,
		SupplierGroupID: form.SupplierGroupID,
		SupplierCode:    form.SupplierCode,
		SupplierName:    form.SupplierName,
		SupplierType:    form.SupplierType,
		Address:         form.Address,
		Phone:           form.Phone,
		Email:           form.Email,
		PICName:         form.PICName,
		PaymentTermDays: form.PaymentTermDays,
		IsActive:        form.IsActive == 1,
	})
	if err != nil {
		renderSupplierPage(c, supplierService, err.Error(), models.SupplierListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	c.Redirect(http.StatusSeeOther, "/suppliers")
}

func SupplierUpdate(c *gin.Context) {
	type supplierForm struct {
		ID              int    `form:"id" binding:"required"`
		StoreID         int    `form:"store_id" binding:"required"`
		SupplierGroupID int    `form:"supplier_group_id"`
		SupplierCode    string `form:"supplier_code" binding:"required"`
		SupplierName    string `form:"supplier_name" binding:"required"`
		SupplierType    string `form:"supplier_type"`
		Address         string `form:"address"`
		Phone           string `form:"phone"`
		Email           string `form:"email"`
		PICName         string `form:"pic_name"`
		PaymentTermDays int    `form:"payment_term_days"`
		IsActive        int    `form:"is_active"`
	}

	var form supplierForm
	supplierService := buildSupplierService()

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierPage(c, supplierService, "Form supplier tidak lengkap", models.SupplierListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	hasStoreAccess, err := currentUserHasStoreAccess(c, form.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		renderSupplierPage(c, supplierService, "Store tidak tersedia untuk user login", models.SupplierListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	err = supplierService.UpdateSupplier(models.SupplierUpdateInput{
		ID:              form.ID,
		StoreID:         form.StoreID,
		SupplierGroupID: form.SupplierGroupID,
		SupplierCode:    form.SupplierCode,
		SupplierName:    form.SupplierName,
		SupplierType:    form.SupplierType,
		Address:         form.Address,
		Phone:           form.Phone,
		Email:           form.Email,
		PICName:         form.PICName,
		PaymentTermDays: form.PaymentTermDays,
		IsActive:        form.IsActive == 1,
	})
	if err != nil {
		renderSupplierPage(c, supplierService, err.Error(), models.SupplierListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	c.Redirect(http.StatusSeeOther, "/suppliers")
}

func SupplierDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	supplierService := buildSupplierService()
	if err := supplierService.DeleteSupplier(id); err != nil {
		renderSupplierPage(c, supplierService, err.Error(), models.SupplierListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	c.Redirect(http.StatusSeeOther, "/suppliers")
}

func buildSupplierService() *services.SupplierService {
	supplierRepo := &repositories.SupplierRepository{DB: config.DB}
	return &services.SupplierService{Repo: supplierRepo}
}

func buildStoreService() *services.StoreService {
	storeRepo := &repositories.StoreRepository{DB: config.DB}
	return &services.StoreService{Repo: storeRepo}
}

func activeStoresOnly(stores []models.Store) []models.Store {
	active := make([]models.Store, 0, len(stores))
	for _, store := range stores {
		if store.IsActive {
			active = append(active, store)
		}
	}
	return active
}

func renderSupplierPage(c *gin.Context, supplierService *services.SupplierService, message string, filter models.SupplierListFilter) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 150
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

	groups, err := supplierService.GetSupplierGroups()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stores, err := getStoreOptionsForCurrentUser(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	allowedStoreSet := buildStoreAccessSet(stores)
	if !currentUserHasRole(c, "super-admin") {
		groups = filterSupplierGroupsByStoreAccess(groups, allowedStoreSet)
	}

	pagination := buildSupplierPagination(filter, totalItems)

	Render(c, "supplier.html", gin.H{
		"Title":      "Direktori Supplier",
		"Page":       "supplier",
		"suppliers":  suppliers,
		"Stats":      stats,
		"Groups":     groups,
		"Stores":     stores,
		"Filters":    filter,
		"Pagination": pagination,
		"Error":      message,
	})
}

func buildSupplierPagination(filter models.SupplierListFilter, totalItems int) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: filter.Page,
		PageSize:    filter.Limit,
		TotalItems:  totalItems,
	}

	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 150
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
		pagination.PrevURL = buildSupplierPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildSupplierPageURL(filter, pagination.CurrentPage+1)
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
			URL:    buildSupplierPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildSupplierPageURL(filter models.SupplierListFilter, page int) string {
	values := url.Values{}
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	if filter.Status != "" {
		values.Set("status", filter.Status)
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
		return "/suppliers"
	}
	return "/suppliers?" + encoded
}

func renderSupplierDetailPage(c *gin.Context, supplierService *services.SupplierService, id int, errorMessage string, successMessage string, form models.SupplierProductCreateInput) {
	supplier, err := supplierService.GetSupplierByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Supplier not found",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	products, err := supplierService.GetSuppliedProducts(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	productOptions, err := supplierService.GetAvailableProductOptions(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "supplier_detail.html", gin.H{
		"Title":          supplier.SupplierName,
		"Page":           "supplierDetail",
		"Supplier":       supplier,
		"Products":       products,
		"ProductOptions": productOptions,
		"Form":           form,
		"Error":          errorMessage,
		"Success":        successMessage,
	})
}

func buildSupplierDetailURL(id int, successMessage string) string {
	if successMessage == "" {
		return fmt.Sprintf("/suppliers/%d", id)
	}

	values := url.Values{}
	values.Set("success", successMessage)
	return fmt.Sprintf("/suppliers/%d?%s", id, values.Encode())
}

func getStoreOptionsForCurrentUser(c *gin.Context) ([]models.Store, error) {
	if currentUserHasRole(c, "super-admin") {
		stores, err := buildStoreService().GetStores()
		if err != nil {
			return nil, err
		}
		return activeStoresOnly(stores), nil
	}

	return buildStockCheckSessionService().GetStoreOptionsByUserID(extractCurrentUserID(c))
}

func currentUserHasStoreAccess(c *gin.Context, storeID int) (bool, error) {
	stores, err := getStoreOptionsForCurrentUser(c)
	if err != nil {
		return false, err
	}
	return isStoreAccessible(buildStoreAccessSet(stores), storeID), nil
}

func filterSupplierGroupsByStoreAccess(groups []models.SupplierGroup, allowedStores map[int]struct{}) []models.SupplierGroup {
	if len(groups) == 0 {
		return []models.SupplierGroup{}
	}
	filtered := make([]models.SupplierGroup, 0, len(groups))
	for _, group := range groups {
		if isStoreAccessible(allowedStores, group.StoreID) {
			filtered = append(filtered, group)
		}
	}
	return filtered
}
