package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func DashboardIndex(c *gin.Context) {
	dashboardService := buildDashboardService()
	currentUserID := extractCurrentUserID(c)
	dashboard, err := dashboardService.GetBuyerDashboard(currentUserID, extractCurrentUserName(c))
	if err != nil {
		c.String(500, err.Error())
		return
	}

	Render(c, "dashboard.html", gin.H{
		"Title":       "Dashboard Buyer",
		"Page":        "dashboard",
		"Dashboard":   dashboard,
		"CurrentUser": extractCurrentUserName(c),
		"CurrentRole": extractCurrentUserRole(c),
	})
}

func buildDashboardService() *services.DashboardService {
	repo := &repositories.DashboardRepository{DB: config.DB}
	return &services.DashboardService{Repo: repo}
}

func extractCurrentUserName(c *gin.Context) string {
	session := sessions.Default(c)
	if user := session.Get("user"); user != nil {
		switch val := user.(type) {
		case models.SessionUser:
			return val.Name
		case map[string]interface{}:
			if name, ok := val["name"].(string); ok {
				return name
			}
		case gin.H:
			if name, ok := val["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}

func extractCurrentUserRole(c *gin.Context) string {
	session := sessions.Default(c)
	if role, ok := session.Get("role").(string); ok {
		return role
	}
	return "Buyer"
}
