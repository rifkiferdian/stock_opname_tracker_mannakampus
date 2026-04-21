package repositories

import (
	"database/sql"
	"fmt"
	"gobase-app/models"
	"strings"
)

type ProductRepository struct {
	DB *sql.DB
}

func (r *ProductRepository) GetAll(filter models.ProductListFilter) ([]models.Product, error) {
	query, args := buildProductListQuery(filter, false)
	query += buildProductOrderClause(filter)
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *ProductRepository) CountAll(filter models.ProductListFilter) (int, error) {
	query, args := buildProductListQuery(filter, true)

	var total int
	err := r.DB.QueryRow(query, args...).Scan(&total)
	return total, err
}

func (r *ProductRepository) GetStats() (models.ProductStats, error) {
	var stats models.ProductStats

	if err := r.DB.QueryRow(`
		SELECT
			COUNT(*) AS total_products,
			COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0) AS active_products,
			COALESCE(SUM(CASE WHEN is_active = 0 THEN 1 ELSE 0 END), 0) AS inactive_products
		FROM products
	`).Scan(&stats.TotalProducts, &stats.ActiveProducts, &stats.InactiveProducts); err != nil {
		return stats, err
	}

	if err := r.DB.QueryRow(`
		SELECT COUNT(DISTINCT supplier_id)
		FROM product_suppliers
		WHERE is_active = 1
	`).Scan(&stats.LinkedSuppliers); err != nil {
		return stats, err
	}

	return stats, nil
}

func (r *ProductRepository) GetCategories() ([]models.ProductCategory, error) {
	rows, err := r.DB.Query(`
		SELECT id, category_code, category_name
		FROM product_categories
		WHERE is_active = 1
		ORDER BY category_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.ProductCategory
	for rows.Next() {
		var category models.ProductCategory
		if err := rows.Scan(&category.ID, &category.CategoryCode, &category.CategoryName); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *ProductRepository) GetUnits() ([]models.Unit, error) {
	rows, err := r.DB.Query(`
		SELECT id, unit_code, unit_name, description, created_at, updated_at
		FROM units
		ORDER BY unit_name ASC, id ASC
	`)
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

func (r *ProductRepository) GetSuppliers() ([]models.ProductSupplierOption, error) {
	rows, err := r.DB.Query(`
		SELECT id, supplier_name
		FROM suppliers
		WHERE is_active = 1
		ORDER BY supplier_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.ProductSupplierOption
	for rows.Next() {
		var supplier models.ProductSupplierOption
		if err := rows.Scan(&supplier.ID, &supplier.SupplierName); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, rows.Err()
}

func (r *ProductRepository) GetBrands() ([]string, error) {
	rows, err := r.DB.Query(`
		SELECT DISTINCT brand
		FROM products
		WHERE brand IS NOT NULL AND brand <> ''
		ORDER BY brand ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []string
	for rows.Next() {
		var brand string
		if err := rows.Scan(&brand); err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}

	return brands, rows.Err()
}

func (r *ProductRepository) GetByID(id int) (models.ProductDetail, error) {
	row := r.DB.QueryRow(`
		SELECT
			p.id,
			p.product_code,
			COALESCE(p.barcode, '') AS barcode,
			p.product_name,
			COALESCE(p.category_id, 0) AS category_id,
			COALESCE(pc.category_name, '') AS category_name,
			COALESCE(p.unit_id, 0) AS unit_id,
			COALESCE(u.unit_name, '') AS unit_name,
			COALESCE(p.brand, '') AS brand,
			p.min_stock,
			p.max_stock,
			p.reorder_point,
			p.default_lead_time_days,
			p.pack_size,
			p.is_active,
			COALESCE((
				SELECT ps.supplier_id
				FROM product_suppliers ps
				WHERE ps.product_id = p.id AND ps.is_active = 1
				ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
				LIMIT 1
			), 0) AS primary_supplier_id,
			COALESCE((
				SELECT s.supplier_name
				FROM product_suppliers ps
				INNER JOIN suppliers s ON s.id = ps.supplier_id
				WHERE ps.product_id = p.id AND ps.is_active = 1
				ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
				LIMIT 1
			), '') AS primary_supplier_name,
			COALESCE((
				SELECT ps.last_price
				FROM product_suppliers ps
				WHERE ps.product_id = p.id AND ps.is_active = 1
				ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
				LIMIT 1
			), 0) AS primary_supplier_price,
			p.created_at,
			p.updated_at,
			COALESCE((
				SELECT si.total_qty
				FROM stock_check_session_items si
				INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
				WHERE si.product_id = p.id
				ORDER BY scs.session_date DESC, si.id DESC
				LIMIT 1
			), 0) AS current_stock,
			COALESCE((
				SELECT COALESCE(si.approved_buy_qty, si.suggest_buy_qty, 0)
				FROM stock_check_session_items si
				INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
				WHERE si.product_id = p.id
				ORDER BY scs.session_date DESC, si.id DESC
				LIMIT 1
			), 0) AS on_order_qty,
			COALESCE((
				SELECT COUNT(*)
				FROM product_suppliers ps
				WHERE ps.product_id = p.id AND ps.is_active = 1
			), 0) AS supplier_network_count,
			COALESCE((
				SELECT COUNT(*)
				FROM stock_check_session_items si
				WHERE si.product_id = p.id
			), 0) AS stock_history_count,
			(
				SELECT scs.session_date
				FROM stock_check_session_items si
				INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
				WHERE si.product_id = p.id
				ORDER BY scs.session_date DESC, si.id DESC
				LIMIT 1
			) AS latest_session_date
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN units u ON u.id = p.unit_id
		WHERE p.id = ?
	`, id)

	var (
		detail              models.ProductDetail
		minStock            sql.NullFloat64
		maxStock            sql.NullFloat64
		reorderPoint        sql.NullFloat64
		packSize            sql.NullFloat64
		primarySupplier     sql.NullFloat64
		currentStock        sql.NullFloat64
		onOrderQty          sql.NullFloat64
		isActive            int
		createdAt           sql.NullTime
		updatedAt           sql.NullTime
		latestSessionDate   sql.NullTime
	)

	err := row.Scan(
		&detail.ID,
		&detail.ProductCode,
		&detail.Barcode,
		&detail.ProductName,
		&detail.CategoryID,
		&detail.CategoryName,
		&detail.UnitID,
		&detail.UnitName,
		&detail.Brand,
		&minStock,
		&maxStock,
		&reorderPoint,
		&detail.DefaultLeadTimeDays,
		&packSize,
		&isActive,
		&detail.PrimarySupplierID,
		&detail.PrimarySupplierName,
		&primarySupplier,
		&createdAt,
		&updatedAt,
		&currentStock,
		&onOrderQty,
		&detail.SupplierNetworkCount,
		&detail.StockHistoryCount,
		&latestSessionDate,
	)
	if err != nil {
		return detail, err
	}

	if minStock.Valid {
		detail.MinStock = minStock.Float64
	}
	if maxStock.Valid {
		detail.MaxStock = maxStock.Float64
	}
	if reorderPoint.Valid {
		detail.ReorderPoint = reorderPoint.Float64
	}
	if packSize.Valid {
		detail.PackSize = packSize.Float64
	}
	if primarySupplier.Valid {
		detail.PrimarySupplierPrice = primarySupplier.Float64
	}
	if currentStock.Valid {
		detail.CurrentStock = currentStock.Float64
	}
	if onOrderQty.Valid {
		detail.OnOrderQty = onOrderQty.Float64
	}

	detail.MinStockDisplay = formatProductDecimal(detail.MinStock)
	detail.MaxStockDisplay = formatProductDecimal(detail.MaxStock)
	detail.ReorderPointDisplay = formatProductDecimal(detail.ReorderPoint)
	detail.PackSizeDisplay = formatProductDecimal(detail.PackSize)
	detail.PrimarySupplierPriceDisplay = fmt.Sprintf("Rp %s", formatProductDecimal(detail.PrimarySupplierPrice))
	detail.CurrentStockDisplay = formatProductDecimal(detail.CurrentStock)
	detail.OnOrderQtyDisplay = formatProductDecimal(detail.OnOrderQty)
	detail.IsActive = isActive == 1
	if detail.IsActive {
		detail.StatusLabel = "Aktif"
	} else {
		detail.StatusLabel = "Nonaktif"
	}

	if detail.MaxStock > 0 {
		detail.AvailabilityPercent = int((detail.CurrentStock / detail.MaxStock) * 100)
	}
	if detail.AvailabilityPercent < 0 {
		detail.AvailabilityPercent = 0
	}
	if detail.AvailabilityPercent > 100 {
		detail.AvailabilityPercent = 100
	}

	if createdAt.Valid {
		detail.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		detail.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
	} else {
		detail.CreatedAt = "-"
		detail.CreatedAtDisplay = "-"
	}
	if updatedAt.Valid {
		detail.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		detail.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
	} else {
		detail.UpdatedAt = "-"
		detail.UpdatedAtDisplay = "-"
	}
	if latestSessionDate.Valid {
		detail.LatestSessionDate = latestSessionDate.Time.Format("2006-01-02")
		detail.LatestSessionDisplay = latestSessionDate.Time.Format("02 Jan 2006")
	} else {
		detail.LatestSessionDate = "-"
		detail.LatestSessionDisplay = "-"
	}

	return detail, nil
}

func (r *ProductRepository) GetSupplierNetwork(productID int) ([]models.ProductSupplierNetwork, error) {
	rows, err := r.DB.Query(`
		SELECT
			s.id,
			s.supplier_code,
			s.supplier_name,
			COALESCE(s.supplier_type, '') AS supplier_type,
			COALESCE(s.address, '') AS address,
			ps.last_price,
			ps.moq,
			ps.pack_size,
			ps.lead_time_days,
			ps.priority_no,
			ps.is_primary,
			ps.is_active
		FROM product_suppliers ps
		INNER JOIN suppliers s ON s.id = ps.supplier_id
		WHERE ps.product_id = ?
		ORDER BY ps.is_primary DESC, ps.priority_no ASC, s.supplier_name ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.ProductSupplierNetwork
	for rows.Next() {
		var (
			item      models.ProductSupplierNetwork
			lastPrice sql.NullFloat64
			moq       sql.NullFloat64
			packSize  sql.NullFloat64
			isPrimary int
			isActive  int
		)

		if err := rows.Scan(
			&item.SupplierID,
			&item.SupplierCode,
			&item.SupplierName,
			&item.SupplierType,
			&item.Address,
			&lastPrice,
			&moq,
			&packSize,
			&item.LeadTimeDays,
			&item.PriorityNo,
			&isPrimary,
			&isActive,
		); err != nil {
			return nil, err
		}

		if lastPrice.Valid {
			item.LastPrice = lastPrice.Float64
		}
		if moq.Valid {
			item.MOQ = moq.Float64
		}
		if packSize.Valid {
			item.PackSize = packSize.Float64
		}

		item.LastPriceDisplay = fmt.Sprintf("Rp %s", formatProductDecimal(item.LastPrice))
		item.MOQDisplay = formatProductDecimal(item.MOQ)
		item.PackSizeDisplay = formatProductDecimal(item.PackSize)
		item.IsPrimary = isPrimary == 1
		item.IsActive = isActive == 1
		if item.IsPrimary {
			item.PriorityLabel = "Utama"
		} else {
			item.PriorityLabel = fmt.Sprintf("Prioritas %d", item.PriorityNo)
		}
		if item.IsActive {
			item.StatusLabel = "Aktif"
		} else {
			item.StatusLabel = "Nonaktif"
		}

		suppliers = append(suppliers, item)
	}

	return suppliers, rows.Err()
}

func (r *ProductRepository) GetStockHistory(productID int) ([]models.ProductStockHistory, error) {
	rows, err := r.DB.Query(`
		SELECT
			si.id,
			scs.id,
			scs.session_number,
			scs.session_date,
			st.store_name,
			COALESCE(u.name, '') AS checker_name,
			si.qty_store,
			si.qty_warehouse,
			((si.qty_store + si.qty_warehouse) - (si.system_qty_store + si.system_qty_warehouse)) AS discrepancy,
			si.suggest_buy_qty,
			COALESCE(si.approved_buy_qty, 0) AS approved_buy_qty,
			COALESCE(si.checker_notes, '') AS checker_notes,
			si.status
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		INNER JOIN stores st ON st.store_id = scs.store_id
		LEFT JOIN users u ON u.id = si.created_by
		WHERE si.product_id = ?
		ORDER BY scs.session_date DESC, si.id DESC
		LIMIT 10
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []models.ProductStockHistory
	for rows.Next() {
		var (
			item           models.ProductStockHistory
			sessionDate    sql.NullTime
			qtyStore       sql.NullFloat64
			qtyWarehouse   sql.NullFloat64
			discrepancy    sql.NullFloat64
			suggestBuyQty  sql.NullFloat64
			approvedBuyQty sql.NullFloat64
		)

		if err := rows.Scan(
			&item.ItemID,
			&item.SessionID,
			&item.SessionNumber,
			&sessionDate,
			&item.StoreName,
			&item.CheckerName,
			&qtyStore,
			&qtyWarehouse,
			&discrepancy,
			&suggestBuyQty,
			&approvedBuyQty,
			&item.CheckerNotes,
			&item.Status,
		); err != nil {
			return nil, err
		}

		if qtyStore.Valid {
			item.QtyStore = qtyStore.Float64
		}
		if qtyWarehouse.Valid {
			item.QtyWarehouse = qtyWarehouse.Float64
		}
		if discrepancy.Valid {
			item.Discrepancy = discrepancy.Float64
		}
		if suggestBuyQty.Valid {
			item.SuggestBuyQty = suggestBuyQty.Float64
		}
		if approvedBuyQty.Valid {
			item.ApprovedBuyQty = approvedBuyQty.Float64
		}

		item.QtyStoreDisplay = formatProductDecimal(item.QtyStore)
		item.QtyWarehouseDisplay = formatProductDecimal(item.QtyWarehouse)
		item.DiscrepancyDisplay = formatSignedProductDecimal(item.Discrepancy)
		item.SuggestBuyQtyDisplay = formatProductDecimal(item.SuggestBuyQty)
		item.ApprovedBuyQtyDisplay = formatProductDecimal(item.ApprovedBuyQty)
		item.StatusLabel = formatProductStatusLabel(item.Status)
		item.StatusBadgeClass = productStatusBadgeClass(item.Status)

		if sessionDate.Valid {
			item.SessionDate = sessionDate.Time.Format("2006-01-02")
			item.SessionDateDisplay = sessionDate.Time.Format("02 Jan 2006")
		} else {
			item.SessionDate = "-"
			item.SessionDateDisplay = "-"
		}

		if item.CheckerName == "" {
			item.CheckerName = "-"
		}

		histories = append(histories, item)
	}

	return histories, rows.Err()
}

func (r *ProductRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM products WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *ProductRepository) ExistsByCode(code string, ignoreID int) (bool, error) {
	var (
		count int
		err   error
	)

	if ignoreID > 0 {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM products WHERE product_code = ? AND id <> ?`, code, ignoreID).Scan(&count)
	} else {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM products WHERE product_code = ?`, code).Scan(&count)
	}

	return count > 0, err
}

func (r *ProductRepository) Create(input models.ProductCreateInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec(`
		INSERT INTO products (
			product_code,
			barcode,
			product_name,
			category_id,
			unit_id,
			brand,
			min_stock,
			max_stock,
			reorder_point,
			default_lead_time_days,
			pack_size,
			is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		input.ProductCode,
		nullableString(input.Barcode),
		input.ProductName,
		nullableInt(input.CategoryID),
		nullableInt(input.UnitID),
		nullableString(input.Brand),
		input.MinStock,
		input.MaxStock,
		input.ReorderPoint,
		input.DefaultLeadTimeDays,
		input.PackSize,
		boolToInt(input.IsActive),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	productID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := upsertPrimarySupplier(tx, int(productID), input.SupplierID, input.LastPrice, input.DefaultLeadTimeDays, input.PackSize); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *ProductRepository) Update(input models.ProductUpdateInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE products
		SET
			product_code = ?,
			barcode = ?,
			product_name = ?,
			category_id = ?,
			unit_id = ?,
			brand = ?,
			min_stock = ?,
			max_stock = ?,
			reorder_point = ?,
			default_lead_time_days = ?,
			pack_size = ?,
			is_active = ?
		WHERE id = ?
	`,
		input.ProductCode,
		nullableString(input.Barcode),
		input.ProductName,
		nullableInt(input.CategoryID),
		nullableInt(input.UnitID),
		nullableString(input.Brand),
		input.MinStock,
		input.MaxStock,
		input.ReorderPoint,
		input.DefaultLeadTimeDays,
		input.PackSize,
		boolToInt(input.IsActive),
		input.ID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := upsertPrimarySupplier(tx, input.ID, input.SupplierID, input.LastPrice, input.DefaultLeadTimeDays, input.PackSize); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *ProductRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM products WHERE id = ?`, id)
	return err
}

func buildProductListQuery(filter models.ProductListFilter, countOnly bool) (string, []interface{}) {
	var query string
	if countOnly {
		query = `
		SELECT COUNT(*)
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN units u ON u.id = p.unit_id
		`
	} else {
		query = `
		SELECT
			p.id,
			p.product_code,
			COALESCE(p.barcode, '') AS barcode,
			p.product_name,
			COALESCE(p.category_id, 0) AS category_id,
			COALESCE(pc.category_name, '') AS category_name,
			COALESCE(p.unit_id, 0) AS unit_id,
			COALESCE(u.unit_name, '') AS unit_name,
			COALESCE(p.brand, '') AS brand,
			p.min_stock,
			p.max_stock,
			p.reorder_point,
			p.default_lead_time_days,
			p.pack_size,
			p.is_active,
			COALESCE((
				SELECT ps.supplier_id
				FROM product_suppliers ps
				WHERE ps.product_id = p.id AND ps.is_active = 1
				ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
				LIMIT 1
			), 0) AS primary_supplier_id,
			COALESCE((
				SELECT s.supplier_name
				FROM product_suppliers ps
				INNER JOIN suppliers s ON s.id = ps.supplier_id
				WHERE ps.product_id = p.id AND ps.is_active = 1
				ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
				LIMIT 1
			), '') AS primary_supplier_name,
			COALESCE((
				SELECT ps.last_price
				FROM product_suppliers ps
				WHERE ps.product_id = p.id AND ps.is_active = 1
				ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
				LIMIT 1
			), 0) AS primary_supplier_price,
			p.created_at,
			p.updated_at
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN units u ON u.id = p.unit_id
		`
	}

	conditions := make([]string, 0, 4)
	args := make([]interface{}, 0, 6)

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		conditions = append(conditions, `(LOWER(p.product_code) LIKE ? OR LOWER(COALESCE(p.barcode, '')) LIKE ? OR LOWER(p.product_name) LIKE ? OR LOWER(COALESCE(p.brand, '')) LIKE ?)`)
		args = append(args, keyword, keyword, keyword, keyword)
	}
	if filter.CategoryID > 0 {
		conditions = append(conditions, "p.category_id = ?")
		args = append(args, filter.CategoryID)
	}
	switch filter.Status {
	case "active":
		conditions = append(conditions, "p.is_active = 1")
	case "inactive":
		conditions = append(conditions, "p.is_active = 0")
	}
	if filter.Brand != "" {
		conditions = append(conditions, "LOWER(COALESCE(p.brand, '')) = ?")
		args = append(args, strings.ToLower(filter.Brand))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

func buildProductOrderClause(filter models.ProductListFilter) string {
	switch filter.Sort {
	case "name":
		return " ORDER BY p.product_name ASC, p.id ASC"
	case "code":
		return " ORDER BY p.product_code ASC, p.id ASC"
	case "brand":
		return " ORDER BY COALESCE(p.brand, '') ASC, p.product_name ASC"
	default:
		return " ORDER BY COALESCE(p.updated_at, p.created_at) DESC, p.id DESC"
	}
}

func scanProduct(scanner interface {
	Scan(dest ...interface{}) error
}) (models.Product, error) {
	var (
		product              models.Product
		minStock             sql.NullFloat64
		maxStock             sql.NullFloat64
		reorderPoint         sql.NullFloat64
		packSize             sql.NullFloat64
		primarySupplierPrice sql.NullFloat64
		isActive             int
		createdAt            sql.NullTime
		updatedAt            sql.NullTime
	)

	err := scanner.Scan(
		&product.ID,
		&product.ProductCode,
		&product.Barcode,
		&product.ProductName,
		&product.CategoryID,
		&product.CategoryName,
		&product.UnitID,
		&product.UnitName,
		&product.Brand,
		&minStock,
		&maxStock,
		&reorderPoint,
		&product.DefaultLeadTimeDays,
		&packSize,
		&isActive,
		&product.PrimarySupplierID,
		&product.PrimarySupplierName,
		&primarySupplierPrice,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return product, err
	}

	if minStock.Valid {
		product.MinStock = minStock.Float64
	}
	if maxStock.Valid {
		product.MaxStock = maxStock.Float64
	}
	if reorderPoint.Valid {
		product.ReorderPoint = reorderPoint.Float64
	}
	if packSize.Valid {
		product.PackSize = packSize.Float64
	}
	if primarySupplierPrice.Valid {
		product.PrimarySupplierPrice = primarySupplierPrice.Float64
	}

	product.MinStockDisplay = formatProductDecimal(product.MinStock)
	product.MaxStockDisplay = formatProductDecimal(product.MaxStock)
	product.ReorderPointDisplay = formatProductDecimal(product.ReorderPoint)
	product.PackSizeDisplay = formatProductDecimal(product.PackSize)
	product.PrimarySupplierPriceDisplay = fmt.Sprintf("Rp %s", formatProductDecimal(product.PrimarySupplierPrice))
	product.IsActive = isActive == 1
	if product.IsActive {
		product.StatusLabel = "Aktif"
	} else {
		product.StatusLabel = "Nonaktif"
	}

	if createdAt.Valid {
		product.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		product.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
	} else {
		product.CreatedAt = "-"
		product.CreatedAtDisplay = "-"
	}
	if updatedAt.Valid {
		product.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		product.UpdatedAtDisplay = updatedAt.Time.Format("02 Jan 2006 15:04")
	} else {
		product.UpdatedAt = "-"
		product.UpdatedAtDisplay = "-"
	}

	return product, nil
}

func upsertPrimarySupplier(tx *sql.Tx, productID, supplierID int, lastPrice float64, leadTimeDays int, packSize float64) error {
	if _, err := tx.Exec(`
		UPDATE product_suppliers
		SET is_primary = 0
		WHERE product_id = ?
	`, productID); err != nil {
		return err
	}

	if supplierID <= 0 {
		return nil
	}

	var existingID int
	err := tx.QueryRow(`
		SELECT id
		FROM product_suppliers
		WHERE product_id = ? AND supplier_id = ?
		LIMIT 1
	`, productID, supplierID).Scan(&existingID)

	if err != nil && err != sql.ErrNoRows {
		return err
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
			) VALUES (?, ?, 1, 1, ?, 0, ?, ?, 1)
		`, productID, supplierID, lastPrice, leadTimeDays, packSize)
		return err
	}

	_, err = tx.Exec(`
		UPDATE product_suppliers
		SET
			is_primary = 1,
			priority_no = 1,
			last_price = ?,
			lead_time_days = ?,
			pack_size = ?,
			is_active = 1
		WHERE id = ?
	`, lastPrice, leadTimeDays, packSize, existingID)
	return err
}

func formatProductDecimal(value float64) string {
	return fmt.Sprintf("%0.2f", value)
}

func formatSignedProductDecimal(value float64) string {
	if value > 0 {
		return "+" + formatProductDecimal(value)
	}
	return formatProductDecimal(value)
}

func formatProductStatusLabel(status string) string {
	if status == "" {
		return "-"
	}
	switch status {
	case "draft":
		return "Draft"
	case "submitted":
		return "Diajukan"
	case "reviewed":
		return "Ditinjau"
	case "approved":
		return "Disetujui"
	case "rejected":
		return "Ditolak"
	case "po_created":
		return "PO Dibuat"
	case "closed":
		return "Selesai"
	case "cancelled":
		return "Dibatalkan"
	default:
		parts := strings.Split(status, "_")
		for i, part := range parts {
			if part == "" {
				continue
			}
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
		return strings.Join(parts, " ")
	}
}

func productStatusBadgeClass(status string) string {
	switch status {
	case "approved", "po_created", "closed":
		return "bg-[#dcfae7] text-[#19a856]"
	case "reviewed", "submitted":
		return "bg-[#fff2c9] text-[#b7791f]"
	case "rejected", "cancelled":
		return "bg-rose-50 text-rose-600"
	default:
		return "bg-slate-100 text-[#566173]"
	}
}
