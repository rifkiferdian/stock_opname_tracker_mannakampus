package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"gobase-app/models"

	"github.com/gin-gonic/gin"
)

type checkerRecapStatusSummary struct {
	Label string
	Value int
	Tone  string
}

type checkerRecapStoreSummary struct {
	StoreID        int
	StoreName      string
	TotalSessions  int
	DraftCount     int
	SubmittedCount int
	ClosedCount    int
}

type checkerRecapSupplierSummary struct {
	SupplierID               int
	SupplierCode             string
	SupplierName             string
	TotalSessions            int
	LatestSessionDate        string
	LatestSessionDateDisplay string
}

type checkerRecapSessionItem struct {
	ID                 int
	SessionNumber      string
	SessionDateDisplay string
	StoreName          string
	SupplierName       string
	Status             string
	StatusLabel        string
	StatusBadgeClass   string
	InputURL           string
}

func StockCheckCheckerRecapIndex(c *gin.Context) {
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
	stores, err := sessionService.GetStoreOptionsByUserID(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	defaultDateFrom := now.AddDate(0, 0, -14)
	defaultDateTo := now

	dateFrom := parseQueryDateWithFallback(c.Query("date_from"), defaultDateFrom)
	dateTo := parseQueryDateWithFallback(c.Query("date_to"), defaultDateTo)
	if dateFrom.After(dateTo) {
		dateFrom, dateTo = dateTo, dateFrom
	}

	selectedStoreID := parsePositiveInt(c.Query("store_id"), 0)
	allowedStoreSet := map[int]struct{}{}
	allowedStoreSet = buildStoreAccessSet(stores)
	if selectedStoreID > 0 {
		if !isStoreAccessible(allowedStoreSet, selectedStoreID) {
			selectedStoreID = 0
		}
	}

	selectedStatus := sanitizeStockCheckSessionStatusFilter(c.Query("status"))
	sessions, _, err := sessionService.GetSessions(models.StockCheckSessionListFilter{
		DateFrom: dateFrom.Format("2006-01-02"),
		DateTo:   dateTo.Format("2006-01-02"),
		StoreID:  selectedStoreID,
		Status:   selectedStatus,
		Page:     1,
		Limit:    1000,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	filteredSessions := make([]models.StockCheckSession, 0, len(sessions))
	for _, session := range sessions {
		if !isStoreAccessible(allowedStoreSet, session.StoreID) {
			continue
		}
		filteredSessions = append(filteredSessions, session)
	}

	todayKey := now.Format("2006-01-02")
	totalSessions := len(filteredSessions)
	todaySessions := 0
	openSessions := 0
	closedSessions := 0

	statusCounts := map[string]int{}
	storeSummaryMap := map[int]*checkerRecapStoreSummary{}
	for _, store := range stores {
		storeSummaryMap[store.StoreID] = &checkerRecapStoreSummary{
			StoreID:   store.StoreID,
			StoreName: store.StoreName,
		}
	}

	supplierSummaryMap := map[int]*checkerRecapSupplierSummary{}
	for _, session := range filteredSessions {
		if session.SessionDate == todayKey {
			todaySessions++
		}

		statusCounts[session.Status]++
		if isCheckerRecapOpenStatus(session.Status) {
			openSessions++
		}
		if session.Status == "closed" {
			closedSessions++
		}

		if storeSummary, ok := storeSummaryMap[session.StoreID]; ok {
			storeSummary.TotalSessions++
			switch session.Status {
			case "draft", "in_progress":
				storeSummary.DraftCount++
			case "submitted", "reviewed":
				storeSummary.SubmittedCount++
			case "closed":
				storeSummary.ClosedCount++
			}
		}

		if session.SupplierID <= 0 {
			continue
		}
		supplierSummary, exists := supplierSummaryMap[session.SupplierID]
		if !exists {
			supplierSummary = &checkerRecapSupplierSummary{
				SupplierID:   session.SupplierID,
				SupplierCode: session.SupplierCode,
				SupplierName: session.SupplierName,
			}
			supplierSummaryMap[session.SupplierID] = supplierSummary
		}
		supplierSummary.TotalSessions++
		if session.SessionDate > supplierSummary.LatestSessionDate {
			supplierSummary.LatestSessionDate = session.SessionDate
			supplierSummary.LatestSessionDateDisplay = session.SessionDateDisplay
		}
	}

	statusSummary := buildCheckerRecapStatusSummary(statusCounts)
	storeSummaries := make([]checkerRecapStoreSummary, 0, len(storeSummaryMap))
	for _, summary := range storeSummaryMap {
		storeSummaries = append(storeSummaries, *summary)
	}
	sort.Slice(storeSummaries, func(i, j int) bool {
		if storeSummaries[i].TotalSessions == storeSummaries[j].TotalSessions {
			return strings.ToLower(storeSummaries[i].StoreName) < strings.ToLower(storeSummaries[j].StoreName)
		}
		return storeSummaries[i].TotalSessions > storeSummaries[j].TotalSessions
	})

	supplierSummaries := make([]checkerRecapSupplierSummary, 0, len(supplierSummaryMap))
	for _, summary := range supplierSummaryMap {
		supplierSummaries = append(supplierSummaries, *summary)
	}
	sort.Slice(supplierSummaries, func(i, j int) bool {
		if supplierSummaries[i].TotalSessions == supplierSummaries[j].TotalSessions {
			return supplierSummaries[i].LatestSessionDate > supplierSummaries[j].LatestSessionDate
		}
		return supplierSummaries[i].TotalSessions > supplierSummaries[j].TotalSessions
	})
	if len(supplierSummaries) > 5 {
		supplierSummaries = supplierSummaries[:5]
	}

	recentSessions := make([]checkerRecapSessionItem, 0, len(filteredSessions))
	for index, session := range filteredSessions {
		if index >= 12 {
			break
		}
		statusLabel, badgeClass := checkerRecapStatusMeta(session.Status, session.StatusLabel)
		recentSessions = append(recentSessions, checkerRecapSessionItem{
			ID:                 session.ID,
			SessionNumber:      session.SessionNumber,
			SessionDateDisplay: session.SessionDateDisplay,
			StoreName:          session.StoreName,
			SupplierName:       session.SupplierName,
			Status:             session.Status,
			StatusLabel:        statusLabel,
			StatusBadgeClass:   badgeClass,
			InputURL:           "/stock-checker/sessions/" + strconv.Itoa(session.ID) + "/input?location=store",
		})
	}

	Render(c, "stock_check_checker_recap.html", gin.H{
		"Title":             "Rekap SO Checker",
		"Page":              "stock_check_checker_recap",
		"CurrentRole":       extractCurrentUserRole(c),
		"TodayLabel":        now.Format("Monday, 02 Jan 2006"),
		"DateFrom":          dateFrom.Format("2006-01-02"),
		"DateTo":            dateTo.Format("2006-01-02"),
		"SelectedStoreID":   selectedStoreID,
		"SelectedStatus":    selectedStatus,
		"Stores":            stores,
		"StatusSummary":     statusSummary,
		"StoreSummaries":    storeSummaries,
		"SupplierSummaries": supplierSummaries,
		"RecentSessions":    recentSessions,
		"TotalSessions":     totalSessions,
		"TodaySessions":     todaySessions,
		"OpenSessions":      openSessions,
		"ClosedSessions":    closedSessions,
	})
}

func parseQueryDateWithFallback(value string, fallback time.Time) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return fallback
	}
	return parsed
}

func isCheckerRecapOpenStatus(status string) bool {
	switch status {
	case "draft", "in_progress", "submitted", "reviewed":
		return true
	default:
		return false
	}
}

func buildCheckerRecapStatusSummary(statusCounts map[string]int) []checkerRecapStatusSummary {
	type statusMeta struct {
		Key   string
		Label string
		Tone  string
	}

	ordered := []statusMeta{
		{Key: "draft", Label: "Draft", Tone: "slate"},
		{Key: "in_progress", Label: "In Progress", Tone: "amber"},
		{Key: "submitted", Label: "Submitted", Tone: "blue"},
		{Key: "reviewed", Label: "Reviewed", Tone: "violet"},
		{Key: "closed", Label: "Closed", Tone: "emerald"},
		{Key: "cancelled", Label: "Cancelled", Tone: "rose"},
	}

	result := make([]checkerRecapStatusSummary, 0, len(ordered))
	for _, item := range ordered {
		result = append(result, checkerRecapStatusSummary{
			Label: item.Label,
			Value: statusCounts[item.Key],
			Tone:  item.Tone,
		})
	}
	return result
}

func checkerRecapStatusMeta(status string, fallback string) (string, string) {
	switch status {
	case "draft":
		return "Draft", "bg-slate-100 text-slate-600"
	case "in_progress":
		return "In Progress", "bg-amber-50 text-amber-700"
	case "submitted":
		return "Submitted", "bg-blue-50 text-blue-700"
	case "reviewed":
		return "Reviewed", "bg-violet-50 text-violet-700"
	case "closed":
		return "Closed", "bg-emerald-50 text-emerald-700"
	case "cancelled":
		return "Cancelled", "bg-rose-50 text-rose-700"
	default:
		label := strings.TrimSpace(fallback)
		if label == "" {
			label = "Unknown"
		}
		return label, "bg-slate-100 text-slate-600"
	}
}

func buildStoreAccessSet(stores []models.Store) map[int]struct{} {
	result := map[int]struct{}{}
	for _, store := range stores {
		if store.StoreID <= 0 {
			continue
		}
		result[store.StoreID] = struct{}{}
	}
	return result
}

func isStoreAccessible(allowedStores map[int]struct{}, storeID int) bool {
	if len(allowedStores) == 0 || storeID <= 0 {
		return false
	}
	_, ok := allowedStores[storeID]
	return ok
}
