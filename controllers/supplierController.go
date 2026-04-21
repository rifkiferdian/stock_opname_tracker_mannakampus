package controllers

import (
	"database/sql"
	"errors"
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
		Type:   c.Query("type"),
		Sort:   c.DefaultQuery("sort", "recent"),
		Page:   page,
		Limit:  10,
	})
}

func SupplierDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	supplierService := buildSupplierService()
	renderSupplierDetailPage(c, supplierService, id)
}

func SupplierStore(c *gin.Context) {
	type supplierForm struct {
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
		renderSupplierPage(c, supplierService, "Form supplier tidak lengkap", models.SupplierListFilter{Sort: "recent"})
		return
	}

	err := supplierService.CreateSupplier(models.SupplierCreateInput{
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
		renderSupplierPage(c, supplierService, err.Error(), models.SupplierListFilter{Sort: "recent"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/suppliers")
}

func SupplierUpdate(c *gin.Context) {
	type supplierForm struct {
		ID              int    `form:"id" binding:"required"`
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
		renderSupplierPage(c, supplierService, "Form supplier tidak lengkap", models.SupplierListFilter{Sort: "recent"})
		return
	}

	err := supplierService.UpdateSupplier(models.SupplierUpdateInput{
		ID:              form.ID,
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
		renderSupplierPage(c, supplierService, err.Error(), models.SupplierListFilter{Sort: "recent"})
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
		renderSupplierPage(c, supplierService, err.Error(), models.SupplierListFilter{Sort: "recent"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/suppliers")
}

func buildSupplierService() *services.SupplierService {
	supplierRepo := &repositories.SupplierRepository{DB: config.DB}
	return &services.SupplierService{Repo: supplierRepo}
}

func renderSupplierPage(c *gin.Context, supplierService *services.SupplierService, message string, filter models.SupplierListFilter) {
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

	groups, err := supplierService.GetSupplierGroups()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	types, err := supplierService.GetSupplierTypes()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	pagination := buildSupplierPagination(filter, totalItems)

	Render(c, "supplier.html", gin.H{
		"Title":      "Supplier Directory",
		"Page":       "supplier",
		"suppliers":  suppliers,
		"Stats":      stats,
		"Groups":     groups,
		"Types":      types,
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
		return "/suppliers"
	}
	return "/suppliers?" + encoded
}

func renderSupplierDetailPage(c *gin.Context, supplierService *services.SupplierService, id int) {
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

	Render(c, "supplier_detail.html", gin.H{
		"Title":    supplier.SupplierName,
		"Page":     "supplierDetail",
		"Supplier": supplier,
		"Products": products,
	})
}
