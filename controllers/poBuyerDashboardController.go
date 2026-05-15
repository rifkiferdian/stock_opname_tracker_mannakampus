package controllers

import (
	"net/http"
	"time"

	"gobase-app/models"

	"github.com/gin-gonic/gin"
)

func POBuyerDashboardIndex(c *gin.Context) {
	if !currentUserHasRole(c, "po-buyer") {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	currentUserID := extractCurrentUserID(c)
	if currentUserID <= 0 {
		c.Redirect(http.StatusFound, "/logout")
		return
	}

	sessionService := buildStockCheckSessionService()
	queryFilter := models.StockCheckSessionListFilter{
		Status: "closed",
		Page:   1,
		Limit:  500,
	}

	sessions, _, err := sessionService.GetSessions(queryFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	filtered := make([]models.StockCheckSession, 0, len(sessions))
	if currentUserHasRole(c, "super-admin") {
		filtered = sessions
	} else {
		stores, err := sessionService.GetStoreOptionsByUserID(currentUserID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		allowedStoreSet := buildStoreAccessSet(stores)
		for _, session := range sessions {
			if !isStoreAccessible(allowedStoreSet, session.StoreID) {
				continue
			}
			filtered = append(filtered, session)
		}
	}

	todayKey := time.Now().Format("2006-01-02")
	todayReadyCount := 0
	distinctSupplierSet := map[int]struct{}{}
	distinctStoreSet := map[int]struct{}{}

	for _, session := range filtered {
		if session.SessionDate == todayKey {
			todayReadyCount++
		}
		if session.SupplierID > 0 {
			distinctSupplierSet[session.SupplierID] = struct{}{}
		}
		if session.StoreID > 0 {
			distinctStoreSet[session.StoreID] = struct{}{}
		}
	}

	recentSessions := filtered
	if len(recentSessions) > 8 {
		recentSessions = recentSessions[:8]
	}

	Render(c, "po_buyer_dashboard.html", gin.H{
		"Title":                 "Dashboard PO Buyer",
		"Page":                  "po_buyer_dashboard",
		"CurrentUser":           extractCurrentUserName(c),
		"CurrentRole":           extractCurrentUserRole(c),
		"TodayLabel":            time.Now().Format("Monday, 02 Jan 2006"),
		"ReadyForPOCount":       len(filtered),
		"ReadyForPOTodayCount":  todayReadyCount,
		"DistinctSupplierCount": len(distinctSupplierSet),
		"DistinctStoreCount":    len(distinctStoreSet),
		"RecentSessions":        recentSessions,
	})
}
