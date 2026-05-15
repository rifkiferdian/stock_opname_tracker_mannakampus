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
	QtyStoreCarton     int
	QtyStoreBox        int
	QtyStorePcs        int
	QtyStore           float64
	QtyWarehouseCarton int
	QtyWarehouseBox    int
	QtyWarehousePcs    int
	QtyWarehouse       float64
	SystemQtyStore     float64
	SystemQtyWarehouse float64
	SuggestCarton      int
	SuggestBox         int
	SuggestPcs         int
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
	PcsPerBox           int
	BoxPerCarton        int
	PcsPerCarton        int
	Current             stockOpnameReportSnapshot
	HistoryByDate       map[string]stockOpnameReportSnapshot
}

type stockOpnameReportSummaryMetrics struct {
	CurrentProductCount int
	TotalSessionCount   int
	TotalSystemQty      float64
	TotalActualQty      float64
	PendingItems        int
	SubmittedItems      int
}

type stockOpnamePOMonthlyTrend struct {
	Labels []string
	Values []float64
}

func (s *StockOpnameReportService) GetDetailPage(supplierID int, filter models.StockOpnameReportFilter) (models.StockOpnameReportPage, error) {
	if supplierID <= 0 {
		return models.StockOpnameReportPage{}, errors.New("supplier id tidak valid")
	}

	totalSessions, err := s.Repo.CountSessions(supplierID)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	sessionDates, err := s.Repo.GetDistinctSessionDates(supplierID, 21)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	page := models.StockOpnameReportPage{
		Filter:                 filter,
		CurrentStatusLabel:     stockOpnameReportStatusFilterLabel(filter.Status),
		ReviewListURL:          fmt.Sprintf("/stock-check-sessions?supplier_id=%d", supplierID),
		CurrentDateLabel:       "-",
		CurrentSessionStatus:   "-",
		HistoryDateLabels:      []string{},
		SummaryCards:           buildStockOpnameReportSummaryCards(stockOpnameReportSummaryMetrics{}, time.Time{}),
		TrendBars:              buildStockOpnameReportTrendBars(time.Now(), map[string]float64{}),
		AuditFindings:          []models.StockOpnameReportAuditFinding{},
		ReportRows:             []models.StockOpnameReportRow{},
		LatestSessionCount:     0,
		TotalRows:              0,
		ApplyAllSubmittedCount: 0,
	}

	if len(sessionDates) == 0 {
		return page, nil
	}

	currentDate := sessionDates[0]
	latestSessionStatus, err := s.Repo.GetLatestSessionStatus(supplierID)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}
	page.CurrentSessionStatus = stockOpnameReportSessionStatusLabel(latestSessionStatus)

	statuses, err := s.Repo.GetStatusOptions(supplierID, currentDate)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}
	statusOptions := buildStockOpnameReportStatusOptions(statuses)
	filter.Status = normalizeStockOpnameReportStatusFilter(filter.Status, statusOptions)
	filter.ItemName = strings.TrimSpace(filter.ItemName)
	page.Filter = filter
	page.StatusOptions = statusOptions
	page.CurrentStatusLabel = stockOpnameReportStatusFilterLabel(filter.Status)

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

	records, err := s.Repo.GetReportRecords(supplierID, filter.Status, filter.ItemName, currentDate, sessionDates)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}

	poTrendStart := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location()).AddDate(0, -12, 0)
	poMonthlyRecords, err := s.Repo.GetProductMonthlyPORecords(supplierID, filter.Status, filter.ItemName, currentDate, poTrendStart)
	if err != nil {
		return models.StockOpnameReportPage{}, err
	}
	poTrendMap := buildStockOpnameProductMonthlyPOTrendMap(poMonthlyRecords, currentDate)

	rows, auditFindings, summaryMetrics := buildStockOpnameReportRows(records, historyDates, currentDate, poTrendMap)
	summaryMetrics.TotalSessionCount = totalSessions
	page.SummaryCards = buildStockOpnameReportSummaryCards(summaryMetrics, currentDate)
	page.LatestSessionCount = summaryMetrics.TotalSessionCount
	page.TotalRows = len(rows)
	page.ApplyAllSubmittedCount = summaryMetrics.SubmittedItems

	anchorDate := currentDate
	monthlyCounts, err := s.Repo.GetMonthlyApprovalCounts(supplierID, time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -12, 0))
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
				PcsPerBox:           record.PcsPerBox,
				BoxPerCarton:        record.BoxPerCarton,
				PcsPerCarton:        record.PcsPerCarton,
				HistoryByDate:       map[string]stockOpnameReportSnapshot{},
			}
			productMap[record.ProductID] = aggregate
		}

		dateKey := record.SessionDate.Format("2006-01-02")
		snapshot := aggregate.HistoryByDate[dateKey]
		snapshot.ItemCount++
		if snapshot.ItemID == 0 {
			snapshot.ItemID = record.ItemID
			snapshot.SuggestCarton = record.SuggestBuyCarton
			snapshot.SuggestBox = record.SuggestBuyBox
			snapshot.SuggestPcs = record.SuggestBuyPcs
			snapshot.SuggestQty = record.SuggestBuyQty
			snapshot.ApprovedQty = record.ApprovedBuyQty
			snapshot.CheckerNotes = strings.TrimSpace(record.CheckerNotes)
			snapshot.BuyerNotes = strings.TrimSpace(record.BuyerNotes)
		}
		snapshot.QtyStoreCarton += record.QtyStoreCarton
		snapshot.QtyStoreBox += record.QtyStoreBox
		snapshot.QtyStorePcs += record.QtyStorePcs
		snapshot.QtyStore += record.QtyStore
		snapshot.QtyWarehouseCarton += record.QtyWarehouseCarton
		snapshot.QtyWarehouseBox += record.QtyWarehouseBox
		snapshot.QtyWarehousePcs += record.QtyWarehousePcs
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
			if record.Status == "submitted" {
				metrics.SubmittedItems++
			}
		}
	}

	rows := make([]models.StockOpnameReportRow, 0, len(productMap))
	findings := make([]stockOpnameAuditCandidate, 0, len(productMap))

	for _, aggregate := range productMap {
		current := aggregate.HistoryByDate[currentDateKey]
		aggregate.Current = current
		currentShopBreakdown := reportFormatSnapshotUnitBreakdown(current.LatestSessionID > 0, current.QtyStoreCarton, current.QtyStoreBox, current.QtyStorePcs)
		currentWHBreakdown := reportFormatSnapshotUnitBreakdown(current.LatestSessionID > 0, current.QtyWarehouseCarton, current.QtyWarehouseBox, current.QtyWarehousePcs)
		currentSuggestCarton, currentSuggestBox, currentSuggestPcs, currentSuggestBreakdown := reportResolveSuggestBreakdown(
			current.SuggestQty,
			current.SuggestCarton,
			current.SuggestBox,
			current.SuggestPcs,
			aggregate.PcsPerBox,
			aggregate.PcsPerCarton,
		)
		currentPOCarton, currentPOBox, currentPOPcs, currentPOBreakdown := reportFormatSuggestBreakdown(current.PurchaseQty, aggregate.PcsPerBox, aggregate.PcsPerCarton)

		if current.QtyStore > 0 || current.QtyWarehouse > 0 || current.PurchaseQty > 0 || current.LatestSessionID > 0 {
			metrics.CurrentProductCount++
		}
		row := models.StockOpnameReportRow{
			ProductID:               aggregate.ProductID,
			Name:                    aggregate.ProductName,
			SKU:                     aggregate.ProductCode,
			Barcode:                 reportDefaultText(aggregate.Barcode, "-"),
			Brand:                   reportDefaultText(aggregate.Brand, "-"),
			CategoryName:            aggregate.CategoryName,
			LeadTimeLabel:           fmt.Sprintf("%d hari", aggregate.DefaultLeadTimeDays),
			AvatarText:              stockOpnameReportInitials(aggregate.ProductName),
			AvatarClass:             stockOpnameReportAvatarClass(aggregate.ProductName),
			CurrentShop:             reportFormatSnapshotWholeNumber(current.LatestSessionID > 0, current.QtyStore),
			CurrentShopBreakdown:    currentShopBreakdown,
			CurrentShopCarton:       current.QtyStoreCarton,
			CurrentShopBox:          current.QtyStoreBox,
			CurrentShopPcs:          current.QtyStorePcs,
			CurrentWH:               reportFormatSnapshotWholeNumber(current.LatestSessionID > 0, current.QtyWarehouse),
			CurrentWHBreakdown:      currentWHBreakdown,
			CurrentWHCarton:         current.QtyWarehouseCarton,
			CurrentWHBox:            current.QtyWarehouseBox,
			CurrentWHPcs:            current.QtyWarehousePcs,
			CurrentPO:               reportFormatSnapshotWholeNumber(current.LatestSessionID > 0, current.PurchaseQty),
			CurrentPOBreakdown:      currentPOBreakdown,
			CurrentPOCarton:         currentPOCarton,
			CurrentPOBox:            currentPOBox,
			CurrentPOPcs:            currentPOPcs,
			CurrentStatus:           stockOpnameReportStatusLabel(current.Status),
			StatusTone:              stockOpnameReportStatusTone(current.Status),
			CurrentSessionID:        current.LatestSessionID,
			CurrentItemID:           current.ItemID,
			CurrentSuggestQty:       reportFormatSnapshotWholeNumber(current.ItemID > 0, current.SuggestQty),
			CurrentSuggestBreakdown: currentSuggestBreakdown,
			CurrentSuggestCartonRaw: current.SuggestCarton,
			CurrentSuggestCarton:    currentSuggestCarton,
			CurrentSuggestBox:       currentSuggestBox,
			CurrentSuggestPcs:       currentSuggestPcs,
			CurrentCheckerNote:      reportDefaultText(current.CheckerNotes, "-"),
			CurrentBuyerNotes:       current.BuyerNotes,
			CurrentApproveSeed:      stockOpnameReportApproveSeed(current),
			CanInlineApprove:        current.ItemCount == 1 && current.ItemID > 0 && current.LatestSessionID > 0,
			ActionLabel:             "Lihat Sesi",
			LatestSessionID:         current.LatestSessionID,
			LatestSessionDate:       currentDate.Format("02 Jan 2006"),
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
			row.POTrendSeriesJSON = "[0,0,0,0,0,0,0,0,0,0,0,0,0]"
			row.POTrendLabelsJSON = `["-","-","-","-","-","-","-","-","-","-","-","-","-"]`
			row.POTrendTotalLabel = "0"
		}

		row.History = make([]models.StockOpnameReportHistoryPoint, 0, len(historyDates))
		for _, historyDate := range historyDates {
			historyDateKey := historyDate.Format("2006-01-02")
			history := aggregate.HistoryByDate[historyDateKey]
			historySuggestCarton, historySuggestBox, historySuggestPcs, historySuggestBreakdown := reportResolveSuggestBreakdown(
				history.SuggestQty,
				history.SuggestCarton,
				history.SuggestBox,
				history.SuggestPcs,
				aggregate.PcsPerBox,
				aggregate.PcsPerCarton,
			)
			historyPOCarton, historyPOBox, historyPOPcs, historyPOBreakdown := reportFormatSuggestBreakdown(history.PurchaseQty, aggregate.PcsPerBox, aggregate.PcsPerCarton)
			row.History = append(row.History, models.StockOpnameReportHistoryPoint{
				DateLabel:        historyDate.Format("02 Jan 2006"),
				ShopCount:        reportFormatSnapshotWholeNumber(history.LatestSessionID > 0, history.QtyStore),
				ShopBreakdown:    reportFormatSnapshotUnitBreakdown(history.LatestSessionID > 0, history.QtyStoreCarton, history.QtyStoreBox, history.QtyStorePcs),
				ShopCarton:       history.QtyStoreCarton,
				ShopBox:          history.QtyStoreBox,
				ShopPcs:          history.QtyStorePcs,
				WHSCount:         reportFormatSnapshotWholeNumber(history.LatestSessionID > 0, history.QtyWarehouse),
				WHSBreakdown:     reportFormatSnapshotUnitBreakdown(history.LatestSessionID > 0, history.QtyWarehouseCarton, history.QtyWarehouseBox, history.QtyWarehousePcs),
				WHSCarton:        history.QtyWarehouseCarton,
				WHSBox:           history.QtyWarehouseBox,
				WHSPcs:           history.QtyWarehousePcs,
				POCount:          reportFormatSnapshotWholeNumber(history.LatestSessionID > 0, history.PurchaseQty),
				POBreakdown:      historyPOBreakdown,
				POCarton:         historyPOCarton,
				POBox:            historyPOBox,
				POPcs:            historyPOPcs,
				SuggestCarton:    historySuggestCarton,
				SuggestBox:       historySuggestBox,
				SuggestPcs:       historySuggestPcs,
				SuggestBreakdown: historySuggestBreakdown,
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
	monthKeys := make([]string, 0, 13)
	labels := make([]string, 0, 13)
	startMonth := time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -12, 0)
	for offset := 0; offset < 13; offset++ {
		currentMonth := startMonth.AddDate(0, offset, 0)
		monthKeys = append(monthKeys, currentMonth.Format("2006-01"))
		labels = append(labels, currentMonth.Format("Jan 06"))
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

func buildStockOpnameReportTrendBars(anchorDate time.Time, counts map[string]float64) []models.StockOpnameReportTrendBar {
	if anchorDate.IsZero() {
		anchorDate = time.Now()
	}

	start := time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -12, 0)
	series := make([]float64, 0, 13)
	maxValue := 0.0
	for offset := 0; offset < 13; offset++ {
		currentMonth := start.AddDate(0, offset, 0)
		monthKey := currentMonth.Format("2006-01")
		value := counts[monthKey]
		series = append(series, value)
		if value > maxValue {
			maxValue = value
		}
	}
	if maxValue == 0 {
		maxValue = 1.0
	}

	bars := make([]models.StockOpnameReportTrendBar, 0, 13)
	for offset, value := range series {
		currentMonth := start.AddDate(0, offset, 0)
		height := 18 + int(math.Round((value/maxValue)*72))
		bars = append(bars, models.StockOpnameReportTrendBar{
			Label:     currentMonth.Format("Jan 2006"),
			Height:    fmt.Sprintf("%d%%", height),
			IsCurrent: offset == len(series)-1,
			Value:     reportFormatWholeNumber(value),
		})
	}

	return bars
}

func stockOpnamePurchaseQty(record repositories.StockOpnameReportRecord) float64 {
	return record.ApprovedBuyQty
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

func stockOpnameReportSessionStatusLabel(status string) string {
	switch strings.TrimSpace(status) {
	case "draft":
		return "Draft"
	case "in_progress":
		return "In Progress"
	case "submitted":
		return "Submitted"
	case "reviewed":
		return "Reviewed"
	case "closed":
		return "Closed"
	case "po":
		return "PO"
	case "cancelled":
		return "Cancelled"
	default:
		return "Unknown"
	}
}

func buildStockOpnameReportStatusOptions(statuses []string) []models.StockOpnameReportStatusOption {
	baseOrder := []string{"draft", "submitted", "reviewed", "approved", "po_created", "rejected", "closed", "cancelled"}
	allowed := make(map[string]bool, len(statuses))
	for _, status := range statuses {
		status = strings.TrimSpace(status)
		if status == "" {
			continue
		}
		allowed[status] = true
	}

	options := make([]models.StockOpnameReportStatusOption, 0, len(allowed))
	for _, status := range baseOrder {
		if !allowed[status] {
			continue
		}
		options = append(options, models.StockOpnameReportStatusOption{
			Value: status,
			Label: stockOpnameReportStatusFilterLabel(status),
		})
		delete(allowed, status)
	}

	if len(allowed) > 0 {
		extra := make([]string, 0, len(allowed))
		for status := range allowed {
			extra = append(extra, status)
		}
		sort.Strings(extra)
		for _, status := range extra {
			options = append(options, models.StockOpnameReportStatusOption{
				Value: status,
				Label: stockOpnameReportStatusFilterLabel(status),
			})
		}
	}

	return options
}

func normalizeStockOpnameReportStatusFilter(status string, options []models.StockOpnameReportStatusOption) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return ""
	}
	for _, option := range options {
		if option.Value == status {
			return status
		}
	}
	return ""
}

func stockOpnameReportStatusFilterLabel(status string) string {
	switch strings.TrimSpace(status) {
	case "draft":
		return "Draft"
	case "submitted":
		return "Submitted"
	case "reviewed":
		return "Reviewed"
	case "approved":
		return "Approved"
	case "po_created":
		return "PO Created"
	case "rejected":
		return "Rejected"
	case "closed":
		return "Closed"
	case "cancelled":
		return "Cancelled"
	default:
		return "Semua status"
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
	return strconv.FormatFloat(snapshot.ApprovedQty, 'f', -1, 64)
}

func reportFormatSnapshotWholeNumber(exists bool, value float64) string {
	if !exists {
		return "-"
	}
	return reportFormatWholeNumber(value)
}

func reportFormatSnapshotUnitBreakdown(exists bool, carton int, box int, pcs int) string {
	if !exists {
		return "-"
	}
	return reportFormatUnitBreakdown(carton, box, pcs)
}

func reportFormatSnapshotSuggestBreakdown(exists bool, qty float64, pcsPerBox int, pcsPerCarton int) string {
	if !exists {
		return "-"
	}
	_, _, _, breakdown := reportFormatSuggestBreakdown(qty, pcsPerBox, pcsPerCarton)
	return breakdown
}

func reportFormatUnitBreakdown(carton int, box int, pcs int) string {
	return fmt.Sprintf("%d ctn %d box %d pcs", carton, box, pcs)
}

func reportFormatSuggestBreakdown(qty float64, pcsPerBox int, pcsPerCarton int) (int, int, int, string) {
	totalPcs := int(math.Round(qty))
	if totalPcs <= 0 {
		return 0, 0, 0, "0 pcs"
	}

	carton := 0
	box := 0
	pcs := totalPcs

	if pcsPerCarton > 0 {
		carton = pcs / pcsPerCarton
		pcs = pcs % pcsPerCarton
	}
	if pcsPerBox > 0 {
		box = pcs / pcsPerBox
		pcs = pcs % pcsPerBox
	}

	return carton, box, pcs, reportFormatUnitBreakdown(carton, box, pcs)
}

func reportResolveSuggestBreakdown(qty float64, carton int, box int, pcs int, pcsPerBox int, pcsPerCarton int) (int, int, int, string) {
	if carton == 0 && box == 0 && pcs == 0 && qty > 0 {
		return reportFormatSuggestBreakdown(qty, pcsPerBox, pcsPerCarton)
	}

	return carton, box, pcs, reportFormatUnitBreakdown(carton, box, pcs)
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
