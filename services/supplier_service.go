package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type SupplierService struct {
	Repo *repositories.SupplierRepository
}

func (s *SupplierService) GetSuppliers(filter models.SupplierListFilter) ([]models.Supplier, int, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Type = strings.TrimSpace(filter.Type)

	switch filter.Status {
	case "active", "inactive", "":
	default:
		filter.Status = ""
	}

	switch filter.Sort {
	case "recent", "name", "code", "products":
	default:
		filter.Sort = "recent"
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
		return []models.Supplier{}, 0, nil
	}

	totalPages := (totalItems + filter.Limit - 1) / filter.Limit
	if filter.Page > totalPages {
		filter.Page = totalPages
	}

	suppliers, err := s.Repo.GetAll(filter)
	if err != nil {
		return nil, 0, err
	}

	return suppliers, totalItems, nil
}

func (s *SupplierService) GetSupplierByID(id int) (models.Supplier, error) {
	if id <= 0 {
		return models.Supplier{}, errors.New("supplier id tidak valid")
	}
	return s.Repo.GetByID(id)
}

func (s *SupplierService) GetSuppliedProducts(supplierID int) ([]models.SupplierProduct, error) {
	if supplierID <= 0 {
		return nil, errors.New("supplier id tidak valid")
	}
	return s.Repo.GetSuppliedProducts(supplierID)
}

func (s *SupplierService) GetSupplierStats() (models.SupplierStats, error) {
	return s.Repo.GetStats()
}

func (s *SupplierService) GetSupplierGroups() ([]models.SupplierGroup, error) {
	return s.Repo.GetActiveGroups()
}

func (s *SupplierService) GetSupplierTypes() ([]string, error) {
	return s.Repo.GetTypes()
}

func (s *SupplierService) CreateSupplier(input models.SupplierCreateInput) error {
	sanitizeSupplierCreateInput(&input)

	if err := validateSupplierCreateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByCode(input.SupplierCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode supplier %s sudah digunakan", input.SupplierCode)
	}

	return s.Repo.Create(input)
}

func (s *SupplierService) UpdateSupplier(input models.SupplierUpdateInput) error {
	sanitizeSupplierUpdateInput(&input)

	if input.ID <= 0 {
		return errors.New("supplier id tidak valid")
	}
	if err := validateSupplierUpdateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("supplier id %d tidak ditemukan", input.ID)
	}

	codeExists, err := s.Repo.ExistsByCode(input.SupplierCode, input.ID)
	if err != nil {
		return err
	}
	if codeExists {
		return fmt.Errorf("kode supplier %s sudah digunakan", input.SupplierCode)
	}

	return s.Repo.Update(input)
}

func (s *SupplierService) DeleteSupplier(id int) error {
	if id <= 0 {
		return errors.New("supplier id tidak valid")
	}
	if err := s.Repo.DeleteByID(id); err != nil {
		return errors.New("supplier tidak bisa dihapus karena masih terhubung ke data lain")
	}
	return nil
}

func sanitizeSupplierCreateInput(input *models.SupplierCreateInput) {
	input.SupplierCode = strings.ToUpper(strings.TrimSpace(input.SupplierCode))
	input.SupplierName = strings.TrimSpace(input.SupplierName)
	input.SupplierType = strings.TrimSpace(input.SupplierType)
	input.Address = strings.TrimSpace(input.Address)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.TrimSpace(input.Email)
	input.PICName = strings.TrimSpace(input.PICName)
	if input.PaymentTermDays < 0 {
		input.PaymentTermDays = 0
	}
}

func sanitizeSupplierUpdateInput(input *models.SupplierUpdateInput) {
	input.SupplierCode = strings.ToUpper(strings.TrimSpace(input.SupplierCode))
	input.SupplierName = strings.TrimSpace(input.SupplierName)
	input.SupplierType = strings.TrimSpace(input.SupplierType)
	input.Address = strings.TrimSpace(input.Address)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.TrimSpace(input.Email)
	input.PICName = strings.TrimSpace(input.PICName)
	if input.PaymentTermDays < 0 {
		input.PaymentTermDays = 0
	}
}

func validateSupplierCreateInput(input models.SupplierCreateInput) error {
	if input.SupplierCode == "" {
		return errors.New("kode supplier wajib diisi")
	}
	if input.SupplierName == "" {
		return errors.New("nama supplier wajib diisi")
	}
	return nil
}

func validateSupplierUpdateInput(input models.SupplierUpdateInput) error {
	if input.SupplierCode == "" {
		return errors.New("kode supplier wajib diisi")
	}
	if input.SupplierName == "" {
		return errors.New("nama supplier wajib diisi")
	}
	return nil
}
