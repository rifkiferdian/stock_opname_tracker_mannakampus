package controllers

import "github.com/gin-gonic/gin"

func SupplierIndex(c *gin.Context) {
	Render(c, "supplier.html", gin.H{
		"Title": "Supplier Directory",
		"Page":  "supplier",
	})
}

func SupplierDetail(c *gin.Context) {
	Render(c, "supplier_detail.html", gin.H{
		"Title": "Global Logistics Inc.",
		"Page":  "supplierDetail",
	})
}
