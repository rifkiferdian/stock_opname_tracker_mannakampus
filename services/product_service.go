package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type ProductService struct {
	Repo *repositories.ProductRepository
}

func (s *ProductService) GetProducts(filter models.ProductListFilter) ([]models.Product, int, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Brand = strings.TrimSpace(filter.Brand)

	switch filter.Status {
	case "active", "inactive", "":
	default:
		filter.Status = ""
	}

	switch filter.Sort {
	case "recent", "name", "code", "brand":
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
		return []models.Product{}, 0, nil
	}

	totalPages := (totalItems + filter.Limit - 1) / filter.Limit
	if filter.Page > totalPages {
		filter.Page = totalPages
	}

	products, err := s.Repo.GetAll(filter)
	if err != nil {
		return nil, 0, err
	}

	return products, totalItems, nil
}

func (s *ProductService) GetProductByID(id int) (models.ProductDetail, error) {
	if id <= 0 {
		return models.ProductDetail{}, errors.New("product id tidak valid")
	}
	return s.Repo.GetByID(id)
}

func (s *ProductService) GetProductSupplierNetwork(productID int) ([]models.ProductSupplierNetwork, error) {
	if productID <= 0 {
		return nil, errors.New("product id tidak valid")
	}
	return s.Repo.GetSupplierNetwork(productID)
}

func (s *ProductService) GetProductStockHistory(productID int) ([]models.ProductStockHistory, error) {
	if productID <= 0 {
		return nil, errors.New("product id tidak valid")
	}
	return s.Repo.GetStockHistory(productID)
}

func (s *ProductService) GetProductStats() (models.ProductStats, error) {
	return s.Repo.GetStats()
}

func (s *ProductService) GetCategories() ([]models.ProductCategory, error) {
	return s.Repo.GetCategories()
}

func (s *ProductService) GetUnits() ([]models.Unit, error) {
	return s.Repo.GetUnits()
}

func (s *ProductService) GetSuppliers() ([]models.ProductSupplierOption, error) {
	return s.Repo.GetSuppliers()
}

func (s *ProductService) GetBrands() ([]string, error) {
	return s.Repo.GetBrands()
}

func (s *ProductService) CreateProduct(input models.ProductCreateInput) error {
	sanitizeProductCreateInput(&input)

	if err := validateProductCreateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByCode(input.ProductCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode produk %s sudah digunakan", input.ProductCode)
	}

	return s.Repo.Create(input)
}

func (s *ProductService) UpdateProduct(input models.ProductUpdateInput) error {
	sanitizeProductUpdateInput(&input)

	if input.ID <= 0 {
		return errors.New("product id tidak valid")
	}
	if err := validateProductUpdateInput(input); err != nil {
		return err
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("product id %d tidak ditemukan", input.ID)
	}

	codeExists, err := s.Repo.ExistsByCode(input.ProductCode, input.ID)
	if err != nil {
		return err
	}
	if codeExists {
		return fmt.Errorf("kode produk %s sudah digunakan", input.ProductCode)
	}

	return s.Repo.Update(input)
}

func (s *ProductService) DeleteProduct(id int) error {
	if id <= 0 {
		return errors.New("product id tidak valid")
	}
	if err := s.Repo.DeleteByID(id); err != nil {
		return errors.New("produk tidak bisa dihapus karena masih terhubung ke data lain")
	}
	return nil
}

func sanitizeProductCreateInput(input *models.ProductCreateInput) {
	input.ProductCode = strings.ToUpper(strings.TrimSpace(input.ProductCode))
	input.Barcode = strings.TrimSpace(input.Barcode)
	input.ProductName = strings.TrimSpace(input.ProductName)
	input.Brand = strings.TrimSpace(input.Brand)
	if input.MinStock < 0 {
		input.MinStock = 0
	}
	if input.MaxStock < 0 {
		input.MaxStock = 0
	}
	if input.ReorderPoint < 0 {
		input.ReorderPoint = 0
	}
	if input.DefaultLeadTimeDays < 0 {
		input.DefaultLeadTimeDays = 0
	}
	if input.PackSize <= 0 {
		input.PackSize = 1
	}
	if input.PcsPerBox < 0 {
		input.PcsPerBox = 0
	}
	if input.BoxPerCarton < 0 {
		input.BoxPerCarton = 0
	}
	if input.PcsPerCarton < 0 {
		input.PcsPerCarton = 0
	}
	if input.LastPrice < 0 {
		input.LastPrice = 0
	}
}

func sanitizeProductUpdateInput(input *models.ProductUpdateInput) {
	input.ProductCode = strings.ToUpper(strings.TrimSpace(input.ProductCode))
	input.Barcode = strings.TrimSpace(input.Barcode)
	input.ProductName = strings.TrimSpace(input.ProductName)
	input.Brand = strings.TrimSpace(input.Brand)
	if input.MinStock < 0 {
		input.MinStock = 0
	}
	if input.MaxStock < 0 {
		input.MaxStock = 0
	}
	if input.ReorderPoint < 0 {
		input.ReorderPoint = 0
	}
	if input.DefaultLeadTimeDays < 0 {
		input.DefaultLeadTimeDays = 0
	}
	if input.PackSize <= 0 {
		input.PackSize = 1
	}
	if input.PcsPerBox < 0 {
		input.PcsPerBox = 0
	}
	if input.BoxPerCarton < 0 {
		input.BoxPerCarton = 0
	}
	if input.PcsPerCarton < 0 {
		input.PcsPerCarton = 0
	}
	if input.LastPrice < 0 {
		input.LastPrice = 0
	}
}

func validateProductCreateInput(input models.ProductCreateInput) error {
	if input.ProductCode == "" {
		return errors.New("kode produk wajib diisi")
	}
	if input.ProductName == "" {
		return errors.New("nama produk wajib diisi")
	}
	return nil
}

func validateProductUpdateInput(input models.ProductUpdateInput) error {
	if input.ProductCode == "" {
		return errors.New("kode produk wajib diisi")
	}
	if input.ProductName == "" {
		return errors.New("nama produk wajib diisi")
	}
	return nil
}
