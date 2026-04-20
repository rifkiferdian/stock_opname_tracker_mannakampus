package controllers

import "github.com/gin-gonic/gin"

func ProductIndex(c *gin.Context) {
	Render(c, "product.html", gin.H{
		"Title": "Retail Inventory Master",
		"Page":  "product",
	})
}
