package services

import (
	"database/sql"
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

func (s *StockCheckSessionService) GetStoreOptionsByUserID(userID int) ([]models.Store, error) {
	if userID <= 0 {
		return []models.Store{}, nil
	}
	return s.Repo.GetStoreOptionsByUserID(userID)
}

func (s *StockCheckSessionService) GetSupplierOptions() ([]models.Supplier, error) {
	return s.Repo.GetSupplierOptions()
}

func (s *StockCheckSessionService) GetSessionDetailPage(id int, page int, limit int) (models.StockCheckSessionDetailPage, error) {
	if id <= 0 {
		return models.StockCheckSessionDetailPage{}, errors.New("session id tidak valid")
	}
	if page <= 0 {
		page = 1
	}
	disablePagination := limit <= 0

	session, err := s.Repo.GetByID(id)
	if err != nil {
		return models.StockCheckSessionDetailPage{}, err
	}

	items, err := s.Repo.GetReviewItems(id)
	if err != nil {
		return models.StockCheckSessionDetailPage{}, err
	}

	totalItems := len(items)
	totalPages := 0
	if disablePagination {
		page = 1
		if totalItems > 0 {
			totalPages = 1
		}
	} else if totalItems > 0 {
		totalPages = (totalItems + limit - 1) / limit
		if page > totalPages {
			page = totalPages
		}
	}

	pagedItems := items
	if !disablePagination && totalItems > 0 {
		startIndex := (page - 1) * limit
		endIndex := startIndex + limit
		if endIndex > totalItems {
			endIndex = totalItems
		}
		pagedItems = items[startIndex:endIndex]
	}

	detail := models.StockCheckSessionDetail{
		StockCheckSession: session,
		StageLabel:        stockCheckSessionStageLabel(session.Status),
		StatusBadgeClass:  stockCheckSessionDetailStatusBadgeClass(session.Status),
		ItemCount:         totalItems,
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
	detail.TotalApprovedQtyDisplay = formatStockCheckSessionWholeNumber(detail.TotalApprovedQty)
	detail.SuggestedPurchaseValueDisplay = formatStockCheckSessionCurrency(detail.SuggestedPurchaseValue)
	detail.FinalApprovedValueDisplay = formatStockCheckSessionCurrency(detail.FinalApprovedValue)
	if detail.SuggestedPurchaseValue > 0 {
		detail.ApprovalYieldPercent = (detail.FinalApprovedValue / detail.SuggestedPurchaseValue) * 100
	}
	detail.ApprovalYieldDisplay = fmt.Sprintf("%.1f%%", detail.ApprovalYieldPercent)

	return models.StockCheckSessionDetailPage{
		Session:       detail,
		Items:         pagedItems,
		OverviewCards: buildStockCheckSessionOverviewCards(detail, alertItems),
		Pagination: models.Pagination{
			CurrentPage: page,
			PageSize:    limit,
			TotalItems:  totalItems,
			TotalPages:  totalPages,
		},
	}, nil
}

func (s *StockCheckSessionService) GetCheckerInputPage(sessionID int, userID int) (models.StockCheckSessionCheckerInputPage, error) {
	if sessionID <= 0 {
		return models.StockCheckSessionCheckerInputPage{}, errors.New("session id tidak valid")
	}
	if userID <= 0 {
		return models.StockCheckSessionCheckerInputPage{}, errors.New("user login tidak valid")
	}

	session, err := s.Repo.GetByID(sessionID)
	if err != nil {
		return models.StockCheckSessionCheckerInputPage{}, err
	}

	hasAccess, err := s.userCanAccessCheckerStore(userID, session.StoreID)
	if err != nil {
		return models.StockCheckSessionCheckerInputPage{}, err
	}
	if !hasAccess {
		return models.StockCheckSessionCheckerInputPage{}, errors.New("session tidak tersedia untuk user login")
	}

	items, err := s.Repo.GetCheckerInputItems(sessionID, session.StoreID, session.SupplierID)
	if err != nil {
		return models.StockCheckSessionCheckerInputPage{}, err
	}

	return models.StockCheckSessionCheckerInputPage{
		Session: session,
		Items:   items,
	}, nil
}

func (s *StockCheckSessionService) GetCheckerScanPage(sessionID int, userID int, barcode string) (models.StockCheckSessionCheckerScanPage, error) {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return models.StockCheckSessionCheckerScanPage{}, errors.New("barcode wajib diisi")
	}

	pageData, err := s.GetCheckerInputPage(sessionID, userID)
	if err != nil {
		return models.StockCheckSessionCheckerScanPage{}, err
	}

	normalizedBarcode := strings.ToLower(barcode)
	for _, item := range pageData.Items {
		if strings.ToLower(strings.TrimSpace(item.Barcode)) == normalizedBarcode ||
			strings.ToLower(strings.TrimSpace(item.BarcodeBox)) == normalizedBarcode ||
			strings.ToLower(strings.TrimSpace(item.BarcodeCarton)) == normalizedBarcode {
			history, err := s.Repo.GetCheckerItemRecentSOHistory(sessionID, item.ProductID, pageData.Session.StoreID, pageData.Session.SupplierID, 4)
			if err != nil {
				return models.StockCheckSessionCheckerScanPage{}, err
			}
			item.RecentSOHistory = history
			return models.StockCheckSessionCheckerScanPage{
				Session: pageData.Session,
				Item:    item,
				Items:   pageData.Items,
			}, nil
		}
	}

	return models.StockCheckSessionCheckerScanPage{}, sql.ErrNoRows
}

func (s *StockCheckSessionService) CreateSession(input models.StockCheckSessionCreateInput) (int, error) {
	sanitizeStockCheckSessionCreateInput(&input)

	if input.CreatedBy <= 0 {
		return 0, errors.New("user login tidak valid")
	}
	if err := validateStockCheckSessionCreateInput(input); err != nil {
		return 0, err
	}

	if input.CreatedBy > 0 {
		hasAccess, err := s.userCanAccessCheckerStore(input.CreatedBy, input.StoreID)
		if err != nil {
			return 0, err
		}
		if !hasAccess {
			return 0, errors.New("store tidak tersedia untuk user login")
		}
	}

	storeCode, err := s.Repo.GetStoreCodeByID(input.StoreID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(storeCode) == "" {
		return 0, errors.New("store code tidak ditemukan")
	}

	prefix := fmt.Sprintf("SCS-%s-%s-", storeCode, strings.ReplaceAll(input.SessionDate, "-", ""))
	nextSequence, err := s.Repo.GetNextSessionSequence(prefix)
	if err != nil {
		return 0, err
	}

	input.SessionNumber = fmt.Sprintf("%s%03d", prefix, nextSequence)
	sessionID, err := s.Repo.Create(input)
	if err != nil {
		return 0, err
	}

	if err := s.Repo.SeedItemsFromSupplier(sessionID, input.SupplierID, input.StoreID, input.CreatedBy); err != nil {
		return 0, err
	}

	return sessionID, nil
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

func (s *StockCheckSessionService) UpdateReviewItem(input models.StockCheckSessionReviewItemUpdateInput) error {
	input.BuyerNotes = strings.TrimSpace(input.BuyerNotes)

	if input.SessionID <= 0 {
		return errors.New("session id tidak valid")
	}
	if input.ItemID <= 0 {
		return errors.New("item id tidak valid")
	}
	if input.ApprovedBuyCarton < 0 || input.ApprovedBuyBox < 0 || input.ApprovedBuyPcs < 0 {
		return errors.New("final approve tidak boleh kurang dari 0")
	}
	if len(input.BuyerNotes) > 255 {
		return errors.New("notes maksimal 255 karakter")
	}
	if input.ReviewedBy <= 0 {
		return errors.New("user login tidak valid")
	}
	if input.UpdatedBy <= 0 {
		input.UpdatedBy = input.ReviewedBy
	}

	sessionExists, err := s.Repo.ExistsByID(input.SessionID)
	if err != nil {
		return err
	}
	if !sessionExists {
		return fmt.Errorf("session id %d tidak ditemukan", input.SessionID)
	}

	itemExists, err := s.Repo.ExistsReviewItem(input.SessionID, input.ItemID)
	if err != nil {
		return err
	}
	if !itemExists {
		return fmt.Errorf("item review id %d tidak ditemukan", input.ItemID)
	}

	input.Status = "approved"
	if input.ApprovedBuyCarton == 0 && input.ApprovedBuyBox == 0 && input.ApprovedBuyPcs == 0 {
		input.Status = "rejected"
	}

	return s.Repo.UpdateReviewItem(input)
}

func (s *StockCheckSessionService) ApplyAllSubmittedLatestBySupplier(supplierID int, reviewedBy int) (int, error) {
	if supplierID <= 0 {
		return 0, errors.New("supplier id tidak valid")
	}
	if reviewedBy <= 0 {
		return 0, errors.New("user login tidak valid")
	}

	items, err := s.Repo.GetLatestSubmittedItemsBySupplier(supplierID)
	if err != nil {
		return 0, err
	}

	appliedCount := 0
	for _, item := range items {
		if err := s.UpdateReviewItem(models.StockCheckSessionReviewItemUpdateInput{
			SessionID:         item.SessionID,
			ItemID:            item.ItemID,
			ApprovedBuyCarton: item.SuggestBuyCarton,
			ApprovedBuyBox:    item.SuggestBuyBox,
			ApprovedBuyPcs:    item.SuggestBuyPcs,
			BuyerNotes:        item.BuyerNotes,
			ReviewedBy:        reviewedBy,
			UpdatedBy:         reviewedBy,
		}); err != nil {
			return appliedCount, err
		}

		appliedCount++
	}

	return appliedCount, nil
}

func (s *StockCheckSessionService) ApplyAllSubmittedBySession(sessionID int, reviewedBy int) (int, error) {
	if sessionID <= 0 {
		return 0, errors.New("session id tidak valid")
	}
	if reviewedBy <= 0 {
		return 0, errors.New("user login tidak valid")
	}

	items, err := s.Repo.GetSubmittedItemsBySession(sessionID)
	if err != nil {
		return 0, err
	}

	appliedCount := 0
	for _, item := range items {
		if err := s.UpdateReviewItem(models.StockCheckSessionReviewItemUpdateInput{
			SessionID:         sessionID,
			ItemID:            item.ItemID,
			ApprovedBuyCarton: item.SuggestBuyCarton,
			ApprovedBuyBox:    item.SuggestBuyBox,
			ApprovedBuyPcs:    item.SuggestBuyPcs,
			BuyerNotes:        item.BuyerNotes,
			ReviewedBy:        reviewedBy,
			UpdatedBy:         reviewedBy,
		}); err != nil {
			return appliedCount, err
		}

		appliedCount++
	}

	return appliedCount, nil
}

func (s *StockCheckSessionService) RecordCheckerScan(input models.StockCheckSessionCheckerScanInput) (int, error) {
	input.Location = sanitizeStockCheckCheckerLocation(input.Location)
	input.Barcode = strings.TrimSpace(input.Barcode)

	if input.SessionID <= 0 {
		return 0, errors.New("session id tidak valid")
	}
	if input.UpdatedBy <= 0 {
		return 0, errors.New("user login tidak valid")
	}
	if input.Location == "" {
		return 0, errors.New("lokasi input wajib dipilih")
	}
	if input.Barcode == "" {
		return 0, errors.New("barcode wajib diisi")
	}
	if input.QtyCarton < 0 || input.QtyBox < 0 || input.QtyPcs < 0 {
		return 0, errors.New("qty tidak boleh kurang dari 0")
	}

	session, err := s.Repo.GetByID(input.SessionID)
	if err != nil {
		return 0, err
	}

	hasAccess, err := s.userCanAccessCheckerStore(input.UpdatedBy, session.StoreID)
	if err != nil {
		return 0, err
	}
	if !hasAccess {
		return 0, errors.New("session tidak tersedia untuk user login")
	}

	itemID, err := s.Repo.UpdateCheckerItemQtyByBarcode(
		input.SessionID,
		session.StoreID,
		session.SupplierID,
		input.Location,
		input.Barcode,
		input.QtyCarton,
		input.QtyBox,
		input.QtyPcs,
		input.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("Barcode Item ini tidak ada di supplier ini.")
		}
		return 0, err
	}

	return itemID, nil
}

func (s *StockCheckSessionService) UpdateCheckerSuggest(input models.StockCheckSessionCheckerSuggestInput) error {
	if input.SessionID <= 0 {
		return errors.New("session id tidak valid")
	}
	if input.ItemID <= 0 {
		return errors.New("item id tidak valid")
	}
	if input.SuggestCarton < 0 {
		return errors.New("suggest carton tidak boleh kurang dari 0")
	}
	if input.SuggestBox < 0 {
		return errors.New("suggest box tidak boleh kurang dari 0")
	}
	if input.SuggestPcs < 0 {
		return errors.New("suggest pcs tidak boleh kurang dari 0")
	}
	if input.UpdatedBy <= 0 {
		return errors.New("user login tidak valid")
	}

	session, err := s.Repo.GetByID(input.SessionID)
	if err != nil {
		return err
	}

	hasAccess, err := s.userCanAccessCheckerStore(input.UpdatedBy, session.StoreID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("session tidak tersedia untuk user login")
	}

	itemExists, err := s.Repo.ExistsReviewItem(input.SessionID, input.ItemID)
	if err != nil {
		return err
	}
	if !itemExists {
		return errors.New("item tidak ditemukan pada session ini")
	}

	return s.Repo.UpdateCheckerItemSuggest(
		input.SessionID,
		input.ItemID,
		input.SuggestCarton,
		input.SuggestBox,
		input.SuggestPcs,
		input.UpdatedBy,
	)
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

func (s *StockCheckSessionService) UpdateSessionStatusForPORecap(sessionID int, status string) error {
	if sessionID <= 0 {
		return errors.New("session id tidak valid")
	}

	status = sanitizeStockCheckSessionStatus(status)
	if status != "closed" && status != "po" && status != "reviewed" {
		return errors.New("status harus reviewed, closed, atau po")
	}

	exists, err := s.Repo.ExistsByID(sessionID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("session id %d tidak ditemukan", sessionID)
	}

	return s.Repo.UpdateStatus(sessionID, status)
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
	case "draft", "in_progress", "submitted", "reviewed", "closed", "po", "cancelled":
		return value
	default:
		return ""
	}
}

func sanitizeStockCheckCheckerLocation(value string) string {
	switch strings.TrimSpace(value) {
	case "store", "warehouse":
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
	if input.SessionDate == "" {
		input.SessionDate = time.Now().Format("2006-01-02")
	}
	if input.InitiationType == "" {
		input.InitiationType = "checker_initiative"
	}
	if input.Status == "" {
		input.Status = "in_progress"
	}
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
	case "po":
		return "PO Completed"
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
	case "po":
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

func formatStockCheckSessionWholeNumber(value float64) string {
	return fmt.Sprintf("%.0f", value)
}

func formatStockCheckSessionCurrency(value float64) string {
	return fmt.Sprintf("Rp %s", formatStockCheckSessionDecimal(value))
}

func (s *StockCheckSessionService) userCanAccessCheckerStore(userID int, storeID int) (bool, error) {
	hasStoreAccess, err := s.Repo.UserHasStoreAccess(userID, storeID)
	if err != nil {
		return false, err
	}
	if hasStoreAccess {
		return true, nil
	}

	isChecker, err := s.Repo.UserHasRole(userID, "checker")
	if err != nil {
		return false, err
	}
	if isChecker {
		return true, nil
	}

	return false, nil
}
