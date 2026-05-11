package repositories

import (
	"database/sql"
	"gobase-app/models"
	"strings"
)

type SupplierGroupRepository struct {
	DB *sql.DB
}

func (r *SupplierGroupRepository) GetAll(filter models.SupplierGroupListFilter) ([]models.SupplierGroup, error) {
	query, args := buildSupplierGroupListQuery(filter, false)
	query += buildSupplierGroupOrderClause(filter)
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.SupplierGroup
	for rows.Next() {
		group, err := scanSupplierGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

func (r *SupplierGroupRepository) CountAll(filter models.SupplierGroupListFilter) (int, error) {
	query, args := buildSupplierGroupListQuery(filter, true)

	var total int
	err := r.DB.QueryRow(query, args...).Scan(&total)
	return total, err
}

func buildSupplierGroupListQuery(filter models.SupplierGroupListFilter, countOnly bool) (string, []interface{}) {
	query := `
	`
	if countOnly {
		query = `
		SELECT COUNT(*)
		FROM supplier_groups sg
	`
	} else {
		query = `
		SELECT
			sg.id,
			COALESCE(sg.store_id, 0) AS store_id,
			COALESCE(st.store_name, '') AS store_name,
			sg.group_code,
			sg.group_name,
			COALESCE(sg.description, '') AS description,
			sg.is_active,
			sg.created_at,
			sg.updated_at,
			COUNT(DISTINCT s.id) AS supplier_count
		FROM supplier_groups sg
		LEFT JOIN stores st ON st.store_id = sg.store_id
		LEFT JOIN suppliers s ON s.supplier_group_id = sg.id
	`
	}

	conditions := make([]string, 0, 2)
	args := make([]interface{}, 0, 4)

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		conditions = append(conditions, "(LOWER(sg.group_code) LIKE ? OR LOWER(sg.group_name) LIKE ? OR LOWER(COALESCE(sg.description, '')) LIKE ?)")
		args = append(args, keyword, keyword, keyword)
	}

	switch filter.Status {
	case "active":
		conditions = append(conditions, "sg.is_active = 1")
	case "inactive":
		conditions = append(conditions, "sg.is_active = 0")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if countOnly {
		return query, args
	}

	query += `
		GROUP BY
			sg.id,
			sg.store_id,
			st.store_name,
			sg.group_code,
			sg.group_name,
			sg.description,
			sg.is_active,
			sg.created_at,
			sg.updated_at
	`

	return query, args
}

func buildSupplierGroupOrderClause(filter models.SupplierGroupListFilter) string {
	switch filter.Sort {
	case "name":
		return " ORDER BY sg.group_name ASC, sg.id ASC"
	case "code":
		return " ORDER BY sg.group_code ASC, sg.id ASC"
	case "suppliers":
		return " ORDER BY supplier_count DESC, sg.group_name ASC"
	default:
		return " ORDER BY COALESCE(sg.updated_at, sg.created_at) DESC, sg.id DESC"
	}
}

func (r *SupplierGroupRepository) GetStats() (models.SupplierGroupStats, error) {
	var stats models.SupplierGroupStats

	if err := r.DB.QueryRow(`
		SELECT
			COUNT(*) AS total_groups,
			COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0) AS active_groups,
			COALESCE(SUM(CASE WHEN is_active = 0 THEN 1 ELSE 0 END), 0) AS inactive_groups
		FROM supplier_groups
	`).Scan(&stats.TotalGroups, &stats.ActiveGroups, &stats.InactiveGroups); err != nil {
		return stats, err
	}

	if err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM suppliers
		WHERE supplier_group_id IS NOT NULL
	`).Scan(&stats.LinkedSuppliers); err != nil {
		return stats, err
	}

	return stats, nil
}

func (r *SupplierGroupRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM supplier_groups WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *SupplierGroupRepository) ExistsByCode(code string, storeID int, ignoreID int) (bool, error) {
	var (
		count int
		err   error
	)

	if ignoreID > 0 {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM supplier_groups WHERE group_code = ? AND store_id = ? AND id <> ?`, code, storeID, ignoreID).Scan(&count)
	} else {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM supplier_groups WHERE group_code = ? AND store_id = ?`, code, storeID).Scan(&count)
	}

	return count > 0, err
}

func (r *SupplierGroupRepository) Create(input models.SupplierGroupCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO supplier_groups (
			store_id,
			group_code,
			group_name,
			description,
			is_active
		) VALUES (?, ?, ?, ?, ?)
	`,
		nullableInt(input.StoreID),
		input.GroupCode,
		input.GroupName,
		nullableString(input.Description),
		boolToInt(input.IsActive),
	)

	return err
}

func (r *SupplierGroupRepository) Update(input models.SupplierGroupUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE supplier_groups
		SET
			store_id = ?,
			group_code = ?,
			group_name = ?,
			description = ?,
			is_active = ?
		WHERE id = ?
	`,
		nullableInt(input.StoreID),
		input.GroupCode,
		input.GroupName,
		nullableString(input.Description),
		boolToInt(input.IsActive),
		input.ID,
	)

	return err
}

func (r *SupplierGroupRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM supplier_groups WHERE id = ?`, id)
	return err
}

func scanSupplierGroup(scanner interface {
	Scan(dest ...interface{}) error
}) (models.SupplierGroup, error) {
	var (
		group     models.SupplierGroup
		isActive  int
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	err := scanner.Scan(
		&group.ID,
		&group.StoreID,
		&group.StoreName,
		&group.GroupCode,
		&group.GroupName,
		&group.Description,
		&isActive,
		&createdAt,
		&updatedAt,
		&group.SupplierCount,
	)
	if err != nil {
		return group, err
	}

	group.IsActive = isActive == 1
	if group.IsActive {
		group.StatusLabel = "Aktif"
	} else {
		group.StatusLabel = "Nonaktif"
	}

	if createdAt.Valid {
		group.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		group.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
	} else {
		group.CreatedAt = "-"
		group.CreatedAtDisplay = "-"
	}

	if updatedAt.Valid {
		group.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		group.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
	} else {
		group.UpdatedAt = "-"
		group.UpdatedAtDisplay = "-"
	}

	return group, nil
}
