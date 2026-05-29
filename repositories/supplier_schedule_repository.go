package repositories

import (
	"database/sql"
	"fmt"
	"gobase-app/models"
	"strings"
)

type SupplierScheduleRepository struct {
	DB *sql.DB
}

func (r *SupplierScheduleRepository) GetAll(filter models.SupplierScheduleListFilter, allowedStoreIDs []int) ([]models.SupplierSchedule, error) {
	scopeClause, scopeArgs := buildScheduleStoreScopeClause("sch", allowedStoreIDs)
	if scopeClause == "" {
		scopeClause = "1=0"
	}

	query := `
		SELECT
			sch.id,
			sch.store_id,
			COALESCE(st.store_name, '') AS store_name,
			sch.supplier_id,
			COALESCE(sp.supplier_code, '') AS supplier_code,
			COALESCE(sp.supplier_name, '') AS supplier_name,
			sch.day_of_week,
			COALESCE(TIME_FORMAT(sch.so_time, '%H:%i'), '') AS so_time,
			COALESCE(sch.sequence_no, 1) AS sequence_no,
			sch.is_active,
			COALESCE(sch.notes, '') AS notes,
			sch.updated_at
		FROM supplier_so_schedules sch
		INNER JOIN suppliers sp ON sp.id = sch.supplier_id
		LEFT JOIN stores st ON st.store_id = sch.store_id
	`

	conditions := make([]string, 0, 6)
	args := make([]interface{}, 0, 12)

	conditions = append(conditions, scopeClause)
	args = append(args, scopeArgs...)

	if search := strings.TrimSpace(strings.ToLower(filter.Search)); search != "" {
		conditions = append(conditions, "(LOWER(COALESCE(sp.supplier_name, '')) LIKE ? OR LOWER(COALESCE(sp.supplier_code, '')) LIKE ? OR LOWER(COALESCE(sch.notes, '')) LIKE ?)")
		keyword := "%" + search + "%"
		args = append(args, keyword, keyword, keyword)
	}

	if filter.StoreID > 0 {
		conditions = append(conditions, "sch.store_id = ?")
		args = append(args, filter.StoreID)
	}

	if filter.DayOfWeek >= 1 && filter.DayOfWeek <= 7 {
		conditions = append(conditions, "sch.day_of_week = ?")
		args = append(args, filter.DayOfWeek)
	}

	switch filter.Status {
	case "active":
		conditions = append(conditions, "sch.is_active = 1")
	case "inactive":
		conditions = append(conditions, "sch.is_active = 0")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY sch.day_of_week ASC, sch.sequence_no ASC, sch.so_time ASC, sp.supplier_name ASC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]models.SupplierSchedule, 0)
	for rows.Next() {
		schedule, err := scanSupplierSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

func (r *SupplierScheduleRepository) GetByID(id int) (models.SupplierSchedule, error) {
	row := r.DB.QueryRow(`
		SELECT
			sch.id,
			sch.store_id,
			COALESCE(st.store_name, '') AS store_name,
			sch.supplier_id,
			COALESCE(sp.supplier_code, '') AS supplier_code,
			COALESCE(sp.supplier_name, '') AS supplier_name,
			sch.day_of_week,
			COALESCE(TIME_FORMAT(sch.so_time, '%H:%i'), '') AS so_time,
			COALESCE(sch.sequence_no, 1) AS sequence_no,
			sch.is_active,
			COALESCE(sch.notes, '') AS notes,
			sch.updated_at
		FROM supplier_so_schedules sch
		INNER JOIN suppliers sp ON sp.id = sch.supplier_id
		LEFT JOIN stores st ON st.store_id = sch.store_id
		WHERE sch.id = ?
		LIMIT 1
	`, id)

	return scanSupplierSchedule(row)
}

func (r *SupplierScheduleRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM supplier_so_schedules WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *SupplierScheduleRepository) ExistsDuplicate(storeID int, supplierID int, dayOfWeek int, ignoreID int) (bool, error) {
	var (
		count int
		err   error
	)

	if ignoreID > 0 {
		err = r.DB.QueryRow(`
			SELECT COUNT(1)
			FROM supplier_so_schedules
			WHERE store_id = ? AND supplier_id = ? AND day_of_week = ? AND id <> ?
		`, storeID, supplierID, dayOfWeek, ignoreID).Scan(&count)
	} else {
		err = r.DB.QueryRow(`
			SELECT COUNT(1)
			FROM supplier_so_schedules
			WHERE store_id = ? AND supplier_id = ? AND day_of_week = ?
		`, storeID, supplierID, dayOfWeek).Scan(&count)
	}

	return count > 0, err
}

func (r *SupplierScheduleRepository) SupplierBelongsToStore(storeID int, supplierID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM suppliers
		WHERE id = ? AND store_id = ?
	`, supplierID, storeID).Scan(&count)
	return count > 0, err
}

func (r *SupplierScheduleRepository) Create(input models.SupplierScheduleCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO supplier_so_schedules (
			store_id,
			supplier_id,
			day_of_week,
			so_time,
			sequence_no,
			is_active,
			notes
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		input.StoreID,
		input.SupplierID,
		input.DayOfWeek,
		nullableScheduleTime(input.SOTime),
		input.SequenceNo,
		boolToInt(input.IsActive),
		nullableScheduleText(input.Notes),
	)
	return err
}

func (r *SupplierScheduleRepository) Update(input models.SupplierScheduleUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE supplier_so_schedules
		SET
			store_id = ?,
			supplier_id = ?,
			day_of_week = ?,
			so_time = ?,
			sequence_no = ?,
			is_active = ?,
			notes = ?
		WHERE id = ?
	`,
		input.StoreID,
		input.SupplierID,
		input.DayOfWeek,
		nullableScheduleTime(input.SOTime),
		input.SequenceNo,
		boolToInt(input.IsActive),
		nullableScheduleText(input.Notes),
		input.ID,
	)
	return err
}

func (r *SupplierScheduleRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM supplier_so_schedules WHERE id = ?`, id)
	return err
}

func (r *SupplierScheduleRepository) GetStats(allowedStoreIDs []int) (models.SupplierScheduleStats, error) {
	scopeSuppliersClause, scopeSuppliersArgs := buildScheduleStoreScopeClause("s", allowedStoreIDs)
	scopeSchedulesClause, scopeSchedulesArgs := buildScheduleStoreScopeClause("sch", allowedStoreIDs)
	if scopeSuppliersClause == "" {
		scopeSuppliersClause = "1=0"
	}
	if scopeSchedulesClause == "" {
		scopeSchedulesClause = "1=0"
	}

	query := fmt.Sprintf(`
		SELECT
			(
				SELECT COUNT(1)
				FROM suppliers s
				WHERE s.is_active = 1
				  AND %s
			) AS total_suppliers,
			(
				SELECT COUNT(DISTINCT s.id)
				FROM suppliers s
				WHERE s.is_active = 1
				  AND %s
				  AND EXISTS (
					SELECT 1
					FROM supplier_so_schedules sch
					WHERE sch.store_id = s.store_id
					  AND sch.supplier_id = s.id
					  AND sch.is_active = 1
				  )
			) AS scheduled_suppliers,
			(
				SELECT COUNT(1)
				FROM supplier_so_schedules sch
				WHERE %s
			) AS total_schedules
	`, scopeSuppliersClause, scopeSuppliersClause, scopeSchedulesClause)

	args := make([]interface{}, 0, len(scopeSuppliersArgs)*2+len(scopeSchedulesArgs))
	args = append(args, scopeSuppliersArgs...)
	args = append(args, scopeSuppliersArgs...)
	args = append(args, scopeSchedulesArgs...)

	var stats models.SupplierScheduleStats
	if err := r.DB.QueryRow(query, args...).Scan(
		&stats.TotalSuppliers,
		&stats.ScheduledSuppliers,
		&stats.TotalSchedules,
	); err != nil {
		return stats, err
	}
	stats.UnscheduledSuppliers = stats.TotalSuppliers - stats.ScheduledSuppliers
	if stats.UnscheduledSuppliers < 0 {
		stats.UnscheduledSuppliers = 0
	}

	return stats, nil
}

func (r *SupplierScheduleRepository) GetSupplierOptions(allowedStoreIDs []int) ([]models.Supplier, error) {
	scopeClause, scopeArgs := buildScheduleStoreScopeClause("s", allowedStoreIDs)
	if scopeClause == "" {
		scopeClause = "1=0"
	}

	query := `
		SELECT
			s.id,
			s.supplier_name,
			COALESCE(s.store_id, 0) AS store_id,
			COALESCE(st.store_name, '') AS store_name,
			COALESCE(s.supplier_code, '') AS supplier_code
		FROM suppliers s
		LEFT JOIN stores st ON st.store_id = s.store_id
		WHERE s.is_active = 1
		  AND ` + scopeClause + `
		ORDER BY st.store_name ASC, s.supplier_name ASC
	`

	rows, err := r.DB.Query(query, scopeArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]models.Supplier, 0)
	for rows.Next() {
		var supplier models.Supplier
		if err := rows.Scan(
			&supplier.ID,
			&supplier.SupplierName,
			&supplier.StoreID,
			&supplier.StoreName,
			&supplier.SupplierCode,
		); err != nil {
			return nil, err
		}
		options = append(options, supplier)
	}

	return options, rows.Err()
}

func scanSupplierSchedule(scanner interface {
	Scan(dest ...interface{}) error
}) (models.SupplierSchedule, error) {
	var (
		schedule  models.SupplierSchedule
		isActive  int
		soTime    sql.NullString
		notes     sql.NullString
		updatedAt sql.NullTime
	)

	if err := scanner.Scan(
		&schedule.ID,
		&schedule.StoreID,
		&schedule.StoreName,
		&schedule.SupplierID,
		&schedule.SupplierCode,
		&schedule.SupplierName,
		&schedule.DayOfWeek,
		&soTime,
		&schedule.SequenceNo,
		&isActive,
		&notes,
		&updatedAt,
	); err != nil {
		return schedule, err
	}

	schedule.DayName = supplierScheduleDayName(schedule.DayOfWeek)
	schedule.IsActive = isActive == 1
	if schedule.IsActive {
		schedule.StatusLabel = "Aktif"
	} else {
		schedule.StatusLabel = "Nonaktif"
	}

	if soTime.Valid {
		schedule.SOTime = strings.TrimSpace(soTime.String)
	}
	if notes.Valid {
		schedule.Notes = strings.TrimSpace(notes.String)
	}

	if updatedAt.Valid {
		schedule.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		schedule.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
	} else {
		schedule.UpdatedAt = "-"
		schedule.UpdatedAtDisplay = "-"
	}

	return schedule, nil
}

func supplierScheduleDayName(dayOfWeek int) string {
	switch dayOfWeek {
	case 1:
		return "Senin"
	case 2:
		return "Selasa"
	case 3:
		return "Rabu"
	case 4:
		return "Kamis"
	case 5:
		return "Jumat"
	case 6:
		return "Sabtu"
	case 7:
		return "Minggu"
	default:
		return "-"
	}
}

func nullableScheduleText(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableScheduleTime(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func buildScheduleStoreScopeClause(alias string, allowedStoreIDs []int) (string, []interface{}) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		alias = "s"
	}

	ids := uniquePositiveInts(allowedStoreIDs)
	if len(ids) == 0 {
		return "", []interface{}{}
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	return fmt.Sprintf("%s.store_id IN (%s)", alias, strings.Join(placeholders, ",")), args
}

func uniquePositiveInts(values []int) []int {
	result := make([]int, 0, len(values))
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
