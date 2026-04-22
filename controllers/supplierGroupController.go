package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SupplierGroupIndex(c *gin.Context) {
	supplierGroupService := buildSupplierGroupService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	renderSupplierGroupPage(c, supplierGroupService, "", models.SupplierGroupListFilter{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Sort:   c.DefaultQuery("sort", "recent"),
		Page:   page,
		Limit:  10,
	})
}

func SupplierGroupStore(c *gin.Context) {
	type supplierGroupForm struct {
		GroupCode   string `form:"group_code" binding:"required"`
		GroupName   string `form:"group_name" binding:"required"`
		Description string `form:"description"`
		IsActive    int    `form:"is_active"`
	}

	var form supplierGroupForm
	supplierGroupService := buildSupplierGroupService()

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierGroupPage(c, supplierGroupService, "Form supplier group tidak lengkap", models.SupplierGroupListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	err := supplierGroupService.CreateSupplierGroup(models.SupplierGroupCreateInput{
		GroupCode:   form.GroupCode,
		GroupName:   form.GroupName,
		Description: form.Description,
		IsActive:    form.IsActive == 1,
	})
	if err != nil {
		renderSupplierGroupPage(c, supplierGroupService, err.Error(), models.SupplierGroupListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	c.Redirect(http.StatusSeeOther, "/supplier-groups")
}

func SupplierGroupUpdate(c *gin.Context) {
	type supplierGroupForm struct {
		ID          int    `form:"id" binding:"required"`
		GroupCode   string `form:"group_code" binding:"required"`
		GroupName   string `form:"group_name" binding:"required"`
		Description string `form:"description"`
		IsActive    int    `form:"is_active"`
	}

	var form supplierGroupForm
	supplierGroupService := buildSupplierGroupService()

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierGroupPage(c, supplierGroupService, "Form supplier group tidak lengkap", models.SupplierGroupListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	err := supplierGroupService.UpdateSupplierGroup(models.SupplierGroupUpdateInput{
		ID:          form.ID,
		GroupCode:   form.GroupCode,
		GroupName:   form.GroupName,
		Description: form.Description,
		IsActive:    form.IsActive == 1,
	})
	if err != nil {
		renderSupplierGroupPage(c, supplierGroupService, err.Error(), models.SupplierGroupListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	c.Redirect(http.StatusSeeOther, "/supplier-groups")
}

func SupplierGroupDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier group id")
		return
	}

	supplierGroupService := buildSupplierGroupService()
	if err := supplierGroupService.DeleteSupplierGroup(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/supplier-groups")
}

func buildSupplierGroupService() *services.SupplierGroupService {
	supplierGroupRepo := &repositories.SupplierGroupRepository{DB: config.DB}
	return &services.SupplierGroupService{Repo: supplierGroupRepo}
}

func renderSupplierGroupPage(c *gin.Context, supplierGroupService *services.SupplierGroupService, message string, filter models.SupplierGroupListFilter) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	groups, totalItems, err := supplierGroupService.GetSupplierGroups(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stats, err := supplierGroupService.GetSupplierGroupStats()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	pagination := buildSupplierGroupPagination(filter, totalItems)

	Render(c, "supplier_group.html", gin.H{
		"Title":      "Master Supplier Group",
		"Page":       "supplierGroup",
		"Groups":     groups,
		"Stats":      stats,
		"Filters":    filter,
		"Pagination": pagination,
		"Error":      message,
	})
}

func buildSupplierGroupPagination(filter models.SupplierGroupListFilter, totalItems int) models.Pagination {
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
		pagination.PrevURL = buildSupplierGroupPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildSupplierGroupPageURL(filter, pagination.CurrentPage+1)
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
			URL:    buildSupplierGroupPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildSupplierGroupPageURL(filter models.SupplierGroupListFilter, page int) string {
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
		return "/supplier-groups"
	}
	return "/supplier-groups?" + encoded
}
