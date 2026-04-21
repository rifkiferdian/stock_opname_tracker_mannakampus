package repositories

import (
	"database/sql"
	"gobase-app/models"
	"strings"
)

type UnitRepository struct {
	DB *sql.DB
}

func (r *UnitRepository) GetAll(filter models.UnitListFilter) ([]models.Unit, error) {
	query := `
		SELECT id, unit_code, unit_name, description, created_at, updated_at
		FROM units
	`

	args := make([]interface{}, 0, 2)
	conditions := make([]string, 0, 1)

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		conditions = append(conditions, "(LOWER(unit_code) LIKE ? OR LOWER(unit_name) LIKE ?)")
		args = append(args, keyword, keyword)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch filter.Sort {
	case "code":
		query += " ORDER BY unit_code ASC, id ASC"
	case "name":
		query += " ORDER BY unit_name ASC, id ASC"
	default:
		query += " ORDER BY updated_at DESC, id DESC"
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []models.Unit
	for rows.Next() {
		var (
			unit        models.Unit
			description sql.NullString
			createdAt   sql.NullTime
			updatedAt   sql.NullTime
		)

		if err := rows.Scan(
			&unit.ID,
			&unit.UnitCode,
			&unit.UnitName,
			&description,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		if description.Valid {
			unit.Description = description.String
		}
		if createdAt.Valid {
			unit.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			unit.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
		} else {
			unit.CreatedAt = "-"
			unit.CreatedAtDisplay = "-"
		}
		if updatedAt.Valid {
			unit.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
			unit.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
		} else {
			unit.UpdatedAt = "-"
			unit.UpdatedAtDisplay = "-"
		}

		units = append(units, unit)
	}

	return units, rows.Err()
}

func (r *UnitRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM units WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *UnitRepository) ExistsByCode(unitCode string, ignoreID int) (bool, error) {
	var (
		count int
		err   error
	)

	if ignoreID > 0 {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM units WHERE unit_code = ? AND id <> ?`, unitCode, ignoreID).Scan(&count)
	} else {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM units WHERE unit_code = ?`, unitCode).Scan(&count)
	}

	return count > 0, err
}

func (r *UnitRepository) Create(input models.UnitCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO units (unit_code, unit_name, description)
		VALUES (?, ?, ?)
	`, input.UnitCode, input.UnitName, input.Description)

	return err
}

func (r *UnitRepository) Update(input models.UnitUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE units
		SET unit_code = ?, unit_name = ?, description = ?
		WHERE id = ?
	`, input.UnitCode, input.UnitName, input.Description, input.ID)

	return err
}

func (r *UnitRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM units WHERE id = ?`, id)
	return err
}
