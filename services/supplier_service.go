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
	filter.LastSODate = strings.TrimSpace(filter.LastSODate)

	switch filter.Status {
	case "active", "inactive", "":
	default:
		filter.Status = ""
	}

	switch filter.Sort {
	case "recent", "name", "name_asc", "name_desc", "code", "products":
	default:
		filter.Sort = "recent"
	}
	if filter.DayOfWeek < 1 || filter.DayOfWeek > 7 {
		filter.DayOfWeek = 0
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

func (s *SupplierService) GetSupplierProductGroups(supplierID int) ([]models.SupplierProductGroupItem, error) {
	if supplierID <= 0 {
		return nil, errors.New("supplier id tidak valid")
	}
	return s.Repo.GetSupplierProductGroups(supplierID)
}

func (s *SupplierService) GetAvailableProductOptions(supplierID int) ([]models.SupplierProductOption, error) {
	if supplierID <= 0 {
		return nil, errors.New("supplier id tidak valid")
	}
	return s.Repo.GetAvailableProductOptions(supplierID)
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

	exists, err := s.Repo.ExistsByCode(input.SupplierCode, input.StoreID, 0)
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

	codeExists, err := s.Repo.ExistsByCode(input.SupplierCode, input.StoreID, input.ID)
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

func (s *SupplierService) CreateSupplierProduct(input models.SupplierProductCreateInput) error {
	sanitizeSupplierProductCreateInput(&input)

	if input.SupplierID <= 0 {
		return errors.New("supplier id tidak valid")
	}
	if input.ProductID <= 0 {
		return errors.New("item wajib dipilih")
	}

	supplierExists, err := s.Repo.ExistsByID(input.SupplierID)
	if err != nil {
		return err
	}
	if !supplierExists {
		return fmt.Errorf("supplier id %d tidak ditemukan", input.SupplierID)
	}

	productExists, err := s.Repo.ProductExistsByID(input.ProductID)
	if err != nil {
		return err
	}
	if !productExists {
		return fmt.Errorf("produk id %d tidak ditemukan", input.ProductID)
	}

	if input.SupplierProductGroupID > 0 {
		groupBelongs, err := s.Repo.SupplierProductGroupBelongsToSupplier(input.SupplierProductGroupID, input.SupplierID)
		if err != nil {
			return err
		}
		if !groupBelongs {
			return errors.New("group item tidak ditemukan untuk supplier ini")
		}
	}

	return s.Repo.UpsertProductSupply(input)
}

func (s *SupplierService) UpdateSupplierProduct(input models.SupplierProductCreateInput) error {
	sanitizeSupplierProductCreateInput(&input)

	if input.SupplierID <= 0 {
		return errors.New("supplier id tidak valid")
	}
	if input.ProductID <= 0 {
		return errors.New("item wajib dipilih")
	}

	exists, err := s.Repo.ProductSupplyExists(input.SupplierID, input.ProductID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("tautan item supplier tidak ditemukan")
	}

	if !input.KeepExistingGroup && input.SupplierProductGroupID > 0 {
		groupBelongs, err := s.Repo.SupplierProductGroupBelongsToSupplier(input.SupplierProductGroupID, input.SupplierID)
		if err != nil {
			return err
		}
		if !groupBelongs {
			return errors.New("group item tidak ditemukan untuk supplier ini")
		}
	}

	return s.Repo.UpsertProductSupply(input)
}

func (s *SupplierService) DeleteSupplierProduct(input models.SupplierProductDeleteInput) error {
	if input.SupplierID <= 0 {
		return errors.New("supplier id tidak valid")
	}
	if input.ProductID <= 0 {
		return errors.New("item supplier tidak valid")
	}

	exists, err := s.Repo.ProductSupplyExists(input.SupplierID, input.ProductID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("tautan item supplier tidak ditemukan")
	}

	return s.Repo.DeleteProductSupply(input)
}

func (s *SupplierService) CreateSupplierProductGroup(input models.SupplierProductGroupCreateInput) error {
	sanitizeSupplierProductGroupCreateInput(&input)

	if err := validateSupplierProductGroupCreateInput(input); err != nil {
		return err
	}

	supplierExists, err := s.Repo.ExistsByID(input.SupplierID)
	if err != nil {
		return err
	}
	if !supplierExists {
		return fmt.Errorf("supplier id %d tidak ditemukan", input.SupplierID)
	}

	exists, err := s.Repo.SupplierProductGroupExistsByName(input.SupplierID, input.GroupName, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("nama group item %s sudah digunakan untuk supplier ini", input.GroupName)
	}

	return s.Repo.CreateSupplierProductGroup(input)
}

func (s *SupplierService) UpdateSupplierProductGroup(input models.SupplierProductGroupUpdateInput) error {
	sanitizeSupplierProductGroupUpdateInput(&input)

	if input.ID <= 0 {
		return errors.New("group item supplier tidak valid")
	}
	if err := validateSupplierProductGroupUpdateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.SupplierProductGroupExists(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("group item supplier tidak ditemukan")
	}

	groupBelongs, err := s.Repo.SupplierProductGroupBelongsToSupplier(input.ID, input.SupplierID)
	if err != nil {
		return err
	}
	if !groupBelongs {
		return errors.New("group item tidak ditemukan untuk supplier ini")
	}

	nameExists, err := s.Repo.SupplierProductGroupExistsByName(input.SupplierID, input.GroupName, input.ID)
	if err != nil {
		return err
	}
	if nameExists {
		return fmt.Errorf("nama group item %s sudah digunakan untuk supplier ini", input.GroupName)
	}

	return s.Repo.UpdateSupplierProductGroup(input)
}

func (s *SupplierService) DeleteSupplierProductGroup(id int, supplierID int) error {
	if id <= 0 {
		return errors.New("group item supplier tidak valid")
	}
	if supplierID <= 0 {
		return errors.New("supplier id tidak valid")
	}

	groupBelongs, err := s.Repo.SupplierProductGroupBelongsToSupplier(id, supplierID)
	if err != nil {
		return err
	}
	if !groupBelongs {
		return errors.New("group item tidak ditemukan untuk supplier ini")
	}

	return s.Repo.DeleteSupplierProductGroup(id, supplierID)
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

func sanitizeSupplierProductCreateInput(input *models.SupplierProductCreateInput) {
	if input.LastPrice < 0 {
		input.LastPrice = 0
	}
	if input.MOQ < 0 {
		input.MOQ = 0
	}
	if input.PackSize <= 0 {
		input.PackSize = 1
	}
	if input.LeadTimeDays < 0 {
		input.LeadTimeDays = 0
	}
}

func sanitizeSupplierProductGroupCreateInput(input *models.SupplierProductGroupCreateInput) {
	input.GroupName = strings.TrimSpace(input.GroupName)
	input.Description = strings.TrimSpace(input.Description)
	if input.SortOrder < 0 {
		input.SortOrder = 0
	}
}

func sanitizeSupplierProductGroupUpdateInput(input *models.SupplierProductGroupUpdateInput) {
	input.GroupName = strings.TrimSpace(input.GroupName)
	input.Description = strings.TrimSpace(input.Description)
	if input.SortOrder < 0 {
		input.SortOrder = 0
	}
}

func validateSupplierCreateInput(input models.SupplierCreateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.SupplierCode == "" {
		return errors.New("kode supplier wajib diisi")
	}
	if input.SupplierName == "" {
		return errors.New("nama supplier wajib diisi")
	}
	return nil
}

func validateSupplierUpdateInput(input models.SupplierUpdateInput) error {
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if input.SupplierCode == "" {
		return errors.New("kode supplier wajib diisi")
	}
	if input.SupplierName == "" {
		return errors.New("nama supplier wajib diisi")
	}
	return nil
}

func validateSupplierProductGroupCreateInput(input models.SupplierProductGroupCreateInput) error {
	if input.SupplierID <= 0 {
		return errors.New("supplier wajib dipilih")
	}
	if input.GroupName == "" {
		return errors.New("nama group item wajib diisi")
	}
	return nil
}

func validateSupplierProductGroupUpdateInput(input models.SupplierProductGroupUpdateInput) error {
	if input.SupplierID <= 0 {
		return errors.New("supplier wajib dipilih")
	}
	if input.GroupName == "" {
		return errors.New("nama group item wajib diisi")
	}
	return nil
}
