package services

import (
	"encoding/json"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type DashboardService struct {
	Repo *repositories.DashboardRepository
}

func (s *DashboardService) GetBuyerDashboard(userID int, userName string) (models.BuyerDashboardPage, error) {
	now := currentJakartaTime()
	weekStart, weekEnd := buyerDashboardWeekRange(now)
	poTrendStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -11, 0)

	metrics, err := s.Repo.GetBuyerDashboardMetrics(userID, now.Format("2006-01-02"))
	if err != nil {
		return models.BuyerDashboardPage{}, err
	}

	pipeline, err := s.Repo.GetBuyerDashboardPipeline(userID)
	if err != nil {
		return models.BuyerDashboardPage{}, err
	}

	monthlyPOCounts, err := s.Repo.GetBuyerMonthlyPOCounts(userID, poTrendStart)
	if err != nil {
		return models.BuyerDashboardPage{}, err
	}

	priorities, err := s.Repo.GetBuyerPriorityItems(userID, 6)
	if err != nil {
		return models.BuyerDashboardPage{}, err
	}

	weeklySuppliers, err := s.Repo.GetBuyerWeeklySuppliers(userID, weekStart, weekEnd, 10)
	if err != nil {
		return models.BuyerDashboardPage{}, err
	}

	sessions, err := s.Repo.GetBuyerSessionQueues(userID, 6)
	if err != nil {
		return models.BuyerDashboardPage{}, err
	}

	dashboard := models.BuyerDashboardPage{
		GreetingName:      dashboardGreetingName(userName),
		DateLabel:         formatBuyerDashboardDate(now),
		WeekRangeLabel:    formatBuyerDashboardWeekRange(weekStart, weekEnd),
		MetricCards:       buildBuyerMetricCards(metrics),
		HeroHighlights:    buildBuyerHighlights(metrics),
		Stages:            buildBuyerStages(pipeline),
		POTrendLabelsJSON: buildBuyerPOTrendLabelsJSON(now),
		POTrendSeriesJSON: buildBuyerPOTrendSeriesJSON(now, monthlyPOCounts),
		Priorities:        priorities,
		WeeklySuppliers:   weeklySuppliers,
		Sessions:          sessions,
	}

	return dashboard, nil
}

func dashboardGreetingName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Buyer"
	}
	return name
}

func formatBuyerDashboardDate(now time.Time) string {
	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	month := months[int(now.Month())-1]
	return fmt.Sprintf("%02d %s %d", now.Day(), month, now.Year())
}

func currentJakartaTime() time.Time {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Now()
	}
	return time.Now().In(location)
}

func buyerDashboardWeekRange(now time.Time) (time.Time, time.Time) {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
	end := start.AddDate(0, 0, 6)
	return start, end
}

func formatBuyerDashboardWeekRange(start time.Time, end time.Time) string {
	return fmt.Sprintf("%s - %s", formatBuyerDashboardDate(start), formatBuyerDashboardDate(end))
}

func buildBuyerMetricCards(metrics models.BuyerDashboardMetricsSnapshot) []models.BuyerDashboardMetricCard {
	return []models.BuyerDashboardMetricCard{
		{
			Label:   "Sesi Perlu Review",
			Value:   fmt.Sprintf("%d", metrics.ReviewSessionCount),
			Caption: "Sesi dengan SKU yang masih menunggu keputusan buyer.",
			Icon:    "bx bx-clipboard",
			Tone:    "blue",
		},
		{
			Label:   "SKU Menunggu Keputusan",
			Value:   fmt.Sprintf("%d", metrics.PendingSKUCount),
			Caption: "Item draft atau submitted yang belum difinalkan.",
			Icon:    "bx bx-list-ul",
			Tone:    "amber",
		},
		{
			Label:   "SO Hari Ini",
			Value:   fmt.Sprintf("%d", metrics.TodaySessionCount),
			Caption: "Jumlah sesi stock opname dengan tanggal hari ini.",
			Icon:    "bx bx-calendar-check",
			Tone:    "emerald",
		},
	}
}

func buildBuyerHighlights(metrics models.BuyerDashboardMetricsSnapshot) []models.BuyerDashboardHighlight {
	return []models.BuyerDashboardHighlight{
		{
			Label: "Supplier aktif",
			Value: fmt.Sprintf("%d supplier", metrics.ActiveSupplierCount),
		},
		{
			Label: "Produk aktif",
			Value: fmt.Sprintf("%d item", metrics.ActiveProductCount),
		},
	}
}

func buildBuyerStages(snapshot models.BuyerDashboardPipelineSnapshot) []models.BuyerDashboardStage {
	total := snapshot.DraftCount + snapshot.WaitingCount + snapshot.ApprovedCount + snapshot.RejectedCount
	if total == 0 {
		total = 1
	}

	return []models.BuyerDashboardStage{
		{
			Label:           "Draft Checker",
			Count:           snapshot.DraftCount,
			Note:            "Masih butuh input atau revisi dari checker.",
			ProgressPercent: (snapshot.DraftCount * 100) / total,
			Tone:            "slate",
		},
		{
			Label:           "Menunggu Buyer",
			Count:           snapshot.WaitingCount,
			Note:            "Sudah bisa direview untuk approve, adjust, atau reject.",
			ProgressPercent: (snapshot.WaitingCount * 100) / total,
			Tone:            "amber",
		},
		{
			Label:           "Approved",
			Count:           snapshot.ApprovedCount,
			Note:            "Item siap diteruskan ke proses PO atau finalisasi.",
			ProgressPercent: (snapshot.ApprovedCount * 100) / total,
			Tone:            "emerald",
		},
		{
			Label:           "Rejected",
			Count:           snapshot.RejectedCount,
			Note:            "Perlu komunikasi ulang bila masih menjadi kebutuhan toko.",
			ProgressPercent: (snapshot.RejectedCount * 100) / total,
			Tone:            "rose",
		},
	}
}

func buildBuyerPOTrendLabelsJSON(anchorDate time.Time) string {
	labels := make([]string, 0, 12)
	start := time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -11, 0)
	for offset := 0; offset < 12; offset++ {
		labels = append(labels, start.AddDate(0, offset, 0).Format("Jan 2006"))
	}
	return mustMarshalBuyerDashboardJSON(labels)
}

func buildBuyerPOTrendSeriesJSON(anchorDate time.Time, counts map[string]int) string {
	series := make([]int, 0, 12)
	start := time.Date(anchorDate.Year(), anchorDate.Month(), 1, 0, 0, 0, 0, anchorDate.Location()).AddDate(0, -11, 0)
	for offset := 0; offset < 12; offset++ {
		monthKey := start.AddDate(0, offset, 0).Format("2006-01")
		series = append(series, counts[monthKey])
	}
	return mustMarshalBuyerDashboardJSON(series)
}

func mustMarshalBuyerDashboardJSON(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}
