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

	renderProductPage(c, productService, "", models.ProductListFilter{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Sort:   c.DefaultQuery("sort", "recent"),
		Page:   page,
		Limit:  150,
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

func ProductSupplierNetworkUpdate(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	supplierID, err := strconv.Atoi(c.Param("supplier_id"))
	if err != nil || supplierID <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	type supplierNetworkForm struct {
		LastPrice    float64 `form:"last_price"`
		MOQ          float64 `form:"moq"`
		PackSize     float64 `form:"pack_size"`
		LeadTimeDays int     `form:"lead_time_days"`
		IsPrimary    int     `form:"is_primary"`
	}

	var form supplierNetworkForm
	if err := c.ShouldBind(&form); err != nil {
		c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, "Form edit supplier tidak lengkap", ""))
		return
	}

	supplierService := buildSupplierService()
	err = supplierService.UpdateSupplierProduct(models.SupplierProductCreateInput{
		SupplierID:   supplierID,
		ProductID:    productID,
		LastPrice:    form.LastPrice,
		MOQ:          form.MOQ,
		PackSize:     form.PackSize,
		LeadTimeDays: form.LeadTimeDays,
		IsPrimary:    form.IsPrimary == 1,
	})
	if err != nil {
		c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, err.Error(), ""))
		return
	}

	c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, "", "Relasi supplier berhasil diperbarui"))
}

func ProductSupplierNetworkDelete(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	supplierID, err := strconv.Atoi(c.Param("supplier_id"))
	if err != nil || supplierID <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	supplierService := buildSupplierService()
	err = supplierService.DeleteSupplierProduct(models.SupplierProductDeleteInput{
		SupplierID: supplierID,
		ProductID:  productID,
	})
	if err != nil {
		c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, err.Error(), ""))
		return
	}

	c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, "", "Relasi supplier berhasil dihapus"))
}

func ProductSupplierNetworkStore(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	type supplierNetworkCreateForm struct {
		SupplierID   int     `form:"supplier_id"`
		LastPrice    float64 `form:"last_price"`
		MOQ          float64 `form:"moq"`
		PackSize     float64 `form:"pack_size"`
		LeadTimeDays int     `form:"lead_time_days"`
		IsPrimary    int     `form:"is_primary"`
	}

	var form supplierNetworkCreateForm
	if err := c.ShouldBind(&form); err != nil {
		c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, "Form tambah supplier tidak lengkap", ""))
		return
	}

	supplierService := buildSupplierService()
	err = supplierService.CreateSupplierProduct(models.SupplierProductCreateInput{
		SupplierID:   form.SupplierID,
		ProductID:    productID,
		LastPrice:    form.LastPrice,
		MOQ:          form.MOQ,
		PackSize:     form.PackSize,
		LeadTimeDays: form.LeadTimeDays,
		IsPrimary:    form.IsPrimary == 1,
	})
	if err != nil {
		c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, err.Error(), ""))
		return
	}

	c.Redirect(http.StatusSeeOther, buildProductDetailURL(productID, "", "Relasi supplier berhasil ditambahkan"))
}

func ProductStore(c *gin.Context) {
	type productForm struct {
		ProductCode         string  `form:"product_code" binding:"required"`
		Barcode             string  `form:"barcode"`
		BarcodeBox          string  `form:"barcode_box"`
		BarcodeCarton       string  `form:"barcode_carton"`
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
		BoxPerCarton        int     `form:"box_per_carton"`
		PcsPerCarton        int     `form:"pcs_per_carton"`
		IsActive            int     `form:"is_active"`
		SupplierID          int     `form:"supplier_id"`
		LastPrice           float64 `form:"last_price"`
	}

	var form productForm
	productService := buildProductService()

	if err := c.ShouldBind(&form); err != nil {
		renderProductPage(c, productService, "Form produk tidak lengkap", models.ProductListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	err := productService.CreateProduct(models.ProductCreateInput{
		ProductCode:         form.ProductCode,
		Barcode:             form.Barcode,
		BarcodeBox:          form.BarcodeBox,
		BarcodeCarton:       form.BarcodeCarton,
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
		BoxPerCarton:        form.BoxPerCarton,
		PcsPerCarton:        form.PcsPerCarton,
		IsActive:            form.IsActive == 1,
		SupplierID:          form.SupplierID,
		LastPrice:           form.LastPrice,
	})
	if err != nil {
		renderProductPage(c, productService, err.Error(), models.ProductListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	c.Redirect(http.StatusSeeOther, "/products")
}

func ProductUpdate(c *gin.Context) {
	type productForm struct {
		ID                  int     `form:"id" binding:"required"`
		ProductCode         string  `form:"product_code" binding:"required"`
		Barcode             string  `form:"barcode"`
		BarcodeBox          string  `form:"barcode_box"`
		BarcodeCarton       string  `form:"barcode_carton"`
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
		BoxPerCarton        int     `form:"box_per_carton"`
		PcsPerCarton        int     `form:"pcs_per_carton"`
		IsActive            int     `form:"is_active"`
		SupplierID          int     `form:"supplier_id"`
		LastPrice           float64 `form:"last_price"`
		RedirectTo          string  `form:"redirect_to"`
	}

	var form productForm
	productService := buildProductService()

	if err := c.ShouldBind(&form); err != nil {
		renderProductPage(c, productService, "Form produk tidak lengkap", models.ProductListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	err := productService.UpdateProduct(models.ProductUpdateInput{
		ID:                  form.ID,
		ProductCode:         form.ProductCode,
		Barcode:             form.Barcode,
		BarcodeBox:          form.BarcodeBox,
		BarcodeCarton:       form.BarcodeCarton,
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
		BoxPerCarton:        form.BoxPerCarton,
		PcsPerCarton:        form.PcsPerCarton,
		IsActive:            form.IsActive == 1,
		SupplierID:          form.SupplierID,
		LastPrice:           form.LastPrice,
	})
	if err != nil {
		renderProductPage(c, productService, err.Error(), models.ProductListFilter{Sort: "recent", Page: 1, Limit: 150})
		return
	}

	redirectTo := buildProductUpdateRedirectURL(form.RedirectTo, form.ID)
	c.Redirect(http.StatusSeeOther, redirectTo)
}

func ProductDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid product id")
		return
	}

	productService := buildProductService()
	if err := productService.DeleteProduct(id); err != nil {
		renderProductPage(c, productService, err.Error(), models.ProductListFilter{Sort: "recent", Page: 1, Limit: 150})
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
		filter.Limit = 150
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

	pagination := buildProductPagination(filter, totalItems)
	for i := range products {
		products[i].RowNumber = pagination.StartItem + i
	}

	Render(c, "product.html", gin.H{
		"Title":      "Master Inventaris Produk",
		"Page":       "product",
		"products":   products,
		"Stats":      stats,
		"Categories": categories,
		"Units":      units,
		"Suppliers":  suppliers,
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

	historyPage, _ := strconv.Atoi(c.DefaultQuery("history_page", "1"))
	if historyPage <= 0 {
		historyPage = 1
	}
	const historyPageSize = 50

	histories, totalHistoryItems, err := productService.GetProductStockHistory(id, historyPage, historyPageSize)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	historyPagination := buildProductDetailHistoryPagination(id, historyPage, historyPageSize, totalHistoryItems)

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

	supplierOptions, err := productService.GetSuppliers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "product_detail.html", gin.H{
		"Title":             product.ProductName,
		"Page":              "product",
		"Product":           product,
		"Error":             c.Query("error"),
		"Success":           c.Query("success"),
		"Suppliers":         suppliers,
		"Categories":        categories,
		"Units":             units,
		"SupplierOptions":   supplierOptions,
		"Histories":         histories,
		"HistoryPagination": historyPagination,
	})
}

func buildProductDetailHistoryPagination(productID int, currentPage int, pageSize int, totalItems int) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalItems:  totalItems,
	}

	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 50
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
		pagination.PrevURL = buildProductDetailHistoryPageURL(productID, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildProductDetailHistoryPageURL(productID, pagination.CurrentPage+1)
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
			URL:    buildProductDetailHistoryPageURL(productID, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildProductDetailHistoryPageURL(productID int, page int) string {
	values := url.Values{}
	if page > 1 {
		values.Set("history_page", strconv.Itoa(page))
	}

	basePath := "/products/" + strconv.Itoa(productID)
	encoded := values.Encode()
	if encoded == "" {
		return basePath
	}
	return basePath + "?" + encoded
}

func buildProductDetailURL(productID int, errorMessage string, successMessage string) string {
	values := url.Values{}
	if errorMessage != "" {
		values.Set("error", errorMessage)
	}
	if successMessage != "" {
		values.Set("success", successMessage)
	}

	basePath := "/products/" + strconv.Itoa(productID)
	encoded := values.Encode()
	if encoded == "" {
		return basePath
	}
	return basePath + "?" + encoded
}

func buildProductUpdateRedirectURL(redirectTo string, productID int) string {
	if redirectTo == "/products" {
		return redirectTo
	}
	expectedDetailURL := "/products/" + strconv.Itoa(productID)
	if redirectTo == expectedDetailURL {
		return redirectTo
	}
	return "/products"
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
		return "/products"
	}
	return "/products?" + encoded
}
