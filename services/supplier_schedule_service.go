package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
	"time"
)

type SupplierScheduleService struct {
	Repo *repositories.SupplierScheduleRepository
}

func (s *SupplierScheduleService) GetSchedules(filter models.SupplierScheduleListFilter, allowedStoreIDs []int) ([]models.SupplierSchedule, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	switch filter.Status {
	case "active", "inactive", "":
	default:
		filter.Status = ""
	}
	if filter.DayOfWeek < 1 || filter.DayOfWeek > 7 {
		filter.DayOfWeek = 0
	}

	return s.Repo.GetAll(filter, allowedStoreIDs)
}

func (s *SupplierScheduleService) GetScheduleByID(id int) (models.SupplierSchedule, error) {
	if id <= 0 {
		return models.SupplierSchedule{}, errors.New("jadwal tidak valid")
	}
	return s.Repo.GetByID(id)
}

func (s *SupplierScheduleService) GetStats(allowedStoreIDs []int) (models.SupplierScheduleStats, error) {
	return s.Repo.GetStats(allowedStoreIDs)
}

func (s *SupplierScheduleService) GetSupplierOptions(allowedStoreIDs []int) ([]models.Supplier, error) {
	return s.Repo.GetSupplierOptions(allowedStoreIDs)
}

func (s *SupplierScheduleService) CreateSchedule(input models.SupplierScheduleCreateInput) error {
	sanitizeScheduleCreateInput(&input)
	if err := validateScheduleCreateInput(input); err != nil {
		return err
	}
	normalizedTime, _ := normalizeScheduleTime(input.SOTime)
	input.SOTime = normalizedTime

	belongsToStore, err := s.Repo.SupplierBelongsToStore(input.StoreID, input.SupplierID)
	if err != nil {
		return err
	}
	if !belongsToStore {
		return errors.New("supplier tidak sesuai dengan store yang dipilih")
	}

	duplicate, err := s.Repo.ExistsDuplicate(input.StoreID, input.SupplierID, input.DayOfWeek, 0)
	if err != nil {
		return err
	}
	if duplicate {
		return errors.New("jadwal untuk supplier tersebut di hari yang sama sudah ada")
	}

	return s.Repo.Create(input)
}

func (s *SupplierScheduleService) UpdateSchedule(input models.SupplierScheduleUpdateInput) error {
	sanitizeScheduleUpdateInput(&input)

	if input.ID <= 0 {
		return errors.New("jadwal tidak valid")
	}
	if err := validateScheduleUpdateInput(input); err != nil {
		return err
	}
	normalizedTime, _ := normalizeScheduleTime(input.SOTime)
	input.SOTime = normalizedTime

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("jadwal id %d tidak ditemukan", input.ID)
	}

	belongsToStore, err := s.Repo.SupplierBelongsToStore(input.StoreID, input.SupplierID)
	if err != nil {
		return err
	}
	if !belongsToStore {
		return errors.New("supplier tidak sesuai dengan store yang dipilih")
	}

	duplicate, err := s.Repo.ExistsDuplicate(input.StoreID, input.SupplierID, input.DayOfWeek, input.ID)
	if err != nil {
		return err
	}
	if duplicate {
		return errors.New("jadwal untuk supplier tersebut di hari yang sama sudah ada")
	}

	return s.Repo.Update(input)
}

func (s *SupplierScheduleService) DeleteSchedule(id int) error {
	if id <= 0 {
		return errors.New("jadwal tidak valid")
	}
	return s.Repo.DeleteByID(id)
}

func sanitizeScheduleCreateInput(input *models.SupplierScheduleCreateInput) {
	input.Notes = strings.TrimSpace(input.Notes)
	input.SOTime = strings.TrimSpace(input.SOTime)
	if input.SequenceNo <= 0 {
		input.SequenceNo = 1
	}
}

func sanitizeScheduleUpdateInput(input *models.SupplierScheduleUpdateInput) {
	input.Notes = strings.TrimSpace(input.Notes)
	input.SOTime = strings.TrimSpace(input.SOTime)
	if input.SequenceNo <= 0 {
		input.SequenceNo = 1
	}
}

func validateScheduleCreateInput(input models.SupplierScheduleCreateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.SupplierID <= 0 {
		return errors.New("supplier wajib dipilih")
	}
	if input.DayOfWeek < 1 || input.DayOfWeek > 7 {
		return errors.New("hari jadwal tidak valid")
	}
	if input.SequenceNo <= 0 {
		return errors.New("urutan jadwal harus lebih dari 0")
	}
	if _, err := normalizeScheduleTime(input.SOTime); err != nil {
		return err
	}

	return nil
}

func validateScheduleUpdateInput(input models.SupplierScheduleUpdateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.SupplierID <= 0 {
		return errors.New("supplier wajib dipilih")
	}
	if input.DayOfWeek < 1 || input.DayOfWeek > 7 {
		return errors.New("hari jadwal tidak valid")
	}
	if input.SequenceNo <= 0 {
		return errors.New("urutan jadwal harus lebih dari 0")
	}
	if _, err := normalizeScheduleTime(input.SOTime); err != nil {
		return err
	}

	return nil
}

func normalizeScheduleTime(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	if len(value) == len("15:04") {
		value += ":00"
	}

	parsed, err := time.Parse("15:04:05", value)
	if err != nil {
		return "", errors.New("format jam jadwal tidak valid")
	}

	return parsed.Format("15:04:05"), nil
}
