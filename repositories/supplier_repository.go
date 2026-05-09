package repositories

import (
	"database/sql"
	"fmt"
	"gobase-app/models"
	"strings"
)

type SupplierRepository struct {
	DB *sql.DB
}

func (r *SupplierRepository) GetAll(filter models.SupplierListFilter) ([]models.Supplier, error) {
	query, args := buildSupplierListQuery(filter, false)
	query += buildSupplierOrderClause(filter)
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		supplier, err := scanSupplier(rows)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, rows.Err()
}

func (r *SupplierRepository) CountAll(filter models.SupplierListFilter) (int, error) {
	query, args := buildSupplierListQuery(filter, true)

	var total int
	err := r.DB.QueryRow(query, args...).Scan(&total)
	return total, err
}

func buildSupplierListQuery(filter models.SupplierListFilter, countOnly bool) (string, []interface{}) {
	query := `
`
	if countOnly {
		query = `
		SELECT COUNT(*)
		FROM suppliers s
	`
	} else {
		query = `
		SELECT
			s.id,
			COALESCE(s.supplier_group_id, 0) AS supplier_group_id,
			COALESCE(sg.group_name, '') AS supplier_group_name,
			s.supplier_code,
			s.supplier_name,
			COALESCE(s.supplier_type, '') AS supplier_type,
			COALESCE(s.address, '') AS address,
			COALESCE(s.phone, '') AS phone,
			COALESCE(s.email, '') AS email,
			COALESCE(s.pic_name, '') AS pic_name,
			COALESCE(s.payment_term_days, 0) AS payment_term_days,
			s.is_active,
			s.created_at,
			s.updated_at,
			COUNT(DISTINCT CASE WHEN ps.is_active = 1 THEN ps.product_id END) AS product_count,
			MAX(scs.session_date) AS last_so_date
		FROM suppliers s
		LEFT JOIN supplier_groups sg ON sg.id = s.supplier_group_id
		LEFT JOIN product_suppliers ps ON ps.supplier_id = s.id
		LEFT JOIN stock_check_sessions scs ON scs.supplier_id = s.id
	`
	}

	conditions := make([]string, 0, 3)
	args := make([]interface{}, 0, 6)

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		if filter.SearchMode == "name_code" {
			conditions = append(conditions, `(LOWER(s.supplier_code) LIKE ? OR LOWER(s.supplier_name) LIKE ?)`)
			args = append(args, keyword, keyword)
		} else {
			conditions = append(conditions, `(LOWER(s.supplier_code) LIKE ? OR LOWER(s.supplier_name) LIKE ? OR LOWER(COALESCE(s.pic_name, '')) LIKE ? OR LOWER(COALESCE(s.phone, '')) LIKE ?)`)
			args = append(args, keyword, keyword, keyword, keyword)
		}
	}

	switch filter.Status {
	case "active":
		conditions = append(conditions, "s.is_active = 1")
	case "inactive":
		conditions = append(conditions, "s.is_active = 0")
	}

	if filter.Type != "" {
		conditions = append(conditions, "LOWER(COALESCE(s.supplier_type, '')) = ?")
		args = append(args, strings.ToLower(filter.Type))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if countOnly {
		return query, args
	}

	query += `
		GROUP BY
			s.id,
			s.supplier_group_id,
			sg.group_name,
			s.supplier_code,
			s.supplier_name,
			s.supplier_type,
			s.address,
			s.phone,
			s.email,
			s.pic_name,
			s.payment_term_days,
			s.is_active,
			s.created_at,
			s.updated_at
	`

	return query, args
}

func buildSupplierOrderClause(filter models.SupplierListFilter) string {
	switch filter.Sort {
	case "name":
		return " ORDER BY s.supplier_name ASC, s.id ASC"
	case "code":
		return " ORDER BY s.supplier_code ASC, s.id ASC"
	case "products":
		return " ORDER BY product_count DESC, s.supplier_name ASC"
	default:
		return " ORDER BY COALESCE(s.updated_at, s.created_at) DESC, s.id DESC"
	}
}

func (r *SupplierRepository) GetByID(id int) (models.Supplier, error) {
	query := `
		SELECT
			s.id,
			COALESCE(s.supplier_group_id, 0) AS supplier_group_id,
			COALESCE(sg.group_name, '') AS supplier_group_name,
			s.supplier_code,
			s.supplier_name,
			COALESCE(s.supplier_type, '') AS supplier_type,
			COALESCE(s.address, '') AS address,
			COALESCE(s.phone, '') AS phone,
			COALESCE(s.email, '') AS email,
			COALESCE(s.pic_name, '') AS pic_name,
			COALESCE(s.payment_term_days, 0) AS payment_term_days,
			s.is_active,
			s.created_at,
			s.updated_at,
			COUNT(DISTINCT CASE WHEN ps.is_active = 1 THEN ps.product_id END) AS product_count,
			MAX(scs.session_date) AS last_so_date
		FROM suppliers s
		LEFT JOIN supplier_groups sg ON sg.id = s.supplier_group_id
		LEFT JOIN product_suppliers ps ON ps.supplier_id = s.id
		LEFT JOIN stock_check_sessions scs ON scs.supplier_id = s.id
		WHERE s.id = ?
		GROUP BY
			s.id,
			s.supplier_group_id,
			sg.group_name,
			s.supplier_code,
			s.supplier_name,
			s.supplier_type,
			s.address,
			s.phone,
			s.email,
			s.pic_name,
			s.payment_term_days,
			s.is_active,
			s.created_at,
			s.updated_at
	`

	row := r.DB.QueryRow(query, id)
	return scanSupplier(row)
}

func (r *SupplierRepository) GetSuppliedProducts(supplierID int) ([]models.SupplierProduct, error) {
	rows, err := r.DB.Query(`
		SELECT
			p.id,
			p.product_code,
			COALESCE(p.barcode, '') AS barcode,
			p.product_name,
			COALESCE(pc.category_name, '') AS category_name,
			ps.last_price,
			ps.moq,
			ps.pack_size,
			ps.lead_time_days,
			ps.priority_no,
			ps.is_primary,
			ps.is_active,
			ps.updated_at
		FROM product_suppliers ps
		INNER JOIN products p ON p.id = ps.product_id
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		WHERE ps.supplier_id = ?
		ORDER BY ps.is_primary DESC, ps.priority_no ASC, p.product_name ASC
	`, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.SupplierProduct
	for rows.Next() {
		var (
			product   models.SupplierProduct
			lastPrice sql.NullFloat64
			moq       sql.NullFloat64
			packSize  sql.NullFloat64
			isPrimary int
			isActive  int
			updatedAt sql.NullTime
		)

		if err := rows.Scan(
			&product.ProductID,
			&product.ProductCode,
			&product.Barcode,
			&product.ProductName,
			&product.CategoryName,
			&lastPrice,
			&moq,
			&packSize,
			&product.LeadTimeDays,
			&product.PriorityNo,
			&isPrimary,
			&isActive,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		if lastPrice.Valid {
			product.LastPrice = lastPrice.Float64
		}
		if moq.Valid {
			product.MOQ = moq.Float64
		}
		if packSize.Valid {
			product.PackSize = packSize.Float64
		}

		product.LastPriceDisplay = fmt.Sprintf("Rp %s", formatDecimal(product.LastPrice))
		product.MOQDisplay = formatDecimal(product.MOQ)
		product.PackSizeDisplay = formatDecimal(product.PackSize)
		product.IsPrimary = isPrimary == 1
		product.IsActive = isActive == 1
		if product.IsActive {
			product.StatusLabel = "Aktif"
		} else {
			product.StatusLabel = "Nonaktif"
		}
		if updatedAt.Valid {
			product.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
			product.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
		} else {
			product.UpdatedAt = "-"
			product.UpdatedAtDisplay = "-"
		}

		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *SupplierRepository) GetAvailableProductOptions(supplierID int) ([]models.SupplierProductOption, error) {
	rows, err := r.DB.Query(`
		SELECT
			p.id,
			p.product_code,
			COALESCE(p.barcode, '') AS barcode,
			p.product_name,
			COALESCE(pc.category_name, '') AS category_name,
			COALESCE(u.unit_name, '') AS unit_name
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN units u ON u.id = p.unit_id
		WHERE p.is_active = 1
			AND NOT EXISTS (
				SELECT 1
				FROM product_suppliers ps
				WHERE ps.product_id = p.id
					AND ps.supplier_id = ?
					AND ps.is_active = 1
			)
		ORDER BY p.product_name ASC, p.id ASC
	`, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.SupplierProductOption
	for rows.Next() {
		var product models.SupplierProductOption
		if err := rows.Scan(
			&product.ID,
			&product.ProductCode,
			&product.Barcode,
			&product.ProductName,
			&product.CategoryName,
			&product.UnitName,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *SupplierRepository) GetStats() (models.SupplierStats, error) {
	var stats models.SupplierStats

	if err := r.DB.QueryRow(`
		SELECT
			COUNT(*) AS total_suppliers,
			COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0) AS active_suppliers,
			COALESCE(SUM(CASE WHEN is_active = 0 THEN 1 ELSE 0 END), 0) AS inactive_suppliers
		FROM suppliers
	`).Scan(&stats.TotalSuppliers, &stats.ActiveSuppliers, &stats.InactiveSuppliers); err != nil {
		return stats, err
	}

	if err := r.DB.QueryRow(`
		SELECT COUNT(DISTINCT product_id)
		FROM product_suppliers
		WHERE is_active = 1
	`).Scan(&stats.LinkedProducts); err != nil {
		return stats, err
	}

	return stats, nil
}

func (r *SupplierRepository) GetActiveGroups() ([]models.SupplierGroup, error) {
	rows, err := r.DB.Query(`
		SELECT id, group_name
		FROM supplier_groups
		WHERE is_active = 1
		ORDER BY group_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.SupplierGroup
	for rows.Next() {
		var group models.SupplierGroup
		if err := rows.Scan(&group.ID, &group.GroupName); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

func (r *SupplierRepository) GetTypes() ([]string, error) {
	rows, err := r.DB.Query(`
		SELECT DISTINCT supplier_type
		FROM suppliers
		WHERE supplier_type IS NOT NULL AND supplier_type <> ''
		ORDER BY supplier_type ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var supplierType string
		if err := rows.Scan(&supplierType); err != nil {
			return nil, err
		}
		types = append(types, supplierType)
	}

	return types, rows.Err()
}

func (r *SupplierRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM suppliers WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *SupplierRepository) ProductExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM products WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *SupplierRepository) ProductSupplyExists(supplierID, productID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM product_suppliers
		WHERE supplier_id = ? AND product_id = ?
	`, supplierID, productID).Scan(&count)
	return count > 0, err
}

func (r *SupplierRepository) ExistsByCode(code string, ignoreID int) (bool, error) {
	var (
		count int
		err   error
	)

	if ignoreID > 0 {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM suppliers WHERE supplier_code = ? AND id <> ?`, code, ignoreID).Scan(&count)
	} else {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM suppliers WHERE supplier_code = ?`, code).Scan(&count)
	}

	return count > 0, err
}

func (r *SupplierRepository) Create(input models.SupplierCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO suppliers (
			supplier_group_id,
			supplier_code,
			supplier_name,
			supplier_type,
			address,
			phone,
			email,
			pic_name,
			payment_term_days,
			is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		nullableInt(input.SupplierGroupID),
		input.SupplierCode,
		input.SupplierName,
		nullableString(input.SupplierType),
		nullableString(input.Address),
		nullableString(input.Phone),
		nullableString(input.Email),
		nullableString(input.PICName),
		input.PaymentTermDays,
		boolToInt(input.IsActive),
	)

	return err
}

func (r *SupplierRepository) Update(input models.SupplierUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE suppliers
		SET
			supplier_group_id = ?,
			supplier_code = ?,
			supplier_name = ?,
			supplier_type = ?,
			address = ?,
			phone = ?,
			email = ?,
			pic_name = ?,
			payment_term_days = ?,
			is_active = ?
		WHERE id = ?
	`,
		nullableInt(input.SupplierGroupID),
		input.SupplierCode,
		input.SupplierName,
		nullableString(input.SupplierType),
		nullableString(input.Address),
		nullableString(input.Phone),
		nullableString(input.Email),
		nullableString(input.PICName),
		input.PaymentTermDays,
		boolToInt(input.IsActive),
		input.ID,
	)

	return err
}

func (r *SupplierRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM suppliers WHERE id = ?`, id)
	return err
}

func (r *SupplierRepository) UpsertProductSupply(input models.SupplierProductCreateInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if input.IsPrimary {
		if _, err := tx.Exec(`
			UPDATE product_suppliers
			SET is_primary = 0
			WHERE product_id = ?
		`, input.ProductID); err != nil {
			tx.Rollback()
			return err
		}
	}

	var (
		existingID       int
		existingPriority int
	)

	err = tx.QueryRow(`
		SELECT id, priority_no
		FROM product_suppliers
		WHERE product_id = ? AND supplier_id = ?
		LIMIT 1
	`, input.ProductID, input.SupplierID).Scan(&existingID, &existingPriority)

	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	priorityNo := existingPriority
	if input.IsPrimary {
		priorityNo = 1
	}
	if priorityNo <= 0 {
		if err := tx.QueryRow(`
			SELECT COALESCE(MAX(priority_no), 0) + 1
			FROM product_suppliers
			WHERE product_id = ?
		`, input.ProductID).Scan(&priorityNo); err != nil {
			tx.Rollback()
			return err
		}
		if priorityNo <= 0 {
			priorityNo = 1
		}
	}

	if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO product_suppliers (
				product_id,
				supplier_id,
				is_primary,
				priority_no,
				last_price,
				moq,
				lead_time_days,
				pack_size,
				is_active
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			input.ProductID,
			input.SupplierID,
			boolToInt(input.IsPrimary),
			priorityNo,
			input.LastPrice,
			input.MOQ,
			input.LeadTimeDays,
			input.PackSize,
			boolToInt(input.IsActive),
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		if err := normalizePrimarySupplierForProduct(tx, input.ProductID); err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit()
	}

	_, err = tx.Exec(`
		UPDATE product_suppliers
		SET
			is_primary = ?,
			priority_no = ?,
			last_price = ?,
			moq = ?,
			lead_time_days = ?,
			pack_size = ?,
			is_active = ?
		WHERE id = ?
	`,
		boolToInt(input.IsPrimary),
		priorityNo,
		input.LastPrice,
		input.MOQ,
		input.LeadTimeDays,
		input.PackSize,
		boolToInt(input.IsActive),
		existingID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := normalizePrimarySupplierForProduct(tx, input.ProductID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *SupplierRepository) DeleteProductSupply(input models.SupplierProductDeleteInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM product_suppliers
		WHERE supplier_id = ? AND product_id = ?
	`, input.SupplierID, input.ProductID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := normalizePrimarySupplierForProduct(tx, input.ProductID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func normalizePrimarySupplierForProduct(tx *sql.Tx, productID int) error {
	if productID <= 0 {
		return nil
	}

	var primaryCount int
	if err := tx.QueryRow(`
		SELECT COUNT(1)
		FROM product_suppliers
		WHERE product_id = ? AND is_active = 1 AND is_primary = 1
	`, productID).Scan(&primaryCount); err != nil {
		return err
	}

	if primaryCount > 0 {
		return nil
	}

	var supplierID int
	err := tx.QueryRow(`
		SELECT supplier_id
		FROM product_suppliers
		WHERE product_id = ? AND is_active = 1
		ORDER BY priority_no ASC, id ASC
		LIMIT 1
	`, productID).Scan(&supplierID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE product_suppliers
		SET is_primary = CASE WHEN supplier_id = ? THEN 1 ELSE 0 END
		WHERE product_id = ?
	`, supplierID, productID)
	return err
}

func scanSupplier(scanner interface {
	Scan(dest ...interface{}) error
}) (models.Supplier, error) {
	var (
		supplier   models.Supplier
		isActive   int
		createdAt  sql.NullTime
		updatedAt  sql.NullTime
		lastSODate sql.NullTime
	)

	err := scanner.Scan(
		&supplier.ID,
		&supplier.SupplierGroupID,
		&supplier.SupplierGroupName,
		&supplier.SupplierCode,
		&supplier.SupplierName,
		&supplier.SupplierType,
		&supplier.Address,
		&supplier.Phone,
		&supplier.Email,
		&supplier.PICName,
		&supplier.PaymentTermDays,
		&isActive,
		&createdAt,
		&updatedAt,
		&supplier.ProductCount,
		&lastSODate,
	)
	if err != nil {
		return supplier, err
	}

	supplier.IsActive = isActive == 1
	if supplier.IsActive {
		supplier.StatusLabel = "Aktif"
	} else {
		supplier.StatusLabel = "Nonaktif"
	}

	if createdAt.Valid {
		supplier.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		supplier.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
	} else {
		supplier.CreatedAt = "-"
		supplier.CreatedAtDisplay = "-"
	}

	if updatedAt.Valid {
		supplier.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		supplier.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
	} else {
		supplier.UpdatedAt = "-"
		supplier.UpdatedAtDisplay = "-"
	}

	if lastSODate.Valid {
		supplier.LastSODate = lastSODate.Time.Format("2006-01-02")
		supplier.LastSODateDisplay = lastSODate.Time.Format("02 Jan 2006")
	} else {
		supplier.LastSODate = "-"
		supplier.LastSODateDisplay = "Belum ada SO"
	}

	return supplier, nil
}

func nullableString(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableInt(value int) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func formatDecimal(value float64) string {
	return fmt.Sprintf("%0.2f", value)
}
