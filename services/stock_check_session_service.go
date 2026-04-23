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
