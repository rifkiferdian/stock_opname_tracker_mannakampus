package controllers

import (
	"net/http"
	"strconv"
	"time"

	"gobase-app/models"

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

	today := time.Now()
	todayDate := today.Format("2006-01-02")

	stores, err := sessionService.GetStoreOptionsByUserID(currentUserID)
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

	allowedStoreSet := buildStoreAccessSet(stores)
	filteredSessions := make([]models.StockCheckSession, 0, len(sessionsToday))
	for _, session := range sessionsToday {
		if !isStoreAccessible(allowedStoreSet, session.StoreID) {
			continue
		}
		filteredSessions = append(filteredSessions, session)
	}

	recentSOSessions, _, err := sessionService.GetSessions(models.StockCheckSessionListFilter{
		Page:  1,
		Limit: 250,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
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

	recentSOSuppliers := make([]models.Supplier, 0, 5)
	seenSupplier := map[int]struct{}{}
	for _, session := range recentSOSessions {
		if !isStoreAccessible(allowedStoreSet, session.StoreID) {
			continue
		}
		if session.SupplierID <= 0 {
			continue
		}
		if _, exists := seenSupplier[session.SupplierID]; exists {
			continue
		}

		recentSOSuppliers = append(recentSOSuppliers, models.Supplier{
			ID:           session.SupplierID,
			SupplierCode: session.SupplierCode,
			SupplierName: session.SupplierName,
		})
		seenSupplier[session.SupplierID] = struct{}{}

		if len(recentSOSuppliers) >= 5 {
			break
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
		"Suppliers":          recentSOSuppliers,
		"Sessions":           visibleSessions,
		"NextSessionURL":     nextSessionURL,
	})
}
