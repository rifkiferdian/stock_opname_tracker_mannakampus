package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type StockCheckSessionService struct {
	Repo *repositories.StockCheckSessionRepository
}

func (s *StockCheckSessionService) GetSessions(filter models.StockCheckSessionListFilter) ([]models.StockCheckSession, int, error) {
	filter.DateFrom = sanitizeStockCheckSessionDate(filter.DateFrom)
	filter.DateTo = sanitizeStockCheckSessionDate(filter.DateTo)
	filter.Status = sanitizeStockCheckSessionStatus(filter.Status)

	if filter.StoreID < 0 {
		filter.StoreID = 0
	}
	if filter.SupplierID < 0 {
		filter.SupplierID = 0
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	totalItems, err := s.Repo.CountAll(filter)
	if err != nil {
		return nil, 0, err
	}

	if totalItems == 0 {
		return []models.StockCheckSession{}, 0, nil
	}

	totalPages := (totalItems + filter.Limit - 1) / filter.Limit
	if filter.Page > totalPages {
		filter.Page = totalPages
	}

	sessions, err := s.Repo.GetAll(filter)
	if err != nil {
		return nil, 0, err
	}

	return sessions, totalItems, nil
}

func (s *StockCheckSessionService) GetStoreOptions() ([]models.Store, error) {
	return s.Repo.GetStoreOptions()
}

func (s *StockCheckSessionService) GetSupplierOptions() ([]models.Supplier, error) {
	return s.Repo.GetSupplierOptions()
}

func (s *StockCheckSessionService) GetSessionDetailPage(id int) (models.StockCheckSessionDetailPage, error) {
	if id <= 0 {
		return models.StockCheckSessionDetailPage{}, errors.New("session id tidak valid")
	}

	session, err := s.Repo.GetByID(id)
	if err != nil {
		return models.StockCheckSessionDetailPage{}, err
	}

	items, err := s.Repo.GetReviewItems(id)
	if err != nil {
		return models.StockCheckSessionDetailPage{}, err
	}

	detail := models.StockCheckSessionDetail{
		StockCheckSession: session,
		StageLabel:        stockCheckSessionStageLabel(session.Status),
		StatusBadgeClass:  stockCheckSessionDetailStatusBadgeClass(session.Status),
		ItemCount:         len(items),
	}

	distinctSuppliers := map[string]struct{}{}
	alertItems := 0

	for _, item := range items {
		detail.TotalSuggestedQty += item.SuggestBuyQty
		detail.TotalApprovedQty += item.ApprovedBuyQty
		detail.SuggestedPurchaseValue += item.SuggestLineValue

		switch item.Status {
		case "approved", "po_created":
			detail.ApprovedItems++
			detail.FinalApprovedValue += item.ApprovedLineValue
		case "rejected":
			detail.RejectedItems++
		default:
			detail.OnHoldItems++
		}

		if item.ConditionStatus != "good" {
			alertItems++
		}
		if strings.TrimSpace(item.SelectedSupplierName) != "" && item.SelectedSupplierName != "-" {
			distinctSuppliers[item.SelectedSupplierName] = struct{}{}
		}
	}

	detail.DistinctSupplierCount = len(distinctSuppliers)
	detail.TotalSuggestedQtyDisplay = formatStockCheckSessionDecimal(detail.TotalSuggestedQty)
	detail.TotalApprovedQtyDisplay = formatStockCheckSessionDecimal(detail.TotalApprovedQty)
	detail.SuggestedPurchaseValueDisplay = formatStockCheckSessionCurrency(detail.SuggestedPurchaseValue)
	detail.FinalApprovedValueDisplay = formatStockCheckSessionCurrency(detail.FinalApprovedValue)
	if detail.SuggestedPurchaseValue > 0 {
		detail.ApprovalYieldPercent = (detail.FinalApprovedValue / detail.SuggestedPurchaseValue) * 100
	}
	detail.ApprovalYieldDisplay = fmt.Sprintf("%.1f%%", detail.ApprovalYieldPercent)

	return models.StockCheckSessionDetailPage{
		Session:       detail,
		Items:         items,
		OverviewCards: buildStockCheckSessionOverviewCards(detail, alertItems),
	}, nil
}

func (s *StockCheckSessionService) CreateSession(input models.StockCheckSessionCreateInput) error {
	sanitizeStockCheckSessionCreateInput(&input)

	if input.CreatedBy <= 0 {
		return errors.New("user login tidak valid")
	}
	if err := validateStockCheckSessionCreateInput(input); err != nil {
		return err
	}

	storeCode, err := s.Repo.GetStoreCodeByID(input.StoreID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(storeCode) == "" {
		return errors.New("store code tidak ditemukan")
	}

	prefix := fmt.Sprintf("SCS-%s-%s-", storeCode, strings.ReplaceAll(input.SessionDate, "-", ""))
	nextSequence, err := s.Repo.GetNextSessionSequence(prefix)
	if err != nil {
		return err
	}

	input.SessionNumber = fmt.Sprintf("%s%03d", prefix, nextSequence)
	return s.Repo.Create(input)
}

func (s *StockCheckSessionService) UpdateSession(input models.StockCheckSessionUpdateInput) error {
	sanitizeStockCheckSessionUpdateInput(&input)

	if input.ID <= 0 {
		return errors.New("session id tidak valid")
	}
	if err := validateStockCheckSessionUpdateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("session id %d tidak ditemukan", input.ID)
	}

	return s.Repo.Update(input)
}

func (s *StockCheckSessionService) DeleteSession(id int) error {
	if id <= 0 {
		return errors.New("session id tidak valid")
	}

	exists, err := s.Repo.ExistsByID(id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("session id %d tidak ditemukan", id)
	}

	if err := s.Repo.DeleteByID(id); err != nil {
		return errors.New("session tidak bisa dihapus")
	}

	return nil
}

func sanitizeStockCheckSessionDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return ""
	}
	return value
}

func sanitizeStockCheckSessionStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "draft", "in_progress", "submitted", "reviewed", "closed", "cancelled":
		return value
	default:
		return ""
	}
}

func sanitizeStockCheckSessionInitiationType(value string) string {
	switch strings.TrimSpace(value) {
	case "scheduled", "checker_initiative":
		return value
	default:
		return ""
	}
}

func sanitizeStockCheckSessionCreateInput(input *models.StockCheckSessionCreateInput) {
	input.SessionDate = sanitizeStockCheckSessionDate(input.SessionDate)
	input.InitiationType = sanitizeStockCheckSessionInitiationType(input.InitiationType)
	input.Status = sanitizeStockCheckSessionStatus(input.Status)
	input.Notes = strings.TrimSpace(input.Notes)
}

func sanitizeStockCheckSessionUpdateInput(input *models.StockCheckSessionUpdateInput) {
	input.SessionDate = sanitizeStockCheckSessionDate(input.SessionDate)
	input.InitiationType = sanitizeStockCheckSessionInitiationType(input.InitiationType)
	input.Status = sanitizeStockCheckSessionStatus(input.Status)
	input.Notes = strings.TrimSpace(input.Notes)
}

func validateStockCheckSessionCreateInput(input models.StockCheckSessionCreateInput) error {
	if input.SessionDate == "" {
		return errors.New("tanggal session wajib diisi")
	}
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.SupplierID <= 0 {
		return errors.New("supplier wajib dipilih")
	}
	if input.InitiationType == "" {
		return errors.New("type wajib dipilih")
	}
	if input.Status == "" {
		return errors.New("status wajib dipilih")
	}
	return nil
}

func validateStockCheckSessionUpdateInput(input models.StockCheckSessionUpdateInput) error {
	if input.SessionDate == "" {
		return errors.New("tanggal session wajib diisi")
	}
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.SupplierID <= 0 {
		return errors.New("supplier wajib dipilih")
	}
	if input.InitiationType == "" {
		return errors.New("type wajib dipilih")
	}
	if input.Status == "" {
		return errors.New("status wajib dipilih")
	}
	return nil
}

func stockCheckSessionStageLabel(status string) string {
	switch status {
	case "draft":
		return "Preparing Review"
	case "in_progress", "submitted", "reviewed":
		return "Currently Reviewing"
	case "closed":
		return "Finalized Session"
	case "cancelled":
		return "Cancelled Session"
	default:
		return "Session Overview"
	}
}

func stockCheckSessionDetailStatusBadgeClass(status string) string {
	switch status {
	case "closed":
		return "session-badge-success"
	case "cancelled":
		return "session-badge-danger"
	case "reviewed", "submitted", "in_progress":
		return "session-badge-warm"
	default:
		return "session-badge-muted"
	}
}

func buildStockCheckSessionOverviewCards(detail models.StockCheckSessionDetail, alertItems int) []models.StockCheckSessionOverviewCard {
	demandDescription := "No review lines have been added to this session yet."
	if detail.ItemCount > 0 {
		demandDescription = fmt.Sprintf(
			"%d SKU membutuhkan evaluasi dengan usulan pembelian %s unit dan approval sementara %s unit.",
			detail.ItemCount,
			detail.TotalSuggestedQtyDisplay,
			detail.TotalApprovedQtyDisplay,
		)
	}

	riskDescription := "Belum ada supplier risk yang terdeteksi dari sesi ini."
	if detail.RejectedItems > 0 || alertItems > 0 || detail.DistinctSupplierCount > 1 {
		riskDescription = fmt.Sprintf(
			"%d item butuh perhatian, %d item ditolak, dan %d supplier terlibat dalam keputusan pembelian.",
			alertItems,
			detail.RejectedItems,
			detail.DistinctSupplierCount,
		)
	}

	return []models.StockCheckSessionOverviewCard{
		{
			Title:         "Projected Demand",
			Description:   demandDescription,
			Icon:          "bx bx-trending-up",
			IconWrapClass: "overview-tone-blue",
		},
		{
			Title:         "Supplier Risks",
			Description:   riskDescription,
			Icon:          "bx bx-error-alt",
			IconWrapClass: "overview-tone-amber",
		},
	}
}

func formatStockCheckSessionDecimal(value float64) string {
	return fmt.Sprintf("%0.2f", value)
}

func formatStockCheckSessionCurrency(value float64) string {
	return fmt.Sprintf("Rp %s", formatStockCheckSessionDecimal(value))
}
