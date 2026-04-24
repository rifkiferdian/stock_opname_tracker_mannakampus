package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type StockOpnameReportService struct {
	Repo *repositories.StockOpnameReportRepository
}

type stockOpnameReportSnapshot struct {
	ItemID             int
	QtyStore           float64
	QtyWarehouse       float64
	SystemQtyStore     float64
	SystemQtyWarehouse float64
	PurchaseQty        float64
	SuggestQty         float64
	ApprovedQty        float64
	Status             string
	ConditionStatus    string
	LatestSessionID    int
	CheckerNotes       string
	BuyerNotes         string
	ItemCount          int
}

type stockOpnameReportAggregation struct {
	ProductID           int
	ProductCode         string
	Barcode             string
	ProductName         string
	Brand               string
	CategoryName        string
	DefaultLeadTimeDays int
	Current             stockOpnameReportSnapshot
	HistoryByDate       map[string]stockOpnameReportSnapshot
}

type stockOpnameReportSummaryMetrics struct {
	CurrentProductCount int
	TotalSessionCount   int
	TotalSystemQty      float64
	TotalActualQty      float64
	PendingItems        int
}

type stockOpnamePOMonthlyTrend struct {
	Labels []string
	Values []float64
}

func (s *StockOpnameReportService) GetDetailPage(supplierID int, filter models.StockOpnameReportFilter) (models.StockOpnameReportPage, error) {
	if supplierID <= 0 {
		return models.StockOpnameReportPage{}, errors.New("supplier id tidak valid")
	}

	categories, err := s.Repo.GetCategoryOptions(supplierID)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	totalSessions, err := s.Repo.CountSessions(supplierID, filter.CategoryID)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	sessionDates, err := s.Repo.GetDistinctSessionDates(supplierID, filter.CategoryID, 13)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	page := models.StockOpnameReportPage{
		Filter:             filter,
		Categories:         categories,
		ReviewListURL:      fmt.Sprintf("/stock-check-sessions?supplier_id=%d", supplierID),
		CurrentDateLabel:   "-",
		HistoryDateLabels:  []string{},
		SummaryCards:       buildStockOpnameReportSummaryCards(stockOpnameReportSummaryMetrics{}, time.Time{}),
		TrendBars:          buildStockOpnameReportTrendBars(time.Now(), map[string]int{}),
		AuditFindings:      []models.StockOpnameReportAuditFinding{},
		ReportRows:         []models.StockOpnameReportRow{},
		LatestSessionCount: 0,
		TotalRows:          0,
	}

	if len(sessionDates) == 0 {
		return page, nil
	}

	currentDate := sessionDates[0]
	historyDates := []time.Time{}
	historyDepth := 0
	if len(sessionDates) > 1 {
		historyDepth = len(sessionDates) - 1
		historyDates = sessionDates[1:]
	}

	page.CurrentDateLabel = currentDate.Format("02 Jan 2006")
	page.HistoryDateLabels = make([]string, 0, historyDepth)
	for index := 1; index < len(sessionDates); index++ {
		page.HistoryDateLabels = append(page.HistoryDateLabels, sessionDates[index].Format("02 Jan 2006"))
	}

	records, err := s.Repo.GetReportRecords(supplierID, filter.CategoryID, sessionDates)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	poTrendStart := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location()).AddDate(0, -5, 0)
	poMonthlyRecords, err := s.Repo.GetProductMonthlyPORecords(supplierID, filter.CategoryID, poTrendStart)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}
	poTrendMap := buildStockOpnameProductMonthlyPOTrendMap(poMonthlyRecords, currentDate)

	rows, auditFindings, summaryMetrics := buildStockOpnameReportRows(records, historyDates, currentDate, poTrendMap)
	summaryMetrics.TotalSessionCount = totalSessions
	page.SummaryCards = buildStockOpnameReportSummaryCards(summaryMetrics, currentDate)
	page.LatestSessionCount = summaryMetrics.TotalSessionCount
	page.TotalRows = len(rows)

	anchorDate := currentDate
	monthlyCounts, err := s.Repo.GetMonthlyApprovalCounts(supplierID, time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -11, 0))
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	page.TrendBars = buildStockOpnameReportTrendBars(anchorDate, monthlyCounts)
	page.AuditFindings = auditFindings
	page.ReportRows = rows
	page.Pagination = models.Pagination{
		TotalItems: len(rows),
		StartItem:  1,
		EndItem:    len(rows),
	}

	return page, nil
}

func buildStockOpnameReportRows(records []repositories.StockOpnameReportRecord, historyDates []time.Time, currentDate time.Time, poTrendMap map[int]stockOpnamePOMonthlyTrend) ([]models.StockOpnameReportRow, []models.StockOpnameReportAuditFinding, stockOpnameReportSummaryMetrics) {
	currentDateKey := currentDate.Format("2006-01-02")
	productMap := make(map[int]*stockOpnameReportAggregation)
	metrics := stockOpnameReportSummaryMetrics{}

	for _, record := range records {
		aggregate, exists := productMap[record.ProductID]
		if !exists {
			aggregate = &stockOpnameReportAggregation{
				ProductID:           record.ProductID,
				ProductCode:         record.ProductCode,
				Barcode:             record.Barcode,
				ProductName:         record.ProductName,
				Brand:               record.Brand,
				CategoryName:        reportDefaultText(record.CategoryName, "Tanpa Kategori"),
				DefaultLeadTimeDays: record.DefaultLeadTimeDays,
				HistoryByDate:       map[string]stockOpnameReportSnapshot{},
			}
			productMap[record.ProductID] = aggregate
		}

		dateKey := record.SessionDate.Format("2006-01-02")
		snapshot := aggregate.HistoryByDate[dateKey]
		snapshot.ItemCount++
		if snapshot.ItemID == 0 {
			snapshot.ItemID = record.ItemID
			snapshot.SuggestQty = record.SuggestBuyQty
			snapshot.ApprovedQty = record.ApprovedBuyQty
			snapshot.CheckerNotes = strings.TrimSpace(record.CheckerNotes)
			snapshot.BuyerNotes = strings.TrimSpace(record.BuyerNotes)
		}
		snapshot.QtyStore += record.QtyStore
		snapshot.QtyWarehouse += record.QtyWarehouse
		snapshot.SystemQtyStore += record.SystemQtyStore
		snapshot.SystemQtyWarehouse += record.SystemQtyWarehouse
		snapshot.PurchaseQty += stockOpnamePurchaseQty(record)
		if snapshot.Status == "" {
			snapshot.Status = record.Status
		}
		if snapshot.ConditionStatus == "" {
			snapshot.ConditionStatus = record.ConditionStatus
		}
		if record.SessionID > snapshot.LatestSessionID {
			snapshot.LatestSessionID = record.SessionID
		}
		aggregate.HistoryByDate[dateKey] = snapshot

		if dateKey == currentDateKey {
			metrics.TotalSystemQty += record.SystemQtyStore + record.SystemQtyWarehouse
			metrics.TotalActualQty += record.QtyStore + record.QtyWarehouse
			if stockOpnameIsPending(record.Status) {
				metrics.PendingItems++
			}
		}
	}

	rows := make([]models.StockOpnameReportRow, 0, len(productMap))
	findings := make([]stockOpnameAuditCandidate, 0, len(productMap))

	for _, aggregate := range productMap {
		current := aggregate.HistoryByDate[currentDateKey]
		aggregate.Current = current

		if current.QtyStore > 0 || current.QtyWarehouse > 0 || current.PurchaseQty > 0 || current.LatestSessionID > 0 {
			metrics.CurrentProductCount++
		}
		row := models.StockOpnameReportRow{
			ProductID:          aggregate.ProductID,
			Name:               aggregate.ProductName,
			SKU:                aggregate.ProductCode,
			Barcode:            reportDefaultText(aggregate.Barcode, "-"),
			Brand:              reportDefaultText(aggregate.Brand, "-"),
			CategoryName:       aggregate.CategoryName,
			LeadTimeLabel:      fmt.Sprintf("%d hari", aggregate.DefaultLeadTimeDays),
			AvatarText:         stockOpnameReportInitials(aggregate.ProductName),
			AvatarClass:        stockOpnameReportAvatarClass(aggregate.ProductName),
			CurrentShop:        reportFormatSnapshotWholeNumber(current.LatestSessionID > 0, current.QtyStore),
			CurrentWH:          reportFormatSnapshotWholeNumber(current.LatestSessionID > 0, current.QtyWarehouse),
			CurrentPO:          reportFormatSnapshotWholeNumber(current.LatestSessionID > 0, current.PurchaseQty),
			CurrentStatus:      stockOpnameReportStatusLabel(current.Status),
			StatusTone:         stockOpnameReportStatusTone(current.Status),
			CurrentSessionID:   current.LatestSessionID,
			CurrentItemID:      current.ItemID,
			CurrentSuggestQty:  reportFormatSnapshotWholeNumber(current.ItemID > 0, current.SuggestQty),
			CurrentCheckerNote: reportDefaultText(current.CheckerNotes, "-"),
			CurrentBuyerNotes:  current.BuyerNotes,
			CurrentApproveSeed: stockOpnameReportApproveSeed(current),
			CanInlineApprove:   current.ItemCount == 1 && current.ItemID > 0 && current.LatestSessionID > 0,
			ActionLabel:        "Lihat Sesi",
			LatestSessionID:    current.LatestSessionID,
			LatestSessionDate:  currentDate.Format("02 Jan 2006"),
		}
		if current.LatestSessionID > 0 {
			row.ActionURL = fmt.Sprintf("/stock-check-sessions/%d", current.LatestSessionID)
		} else {
			row.ActionURL = fmt.Sprintf("/products/%d", aggregate.ProductID)
		}

		if trend, ok := poTrendMap[aggregate.ProductID]; ok {
			row.POTrendSeriesJSON = mustMarshalStockOpnameReportJSON(trend.Values)
			row.POTrendLabelsJSON = mustMarshalStockOpnameReportJSON(trend.Labels)
			row.POTrendTotalLabel = reportFormatWholeNumber(sumStockOpnameTrendValues(trend.Values))
		} else {
			row.POTrendSeriesJSON = "[0,0,0,0,0,0]"
			row.POTrendLabelsJSON = `["-","-","-","-","-","-"]`
			row.POTrendTotalLabel = "0"
		}

		row.History = make([]models.StockOpnameReportHistoryPoint, 0, len(historyDates))
		for _, historyDate := range historyDates {
			historyDateKey := historyDate.Format("2006-01-02")
			history := aggregate.HistoryByDate[historyDateKey]
			row.History = append(row.History, models.StockOpnameReportHistoryPoint{
				DateLabel: historyDate.Format("02 Jan 2006"),
				ShopCount: reportFormatSnapshotWholeNumber(history.LatestSessionID > 0, history.QtyStore),
				WHSCount:  reportFormatSnapshotWholeNumber(history.LatestSessionID > 0, history.QtyWarehouse),
				POCount:   reportFormatSnapshotWholeNumber(history.LatestSessionID > 0, history.PurchaseQty),
			})
		}

		rows = append(rows, row)
		if candidate, ok := buildStockOpnameAuditCandidate(aggregate, row); ok {
			findings = append(findings, candidate)
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name)
	})
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Score == findings[j].Score {
			return strings.ToLower(findings[i].Row.Name) < strings.ToLower(findings[j].Row.Name)
		}
		return findings[i].Score > findings[j].Score
	})

	auditFindings := make([]models.StockOpnameReportAuditFinding, 0, minInt(len(findings), 5))
	for index, finding := range findings {
		if index >= 5 {
			break
		}
		auditFindings = append(auditFindings, models.StockOpnameReportAuditFinding{
			Title:     finding.Title,
			SKU:       "SKU: " + finding.Row.SKU,
			Icon:      finding.Icon,
			ActionURL: finding.Row.ActionURL,
		})
	}

	return rows, auditFindings, metrics
}

func buildStockOpnameProductMonthlyPOTrendMap(records []repositories.StockOpnameProductMonthlyPORecord, anchorDate time.Time) map[int]stockOpnamePOMonthlyTrend {
	monthKeys := make([]string, 0, 6)
	labels := make([]string, 0, 6)
	startMonth := time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -5, 0)
	for offset := 0; offset < 6; offset++ {
		currentMonth := startMonth.AddDate(0, offset, 0)
		monthKeys = append(monthKeys, currentMonth.Format("2006-01"))
		labels = append(labels, currentMonth.Format("Jan"))
	}

	valuesByProduct := make(map[int]map[string]float64)
	for _, record := range records {
		if _, ok := valuesByProduct[record.ProductID]; !ok {
			valuesByProduct[record.ProductID] = make(map[string]float64)
		}
		valuesByProduct[record.ProductID][record.MonthKey] = record.POQty
	}

	result := make(map[int]stockOpnamePOMonthlyTrend)
	for productID, monthValues := range valuesByProduct {
		values := make([]float64, 0, len(monthKeys))
		for _, monthKey := range monthKeys {
			values = append(values, monthValues[monthKey])
		}
		result[productID] = stockOpnamePOMonthlyTrend{
			Labels: labels,
			Values: values,
		}
	}

	return result
}

type stockOpnameAuditCandidate struct {
	Title string
	Icon  string
	Score int
	Row   models.StockOpnameReportRow
}

func buildStockOpnameAuditCandidate(aggregate *stockOpnameReportAggregation, row models.StockOpnameReportRow) (stockOpnameAuditCandidate, bool) {
	current := aggregate.Current
	totalSystem := current.SystemQtyStore + current.SystemQtyWarehouse
	totalActual := current.QtyStore + current.QtyWarehouse
	discrepancyPercent := 0.0
	if totalSystem > 0 {
		discrepancyPercent = math.Abs(totalActual-totalSystem) / totalSystem * 100
	}

	switch current.ConditionStatus {
	case "empty_rack":
		return stockOpnameAuditCandidate{Title: "Rak Kosong Saat Opname", Icon: "bx bx-error-alt", Score: 95, Row: row}, true
	case "damaged":
		return stockOpnameAuditCandidate{Title: "Stok Rusak Perlu Follow Up", Icon: "bx bx-error", Score: 90, Row: row}, true
	case "missing":
		return stockOpnameAuditCandidate{Title: "Item Hilang dari Stok Fisik", Icon: "bx bx-search-alt", Score: 92, Row: row}, true
	case "overstock":
		return stockOpnameAuditCandidate{Title: "Overstock di Sesi Terbaru", Icon: "bx bx-package", Score: 78, Row: row}, true
	}

	if stockOpnameIsPending(current.Status) {
		return stockOpnameAuditCandidate{Title: "Menunggu Review Buyer", Icon: "bx bx-time-five", Score: 80, Row: row}, true
	}

	if current.Status == "rejected" {
		return stockOpnameAuditCandidate{Title: "Item Ditolak pada Review Terbaru", Icon: "bx bx-x-circle", Score: 84, Row: row}, true
	}

	if discrepancyPercent >= 15 {
		return stockOpnameAuditCandidate{
			Title: fmt.Sprintf("Selisih Stok %.1f%%", discrepancyPercent),
			Icon:  "bx bx-line-chart-down",
			Score: int(math.Round(discrepancyPercent)),
			Row:   row,
		}, true
	}

	return stockOpnameAuditCandidate{}, false
}

func buildStockOpnameReportSummaryCards(metrics stockOpnameReportSummaryMetrics, currentDate time.Time) []models.StockOpnameReportSummaryCard {
	dateNote := "Belum ada sesi tersedia"
	if !currentDate.IsZero() {
		dateNote = "Snapshot " + currentDate.Format("02 Jan 2006")
	}

	pendingTone := "info"
	if metrics.PendingItems > 0 {
		pendingTone = "warning"
	}

	return []models.StockOpnameReportSummaryCard{
		{
			Label: "SKU dalam review",
			Value: fmt.Sprintf("%d", metrics.CurrentProductCount),
			Note:  dateNote,
			Tone:  "info",
		},
		{
			Label: "Jumlah SO yang sudah dilakukan",
			Value: fmt.Sprintf("%d", metrics.TotalSessionCount),
			Note:  "Total sesi stock opname untuk supplier ini",
			Tone:  "info",
		},
		{
			Label: "Menunggu Persetujuan di SO Terbaru",
			Value: fmt.Sprintf("%d", metrics.PendingItems),
			Note:  "Item belum selesai direview buyer",
			Tone:  pendingTone,
		},
	}
}

func buildStockOpnameReportTrendBars(anchorDate time.Time, counts map[string]int) []models.StockOpnameReportTrendBar {
	if anchorDate.IsZero() {
		anchorDate = time.Now()
	}

	start := time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -11, 0)
	series := make([]int, 0, 12)
	maxValue := 0
	for offset := 0; offset < 12; offset++ {
		currentMonth := start.AddDate(0, offset, 0)
		monthKey := currentMonth.Format("2006-01")
		value := counts[monthKey]
		series = append(series, value)
		if value > maxValue {
			maxValue = value
		}
	}
	if maxValue == 0 {
		maxValue = 1
	}

	bars := make([]models.StockOpnameReportTrendBar, 0, 12)
	for offset, value := range series {
		currentMonth := start.AddDate(0, offset, 0)
		height := 18 + int(math.Round((float64(value)/float64(maxValue))*72))
		bars = append(bars, models.StockOpnameReportTrendBar{
			Label:     currentMonth.Format("Jan 2006"),
			Height:    fmt.Sprintf("%d%%", height),
			IsCurrent: offset == len(series)-1,
			Value:     fmt.Sprintf("%d", value),
		})
	}

	return bars
}

func stockOpnamePurchaseQty(record repositories.StockOpnameReportRecord) float64 {
	if record.ApprovedBuyQty > 0 {
		return record.ApprovedBuyQty
	}
	if record.Status == "rejected" {
		return 0
	}
	return record.SuggestBuyQty
}

func stockOpnameIsPending(status string) bool {
	switch status {
	case "approved", "po_created", "rejected", "closed", "cancelled":
		return false
	default:
		return true
	}
}

func stockOpnameReportStatusLabel(status string) string {
	switch status {
	case "approved":
		return "Approved"
	case "po_created":
		return "PO Created"
	case "reviewed":
		return "Reviewed"
	case "submitted":
		return "Submitted"
	case "rejected":
		return "Rejected"
	case "draft":
		return "Draft"
	default:
		return "Pending"
	}
}

func stockOpnameReportStatusTone(status string) string {
	switch status {
	case "approved", "po_created":
		return "stable"
	case "rejected":
		return "low"
	default:
		return "input"
	}
}

func stockOpnameReportInitials(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "PR"
	}
	initials := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		initials += strings.ToUpper(part[:1])
		if len(initials) == 2 {
			break
		}
	}
	if initials == "" {
		return "PR"
	}
	return initials
}

func stockOpnameReportAvatarClass(name string) string {
	sum := 0
	for _, char := range name {
		sum += int(char)
	}

	switch sum % 4 {
	case 0:
		return "tone-ocean"
	case 1:
		return "tone-ember"
	case 2:
		return "tone-mint"
	default:
		return "tone-slate"
	}
}

func reportFormatWholeNumber(value float64) string {
	return fmt.Sprintf("%.0f", math.Round(value))
}

func stockOpnameReportApproveSeed(snapshot stockOpnameReportSnapshot) string {
	value := snapshot.SuggestQty
	if snapshot.Status == "rejected" {
		value = 0
	} else if snapshot.ApprovedQty > 0 {
		value = snapshot.ApprovedQty
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func reportFormatSnapshotWholeNumber(exists bool, value float64) string {
	if !exists {
		return "-"
	}
	return reportFormatWholeNumber(value)
}

func mustMarshalStockOpnameReportJSON(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func sumStockOpnameTrendValues(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}

func reportDefaultText(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
