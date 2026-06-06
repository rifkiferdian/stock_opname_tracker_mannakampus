package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type SupplierGroupService struct {
	Repo *repositories.SupplierGroupRepository
}

func (s *SupplierGroupService) GetSupplierGroups(filter models.SupplierGroupListFilter) ([]models.SupplierGroup, int, error) {
	filter.Search = strings.TrimSpace(filter.Search)

	switch filter.Status {
	case "active", "inactive", "":
	default:
		filter.Status = ""
	}

	switch filter.Sort {
	case "recent", "name", "code", "suppliers":
	default:
		filter.Sort = "recent"
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	totalItems, err := s.Repo.CountAll(filter)
	if err != nil {
		return nil, 0, err
	}

	if totalItems == 0 {
		return []models.SupplierGroup{}, 0, nil
	}

	totalPages := (totalItems + filter.Limit - 1) / filter.Limit
	if filter.Page > totalPages {
		filter.Page = totalPages
	}

	groups, err := s.Repo.GetAll(filter)
	if err != nil {
		return nil, 0, err
	}

	return groups, totalItems, nil
}

func (s *SupplierGroupService) GetSupplierGroupStats() (models.SupplierGroupStats, error) {
	return s.Repo.GetStats()
}

func (s *SupplierGroupService) CreateSupplierGroup(input models.SupplierGroupCreateInput) error {
	sanitizeSupplierGroupCreateInput(&input)

	if err := validateSupplierGroupCreateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByCode(input.GroupCode, input.StoreID, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode supplier group %s sudah digunakan", input.GroupCode)
	}

	return s.Repo.Create(input)
}

func (s *SupplierGroupService) UpdateSupplierGroup(input models.SupplierGroupUpdateInput) error {
	sanitizeSupplierGroupUpdateInput(&input)

	if input.ID <= 0 {
		return errors.New("supplier group id tidak valid")
	}
	if err := validateSupplierGroupUpdateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("supplier group id %d tidak ditemukan", input.ID)
	}

	codeExists, err := s.Repo.ExistsByCode(input.GroupCode, input.StoreID, input.ID)
	if err != nil {
		return err
	}
	if codeExists {
		return fmt.Errorf("kode supplier group %s sudah digunakan", input.GroupCode)
	}

	return s.Repo.Update(input)
}

func (s *SupplierGroupService) DeleteSupplierGroup(id int) error {
	if id <= 0 {
		return errors.New("supplier group id tidak valid")
	}
	return s.Repo.DeleteByID(id)
}

func sanitizeSupplierGroupCreateInput(input *models.SupplierGroupCreateInput) {
	input.GroupCode = strings.ToUpper(strings.TrimSpace(input.GroupCode))
	input.GroupName = strings.TrimSpace(input.GroupName)
	input.Description = strings.TrimSpace(input.Description)
}

func sanitizeSupplierGroupUpdateInput(input *models.SupplierGroupUpdateInput) {
	input.GroupCode = strings.ToUpper(strings.TrimSpace(input.GroupCode))
	input.GroupName = strings.TrimSpace(input.GroupName)
	input.Description = strings.TrimSpace(input.Description)
}

func validateSupplierGroupCreateInput(input models.SupplierGroupCreateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.GroupCode == "" {
		return errors.New("kode supplier group wajib diisi")
	}
	if input.GroupName == "" {
		return errors.New("nama supplier group wajib diisi")
	}
	return nil
}

func validateSupplierGroupUpdateInput(input models.SupplierGroupUpdateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.GroupCode == "" {
		return errors.New("kode supplier group wajib diisi")
	}
	if input.GroupName == "" {
		return errors.New("nama supplier group wajib diisi")
	}
	return nil
}
