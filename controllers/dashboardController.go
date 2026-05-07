package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func DashboardIndex(c *gin.Context) {
	if currentUserHasRole(c, "checker") {
		c.Redirect(http.StatusFound, "/checker/dashboard")
		return
	}

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
	roles := extractCurrentUserRoles(c)
	if len(roles) == 0 {
		return "Buyer"
	}

	return strings.Join(roles, ", ")
}

func currentUserHasRole(c *gin.Context, role string) bool {
	target := normalizeRoleToken(role)
	if target == "" {
		return false
	}

	for _, currentRole := range extractCurrentUserRoles(c) {
		if normalizeRoleToken(currentRole) == target {
			return true
		}
	}

	return false
}

func extractCurrentUserRoles(c *gin.Context) []string {
	return parseRoleList(extractRoleStringFromSession(sessions.Default(c)))
}

func resolveDashboardPathByRole(roleRaw string) string {
	for _, role := range parseRoleList(roleRaw) {
		if normalizeRoleToken(role) == "checker" {
			return "/checker/dashboard"
		}
	}

	return "/dashboard"
}

func extractRoleStringFromSession(session sessions.Session) string {
	if session == nil {
		return ""
	}

	if role, ok := session.Get("role").(string); ok && strings.TrimSpace(role) != "" {
		return role
	}

	userRaw := session.Get("user")
	if userRaw == nil {
		return ""
	}

	switch val := userRaw.(type) {
	case models.SessionUser:
		return val.Role
	case map[string]interface{}:
		if role, ok := val["role"].(string); ok {
			return role
		}
		if role, ok := val["Role"].(string); ok {
			return role
		}
	case gin.H:
		if role, ok := val["role"].(string); ok {
			return role
		}
		if role, ok := val["Role"].(string); ok {
			return role
		}
	}

	return ""
}

func parseRoleList(roleRaw string) []string {
	roleRaw = strings.TrimSpace(roleRaw)
	if roleRaw == "" {
		return []string{}
	}

	parts := strings.Split(roleRaw, ",")
	roles := make([]string, 0, len(parts))
	for _, part := range parts {
		role := strings.TrimSpace(part)
		if role == "" {
			continue
		}
		roles = append(roles, role)
	}

	return roles
}

func normalizeRoleToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
