package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"gobase-app/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CheckerDashboardIndex(c *gin.Context) {
	if !currentUserHasRole(c, "checker") {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	currentUserID := extractCurrentUserID(c)
	if currentUserID <= 0 {
		c.Redirect(http.StatusFound, "/logout")
		return
	}

	sessionService := buildStockCheckSessionService()
	supplierService := buildSupplierService()

	today := time.Now()
	todayDate := today.Format("2006-01-02")

	stores, err := sessionService.GetStoreOptionsByUserID(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	suppliers, _, err := supplierService.GetSuppliers(models.SupplierListFilter{
		Status: "active",
		Sort:   "recent",
		Page:   1,
		Limit:  6,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	sessionsToday, _, err := sessionService.GetSessions(models.StockCheckSessionListFilter{
		DateFrom: todayDate,
		DateTo:   todayDate,
		Page:     1,
		Limit:    200,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	allowedStoreSet := extractCurrentUserStoreIDSet(c)
	filteredSessions := make([]models.StockCheckSession, 0, len(sessionsToday))
	for _, session := range sessionsToday {
		if len(allowedStoreSet) > 0 {
			if _, ok := allowedStoreSet[session.StoreID]; !ok {
				continue
			}
		}
		filteredSessions = append(filteredSessions, session)
	}

	var (
		inProgressCount int
		submittedCount  int
		closedCount     int
		nextSessionURL  = "/stock-checker"
	)

	for _, session := range filteredSessions {
		switch session.Status {
		case "draft", "in_progress":
			inProgressCount++
		case "submitted", "reviewed":
			submittedCount++
		case "closed":
			closedCount++
		}

		if nextSessionURL == "/stock-checker" && session.Status != "closed" && session.Status != "cancelled" {
			nextSessionURL = "/stock-checker/sessions/" + strconv.Itoa(session.ID) + "/input?location=store"
		}
	}

	visibleSessions := filteredSessions
	if len(visibleSessions) > 6 {
		visibleSessions = visibleSessions[:6]
	}

	Render(c, "checker_dashboard.html", gin.H{
		"Title":              "Dashboard Checker",
		"Page":               "checker_dashboard",
		"TodayLabel":         today.Format("Monday, 02 Jan 2006"),
		"CurrentUser":        extractCurrentUserName(c),
		"CurrentRole":        extractCurrentUserRole(c),
		"AssignedStoreCount": len(stores),
		"TodaySessionCount":  len(filteredSessions),
		"InProgressCount":    inProgressCount,
		"SubmittedCount":     submittedCount,
		"ClosedCount":        closedCount,
		"Suppliers":          suppliers,
		"Sessions":           visibleSessions,
		"NextSessionURL":     nextSessionURL,
	})
}

func extractCurrentUserStoreIDSet(c *gin.Context) map[int]struct{} {
	session := sessions.Default(c)
	storeRaw := ""

	if user := session.Get("user"); user != nil {
		switch val := user.(type) {
		case models.SessionUser:
			storeRaw = val.StoreID
		case map[string]interface{}:
			if store, ok := val["store_id"].(string); ok {
				storeRaw = store
			} else if store, ok := val["StoreID"].(string); ok {
				storeRaw = store
			}
		case gin.H:
			if store, ok := val["store_id"].(string); ok {
				storeRaw = store
			} else if store, ok := val["StoreID"].(string); ok {
				storeRaw = store
			}
		}
	}

	result := map[int]struct{}{}
	for _, token := range strings.Split(storeRaw, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(token))
		if err != nil || id <= 0 {
			continue
		}
		result[id] = struct{}{}
	}

	return result
}
