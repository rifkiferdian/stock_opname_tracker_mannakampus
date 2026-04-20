package services

import (
	"errors"
	"fmt"
	"gobase-app/models"
	"gobase-app/repositories"
	"strings"
)

type UnitService struct {
	Repo *repositories.UnitRepository
}

func (s *UnitService) GetUnits() ([]models.Unit, error) {
	return s.Repo.GetAll()
}

func (s *UnitService) CreateUnit(input models.UnitCreateInput) error {
	unitCode := strings.ToUpper(strings.TrimSpace(input.UnitCode))
	unitName := strings.TrimSpace(input.UnitName)
	description := strings.TrimSpace(input.Description)

	if unitCode == "" {
		return errors.New("kode unit wajib diisi")
	}
	if unitName == "" {
		return errors.New("nama unit wajib diisi")
	}

	exists, err := s.Repo.ExistsByCode(unitCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode unit %s sudah digunakan", unitCode)
	}

	return s.Repo.Create(models.UnitCreateInput{
		UnitCode:    unitCode,
		UnitName:    unitName,
		Description: description,
	})
}

func (s *UnitService) UpdateUnit(input models.UnitUpdateInput) error {
	unitCode := strings.ToUpper(strings.TrimSpace(input.UnitCode))
	unitName := strings.TrimSpace(input.UnitName)
	description := strings.TrimSpace(input.Description)

	if input.ID <= 0 {
		return errors.New("unit id tidak valid")
	}
	if unitCode == "" {
		return errors.New("kode unit wajib diisi")
	}
	if unitName == "" {
		return errors.New("nama unit wajib diisi")
	}

	exists, err := s.Repo.ExistsByID(input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("unit id %d tidak ditemukan", input.ID)
	}

	codeExists, err := s.Repo.ExistsByCode(unitCode, input.ID)
	if err != nil {
		return err
	}
	if codeExists {
		return fmt.Errorf("kode unit %s sudah digunakan", unitCode)
	}

	return s.Repo.Update(models.UnitUpdateInput{
		ID:          input.ID,
		UnitCode:    unitCode,
		UnitName:    unitName,
		Description: description,
	})
}

func (s *UnitService) DeleteUnit(id int) error {
	if id <= 0 {
		return errors.New("unit id tidak valid")
	}
	return s.Repo.DeleteByID(id)
}
