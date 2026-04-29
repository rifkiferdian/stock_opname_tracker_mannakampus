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

func ProductIndex(c *gin.Context) {
	productService := buildProductService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	categoryID, _ := strconv.Atoi(c.DefaultQuery("category_id", "0"))

	renderProductPage(c, productService, "", models.ProductListFilter{
		Search:     c.Query("search"),
		CategoryID: categoryID,
		Status:     c.Query("status"),
		Brand:      c.Query("brand"),
		Sort:       c.DefaultQuery("sort", "recent"),
		Page:       page,
		Limit:      10,
	})
}

func ProductDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	productService := buildProductService()
	renderProductDetailPage(c, productService, id)
}

func ProductStore(c *gin.Context) {
	type productForm struct {
		ProductCode         string  `form:"product_code" binding:"required"`
		Barcode             string  `form:"barcode"`
		ProductName         string  `form:"product_name" binding:"required"`
		CategoryID          int     `form:"category_id"`
		UnitID              int     `form:"unit_id"`
		Brand               string  `form:"brand"`
		MinStock            float64 `form:"min_stock"`
		MaxStock            float64 `form:"max_stock"`
		ReorderPoint        float64 `form:"reorder_point"`
		DefaultLeadTimeDays int     `form:"default_lead_time_days"`
		PackSize            float64 `form:"pack_size"`
		PcsPerBox           int     `form:"pcs_per_box"`
		PcsPerCarton        int     `form:"pcs_per_carton"`
		IsActive            int     `form:"is_active"`
		SupplierID          int     `form:"supplier_id"`
		LastPrice           float64 `form:"last_price"`
	}

	var form productForm
	productService := buildProductService()

	if err := c.ShouldBind(&form); err != nil {
		renderProductPage(c, productService, "Form produk tidak lengkap", models.ProductListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	err := productService.CreateProduct(models.ProductCreateInput{
		ProductCode:         form.ProductCode,
		Barcode:             form.Barcode,
		ProductName:         form.ProductName,
		CategoryID:          form.CategoryID,
		UnitID:              form.UnitID,
		Brand:               form.Brand,
		MinStock:            form.MinStock,
		MaxStock:            form.MaxStock,
		ReorderPoint:        form.ReorderPoint,
		DefaultLeadTimeDays: form.DefaultLeadTimeDays,
		PackSize:            form.PackSize,
		PcsPerBox:           form.PcsPerBox,
		PcsPerCarton:        form.PcsPerCarton,
		IsActive:            form.IsActive == 1,
		SupplierID:          form.SupplierID,
		LastPrice:           form.LastPrice,
	})
	if err != nil {
		renderProductPage(c, productService, err.Error(), models.ProductListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	c.Redirect(http.StatusSeeOther, "/products")
}

func ProductUpdate(c *gin.Context) {
	type productForm struct {
		ID                  int     `form:"id" binding:"required"`
		ProductCode         string  `form:"product_code" binding:"required"`
		Barcode             string  `form:"barcode"`
		ProductName         string  `form:"product_name" binding:"required"`
		CategoryID          int     `form:"category_id"`
		UnitID              int     `form:"unit_id"`
		Brand               string  `form:"brand"`
		MinStock            float64 `form:"min_stock"`
		MaxStock            float64 `form:"max_stock"`
		ReorderPoint        float64 `form:"reorder_point"`
		DefaultLeadTimeDays int     `form:"default_lead_time_days"`
		PackSize            float64 `form:"pack_size"`
		PcsPerBox           int     `form:"pcs_per_box"`
		PcsPerCarton        int     `form:"pcs_per_carton"`
		IsActive            int     `form:"is_active"`
		SupplierID          int     `form:"supplier_id"`
		LastPrice           float64 `form:"last_price"`
	}

	var form productForm
	productService := buildProductService()

	if err := c.ShouldBind(&form); err != nil {
		renderProductPage(c, productService, "Form produk tidak lengkap", models.ProductListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	err := productService.UpdateProduct(models.ProductUpdateInput{
		ID:                  form.ID,
		ProductCode:         form.ProductCode,
		Barcode:             form.Barcode,
		ProductName:         form.ProductName,
		CategoryID:          form.CategoryID,
		UnitID:              form.UnitID,
		Brand:               form.Brand,
		MinStock:            form.MinStock,
		MaxStock:            form.MaxStock,
		ReorderPoint:        form.ReorderPoint,
		DefaultLeadTimeDays: form.DefaultLeadTimeDays,
		PackSize:            form.PackSize,
		PcsPerBox:           form.PcsPerBox,
		PcsPerCarton:        form.PcsPerCarton,
		IsActive:            form.IsActive == 1,
		SupplierID:          form.SupplierID,
		LastPrice:           form.LastPrice,
	})
	if err != nil {
		renderProductPage(c, productService, err.Error(), models.ProductListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	c.Redirect(http.StatusSeeOther, "/products")
}

func ProductDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	productService := buildProductService()
	if err := productService.DeleteProduct(id); err != nil {
		renderProductPage(c, productService, err.Error(), models.ProductListFilter{Sort: "recent", Page: 1, Limit: 10})
		return
	}

	c.Redirect(http.StatusSeeOther, "/products")
}

func buildProductService() *services.ProductService {
	productRepo := &repositories.ProductRepository{DB: config.DB}
	return &services.ProductService{Repo: productRepo}
}

func renderProductPage(c *gin.Context, productService *services.ProductService, message string, filter models.ProductListFilter) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	products, totalItems, err := productService.GetProducts(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stats, err := productService.GetProductStats()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	categories, err := productService.GetCategories()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	units, err := productService.GetUnits()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	suppliers, err := productService.GetSuppliers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	brands, err := productService.GetBrands()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	pagination := buildProductPagination(filter, totalItems)

	Render(c, "product.html", gin.H{
		"Title":      "Master Inventaris Produk",
		"Page":       "product",
		"products":   products,
		"Stats":      stats,
		"Categories": categories,
		"Units":      units,
		"Suppliers":  suppliers,
		"Brands":     brands,
		"Filters":    filter,
		"Pagination": pagination,
		"Error":      message,
	})
}

func renderProductDetailPage(c *gin.Context, productService *services.ProductService, id int) {
	product, err := productService.GetProductByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Produk tidak ditemukan",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	suppliers, err := productService.GetProductSupplierNetwork(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	histories, err := productService.GetProductStockHistory(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "product_detail.html", gin.H{
		"Title":     product.ProductName,
		"Page":      "product",
		"Product":   product,
		"Suppliers": suppliers,
		"Histories": histories,
	})
}

func buildProductPagination(filter models.ProductListFilter, totalItems int) models.Pagination {
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
		pagination.PrevURL = buildProductPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildProductPageURL(filter, pagination.CurrentPage+1)
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
			URL:    buildProductPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildProductPageURL(filter models.ProductListFilter, page int) string {
	values := url.Values{}
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	if filter.CategoryID > 0 {
		values.Set("category_id", strconv.Itoa(filter.CategoryID))
	}
	if filter.Status != "" {
		values.Set("status", filter.Status)
	}
	if filter.Brand != "" {
		values.Set("brand", filter.Brand)
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
		return "/products"
	}
	return "/products?" + encoded
}
