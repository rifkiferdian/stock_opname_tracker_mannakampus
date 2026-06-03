package repositories

import (
	"database/sql"
	"fmt"
	helpers "gobase-app/helper"
	"gobase-app/models"
	"math"
	"strconv"
	"strings"
	"time"
)

type StockCheckSessionRepository struct {
	DB *sql.DB
}

type StockCheckLatestSubmittedItem struct {
	SessionID        int
	ItemID           int
	SuggestBuyQty    float64
	SuggestBuyCarton int
	SuggestBuyBox    int
	SuggestBuyPcs    int
	ApprovedBuyQty   float64
	BuyerNotes       string
}

func (r *StockCheckSessionRepository) GetCheckerItemRecentSOHistory(currentSessionID int, productID int, storeID int, supplierID int, limit int) ([]models.StockCheckSessionRecentSOHistory, error) {
	if limit <= 0 {
		limit = 4
	}

	rows, err := r.DB.Query(`
		SELECT
			scs.id,
			COALESCE(scs.session_number, '') AS session_number,
			COALESCE(scs.session_date, '') AS session_date,
			COALESCE(si.qty_store_carton, 0) AS qty_store_carton,
			COALESCE(si.qty_store_box, 0) AS qty_store_box,
			COALESCE(si.qty_store_pcs, 0) AS qty_store_pcs,
			COALESCE(si.qty_warehouse_carton, 0) AS qty_warehouse_carton,
			COALESCE(si.qty_warehouse_box, 0) AS qty_warehouse_box,
			COALESCE(si.qty_warehouse_pcs, 0) AS qty_warehouse_pcs,
			COALESCE(si.approved_buy_carton, 0) AS approved_buy_carton,
			COALESCE(si.approved_buy_box, 0) AS approved_buy_box,
			COALESCE(si.approved_buy_pcs, 0) AS approved_buy_pcs
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		WHERE si.product_id = ?
			AND scs.store_id = ?
			AND scs.supplier_id = ?
			AND scs.id <> ?
		ORDER BY scs.session_date DESC, scs.id DESC
		LIMIT ?
	`, productID, storeID, supplierID, currentSessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	histories := make([]models.StockCheckSessionRecentSOHistory, 0)
	for rows.Next() {
		var history models.StockCheckSessionRecentSOHistory
		if err := rows.Scan(
			&history.SessionID,
			&history.SessionNumber,
			&history.SessionDate,
			&history.QtyStoreCarton,
			&history.QtyStoreBox,
			&history.QtyStorePcs,
			&history.QtyWarehouseCarton,
			&history.QtyWarehouseBox,
			&history.QtyWarehousePcs,
			&history.ApprovedBuyCarton,
			&history.ApprovedBuyBox,
			&history.ApprovedBuyPcs,
		); err != nil {
			return nil, err
		}

		history.TotalQtyCarton = history.QtyStoreCarton + history.QtyWarehouseCarton
		history.TotalQtyBox = history.QtyStoreBox + history.QtyWarehouseBox
		history.TotalQtyPcs = history.QtyStorePcs + history.QtyWarehousePcs
		history.GrandTotalCarton = history.QtyStoreCarton + history.QtyWarehouseCarton + history.ApprovedBuyCarton
		history.GrandTotalBox = history.QtyStoreBox + history.QtyWarehouseBox + history.ApprovedBuyBox
		history.GrandTotalPcs = history.QtyStorePcs + history.QtyWarehousePcs + history.ApprovedBuyPcs
		history.QtyStoreBreakdownDisplay = formatStockCheckUnitBreakdownShort(history.QtyStoreCarton, history.QtyStoreBox, history.QtyStorePcs)
		history.QtyWarehouseBreakdownDisplay = formatStockCheckUnitBreakdownShort(history.QtyWarehouseCarton, history.QtyWarehouseBox, history.QtyWarehousePcs)
		history.TotalQtyBreakdownDisplay = formatStockCheckUnitBreakdownShort(history.TotalQtyCarton, history.TotalQtyBox, history.TotalQtyPcs)
		history.ApprovedBuyBreakdownDisplay = formatStockCheckUnitBreakdownShort(history.ApprovedBuyCarton, history.ApprovedBuyBox, history.ApprovedBuyPcs)
		history.GrandTotalBreakdownDisplay = formatStockCheckUnitBreakdownShort(history.GrandTotalCarton, history.GrandTotalBox, history.GrandTotalPcs)
		history.SessionDateDisplay = history.SessionDate
		if parsedDate, err := time.Parse("2006-01-02", history.SessionDate); err == nil {
			history.SessionDateDisplay = parsedDate.Format("02 Jan 2006")
		}

		histories = append(histories, history)
	}

	return histories, rows.Err()
}

func (r *StockCheckSessionRepository) GetAll(filter models.StockCheckSessionListFilter) ([]models.StockCheckSession, error) {
	query, args := buildStockCheckSessionListQuery(filter, false)
	query += " ORDER BY scs.session_date DESC, scs.id DESC LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.StockCheckSession
	for rows.Next() {
		session, err := scanStockCheckSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (r *StockCheckSessionRepository) CountAll(filter models.StockCheckSessionListFilter) (int, error) {
	query, args := buildStockCheckSessionListQuery(filter, true)

	var total int
	err := r.DB.QueryRow(query, args...).Scan(&total)
	return total, err
}

func (r *StockCheckSessionRepository) GetByID(id int) (models.StockCheckSession, error) {
	row := r.DB.QueryRow(`
		SELECT
			scs.id,
			scs.session_number,
			scs.session_date,
			scs.store_id,
			st.store_name,
			scs.supplier_id,
			COALESCE(sp.supplier_code, '') AS supplier_code,
			sp.supplier_name,
			scs.initiation_type,
			scs.status,
			scs.created_by,
			COALESCE(u.name, '') AS created_by_name,
			COALESCE(scs.notes, '') AS notes,
			scs.created_at
		FROM stock_check_sessions scs
		INNER JOIN stores st ON st.store_id = scs.store_id
		INNER JOIN suppliers sp ON sp.id = scs.supplier_id
		LEFT JOIN users u ON u.id = scs.created_by
		WHERE scs.id = ?
		LIMIT 1
	`, id)

	return scanStockCheckSession(row)
}

func (r *StockCheckSessionRepository) GetReviewItems(sessionID int) ([]models.StockCheckSessionReviewItem, error) {
	rows, err := r.DB.Query(`
		SELECT
			si.id,
			si.product_id,
			p.product_code,
			p.product_name,
			COALESCE(p.brand, '') AS brand,
			COALESCE(un.unit_name, '') AS unit_name,
			COALESCE(si.qty_store_carton, 0) AS qty_store_carton,
			COALESCE(si.qty_store_box, 0) AS qty_store_box,
			COALESCE(si.qty_store_pcs, 0) AS qty_store_pcs,
			(
				COALESCE(si.qty_store_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.qty_store_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.qty_store_pcs, 0)
			) AS qty_store,
			COALESCE(si.qty_warehouse_carton, 0) AS qty_warehouse_carton,
			COALESCE(si.qty_warehouse_box, 0) AS qty_warehouse_box,
			COALESCE(si.qty_warehouse_pcs, 0) AS qty_warehouse_pcs,
			(
				COALESCE(si.qty_warehouse_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.qty_warehouse_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.qty_warehouse_pcs, 0)
			) AS qty_warehouse,
			(
				(
					COALESCE(si.qty_store_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
					COALESCE(si.qty_store_box, 0) * COALESCE(p.pcs_per_box, 0) +
					COALESCE(si.qty_store_pcs, 0)
				) +
				(
					COALESCE(si.qty_warehouse_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
					COALESCE(si.qty_warehouse_box, 0) * COALESCE(p.pcs_per_box, 0) +
					COALESCE(si.qty_warehouse_pcs, 0)
				)
			) AS computed_total_qty,
			0 AS system_total_qty,
			(
				COALESCE(si.suggest_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.suggest_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.suggest_buy_pcs, 0)
			) AS suggest_qty,
			COALESCE(si.suggest_buy_carton, 0) AS suggest_buy_carton,
			COALESCE(si.suggest_buy_box, 0) AS suggest_buy_box,
			COALESCE(si.suggest_buy_pcs, 0) AS suggest_buy_pcs,
			COALESCE(si.approved_buy_carton, 0) AS approved_buy_carton,
			COALESCE(si.approved_buy_box, 0) AS approved_buy_box,
			COALESCE(si.approved_buy_pcs, 0) AS approved_buy_pcs,
			(
				COALESCE(si.approved_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.approved_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.approved_buy_pcs, 0)
			) AS approved_qty,
			COALESCE(sel.supplier_name, sessup.supplier_name, '') AS selected_supplier_name,
			COALESCE(si.checker_notes, '') AS checker_notes,
			COALESCE(si.buyer_notes, '') AS buyer_notes,
			si.condition_status,
			si.status,
			COALESCE(
				(
					SELECT ps.last_price
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
						AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				(
					SELECT ps.last_price
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				0
			) AS unit_price
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		INNER JOIN products p ON p.id = si.product_id
		LEFT JOIN units un ON un.id = p.unit_id
		LEFT JOIN suppliers sessup ON sessup.id = scs.supplier_id
		LEFT JOIN suppliers sel ON sel.id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
		WHERE si.stock_check_session_id = ?
		ORDER BY
			LOWER(COALESCE(p.product_name, '')) ASC,
			p.product_name ASC,
			si.id ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.StockCheckSessionReviewItem
	for rows.Next() {
		var (
			item           models.StockCheckSessionReviewItem
			qtyStore       sql.NullFloat64
			qtyWarehouse   sql.NullFloat64
			totalQty       sql.NullFloat64
			systemTotalQty sql.NullFloat64
			suggestQty     sql.NullFloat64
			approvedQty    sql.NullFloat64
			unitPrice      sql.NullFloat64
		)

		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.ProductCode,
			&item.ProductName,
			&item.Brand,
			&item.UnitName,
			&item.QtyStoreCarton,
			&item.QtyStoreBox,
			&item.QtyStorePcs,
			&qtyStore,
			&item.QtyWarehouseCarton,
			&item.QtyWarehouseBox,
			&item.QtyWarehousePcs,
			&qtyWarehouse,
			&totalQty,
			&systemTotalQty,
			&suggestQty,
			&item.SuggestBuyCarton,
			&item.SuggestBuyBox,
			&item.SuggestBuyPcs,
			&item.ApprovedBuyCarton,
			&item.ApprovedBuyBox,
			&item.ApprovedBuyPcs,
			&approvedQty,
			&item.SelectedSupplierName,
			&item.CheckerNotes,
			&item.BuyerNotes,
			&item.ConditionStatus,
			&item.Status,
			&unitPrice,
		); err != nil {
			return nil, err
		}

		if qtyStore.Valid {
			item.QtyStore = qtyStore.Float64
		}
		if qtyWarehouse.Valid {
			item.QtyWarehouse = qtyWarehouse.Float64
		}
		if totalQty.Valid {
			item.TotalQty = totalQty.Float64
		}
		if systemTotalQty.Valid {
			item.SystemTotalQty = systemTotalQty.Float64
		}
		if suggestQty.Valid {
			item.SuggestBuyQty = suggestQty.Float64
		}
		if approvedQty.Valid {
			item.ApprovedBuyQty = approvedQty.Float64
		}

		linePrice := 0.0
		if unitPrice.Valid {
			linePrice = unitPrice.Float64
		}

		if strings.TrimSpace(item.UnitName) == "" {
			item.UnitName = "unit"
		}
		if strings.TrimSpace(item.SelectedSupplierName) == "" {
			item.SelectedSupplierName = "-"
		}

		item.ProductInitials = strings.ToUpper(helpers.Initials(item.ProductName))
		if item.ProductInitials == "" {
			item.ProductInitials = "PR"
		}
		item.ProductAvatarClass = stockCheckSessionProductAvatarClass(item.ProductName)
		item.QtyStoreDisplay = formatStockCheckWholeNumber(item.QtyStore)
		item.QtyStoreBreakdownDisplay = formatStockCheckUnitBreakdownShort(item.QtyStoreCarton, item.QtyStoreBox, item.QtyStorePcs)
		item.QtyWarehouseDisplay = formatStockCheckWholeNumber(item.QtyWarehouse)
		item.QtyWarehouseBreakdownDisplay = formatStockCheckUnitBreakdownShort(item.QtyWarehouseCarton, item.QtyWarehouseBox, item.QtyWarehousePcs)
		item.TotalQtyCarton = item.QtyStoreCarton + item.QtyWarehouseCarton
		item.TotalQtyBox = item.QtyStoreBox + item.QtyWarehouseBox
		item.TotalQtyPcs = item.QtyStorePcs + item.QtyWarehousePcs
		item.TotalQtyDisplay = formatStockCheckWholeNumber(item.TotalQty)
		item.TotalQtyBreakdownDisplay = formatStockCheckUnitBreakdownShort(item.TotalQtyCarton, item.TotalQtyBox, item.TotalQtyPcs)
		item.SystemTotalQtyDisplay = formatStockCheckWholeNumber(item.SystemTotalQty)
		item.SuggestBuyQtyDisplay = formatStockCheckWholeNumber(item.SuggestBuyQty)
		item.SuggestBuyBreakdownDisplay = formatStockCheckUnitBreakdownShort(item.SuggestBuyCarton, item.SuggestBuyBox, item.SuggestBuyPcs)
		item.ApprovedBuyBreakdownDisplay = formatStockCheckUnitBreakdownShort(item.ApprovedBuyCarton, item.ApprovedBuyBox, item.ApprovedBuyPcs)
		item.ApprovedBuyQtyDisplay = formatStockCheckWholeNumber(item.ApprovedBuyQty)
		item.SuggestLineValue = item.SuggestBuyQty * linePrice
		item.ApprovedLineValue = item.ApprovedBuyQty * linePrice
		item.SuggestLineValueDisplay = formatStockCheckCurrency(item.SuggestLineValue)
		item.ApprovedLineValueDisplay = formatStockCheckCurrency(item.ApprovedLineValue)
		item.ConditionLabel, item.ConditionBadgeClass, item.BuyerNoteAccentClass = stockCheckSessionConditionMeta(item.ConditionStatus)
		item.StatusLabel, item.StatusBadgeClass = stockCheckSessionItemStatusMeta(item.Status)

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StockCheckSessionRepository) GetStoreOptions() ([]models.Store, error) {
	rows, err := r.DB.Query(`
		SELECT store_id, store_name
		FROM stores
		WHERE is_active = 1
		ORDER BY store_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var store models.Store
		if err := rows.Scan(&store.StoreID, &store.StoreName); err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}

	return stores, rows.Err()
}

func (r *StockCheckSessionRepository) GetStoreOptionsByUserID(userID int) ([]models.Store, error) {
	rows, err := r.DB.Query(`
		SELECT s.store_id, s.store_name
		FROM stores s
		INNER JOIN user_stores us ON us.store_id = s.store_id
		WHERE s.is_active = 1 AND us.user_id = ?
		ORDER BY s.store_name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var store models.Store
		if err := rows.Scan(&store.StoreID, &store.StoreName); err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}

	return stores, rows.Err()
}

func (r *StockCheckSessionRepository) UserHasStoreAccess(userID int, storeID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM user_stores us
		INNER JOIN stores s ON s.store_id = us.store_id
		WHERE us.user_id = ? AND us.store_id = ? AND s.is_active = 1
	`, userID, storeID).Scan(&count)
	return count > 0, err
}

func (r *StockCheckSessionRepository) UserHasRole(userID int, roleName string) (bool, error) {
	roleName = strings.TrimSpace(roleName)
	if userID <= 0 || roleName == "" {
		return false, nil
	}

	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM model_has_roles mhr
		INNER JOIN roles r ON r.id = mhr.role_id
		WHERE mhr.model_id = ? AND LOWER(r.name) = LOWER(?)
	`, userID, roleName).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *StockCheckSessionRepository) GetSupplierOptions() ([]models.Supplier, error) {
	rows, err := r.DB.Query(`
		SELECT
			id,
			supplier_name,
			COALESCE(store_id, 0) AS store_id
		FROM suppliers
		WHERE is_active = 1
		ORDER BY supplier_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		if err := rows.Scan(&supplier.ID, &supplier.SupplierName, &supplier.StoreID); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, rows.Err()
}

func (r *StockCheckSessionRepository) GetStoreCodeByID(storeID int) (string, error) {
	var storeCode string
	err := r.DB.QueryRow(`SELECT store_code FROM stores WHERE store_id = ?`, storeID).Scan(&storeCode)
	return strings.ToUpper(strings.TrimSpace(storeCode)), err
}

func (r *StockCheckSessionRepository) GetNextSessionSequence(prefix string) (int, error) {
	var currentMax sql.NullInt64
	err := r.DB.QueryRow(`
		SELECT MAX(CAST(RIGHT(session_number, 3) AS UNSIGNED))
		FROM stock_check_sessions
		WHERE session_number LIKE ?
	`, prefix+"%").Scan(&currentMax)
	if err != nil {
		return 0, err
	}
	if !currentMax.Valid {
		return 1, nil
	}
	return int(currentMax.Int64) + 1, nil
}

func (r *StockCheckSessionRepository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM stock_check_sessions WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

func (r *StockCheckSessionRepository) GetLatestSessionIDBySupplier(supplierID int) (int, error) {
	if supplierID <= 0 {
		return 0, sql.ErrNoRows
	}

	var sessionID int
	err := r.DB.QueryRow(`
		SELECT scs.id
		FROM stock_check_sessions scs
		WHERE scs.supplier_id = ?
		ORDER BY scs.session_date DESC, scs.id DESC
		LIMIT 1
	`, supplierID).Scan(&sessionID)
	if err != nil {
		return 0, err
	}

	return sessionID, nil
}

func (r *StockCheckSessionRepository) ExistsBySessionNumber(sessionNumber string, ignoreID int) (bool, error) {
	var (
		count int
		err   error
	)

	if ignoreID > 0 {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM stock_check_sessions WHERE session_number = ? AND id <> ?`, sessionNumber, ignoreID).Scan(&count)
	} else {
		err = r.DB.QueryRow(`SELECT COUNT(1) FROM stock_check_sessions WHERE session_number = ?`, sessionNumber).Scan(&count)
	}

	return count > 0, err
}

func (r *StockCheckSessionRepository) ExistsReviewItem(sessionID int, itemID int) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM stock_check_session_items
		WHERE stock_check_session_id = ? AND id = ?
	`, sessionID, itemID).Scan(&count)

	return count > 0, err
}

func (r *StockCheckSessionRepository) Create(input models.StockCheckSessionCreateInput) (int, error) {
	res, err := r.DB.Exec(`
		INSERT INTO stock_check_sessions (
			session_number,
			session_date,
			store_id,
			supplier_id,
			initiation_type,
			status,
			created_by,
			notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		input.SessionNumber,
		input.SessionDate,
		input.StoreID,
		input.SupplierID,
		input.InitiationType,
		input.Status,
		input.CreatedBy,
		nullableText(input.Notes),
	)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastInsertID), nil
}

func (r *StockCheckSessionRepository) Update(input models.StockCheckSessionUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE stock_check_sessions
		SET
			session_date = ?,
			store_id = ?,
			supplier_id = ?,
			initiation_type = ?,
			status = ?,
			notes = ?
		WHERE id = ?
	`,
		input.SessionDate,
		input.StoreID,
		input.SupplierID,
		input.InitiationType,
		input.Status,
		nullableText(input.Notes),
		input.ID,
	)

	return err
}

func (r *StockCheckSessionRepository) UpdateStatus(sessionID int, status string) error {
	_, err := r.DB.Exec(`
		UPDATE stock_check_sessions
		SET status = ?
		WHERE id = ?
	`, status, sessionID)

	return err
}

func (r *StockCheckSessionRepository) UpdateReviewItem(input models.StockCheckSessionReviewItemUpdateInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var existing struct {
		ProductID         int
		PcsPerBox         int
		PcsPerCarton      int
		ApprovedBuyCarton int
		ApprovedBuyBox    int
		ApprovedBuyPcs    int
		ApprovedBuyQty    sql.NullFloat64
		BuyerNotes        sql.NullString
		Status            string
	}

	err = tx.QueryRow(`
		SELECT
			si.product_id,
			COALESCE(p.pcs_per_box, 0) AS pcs_per_box,
			COALESCE(p.pcs_per_carton, 0) AS pcs_per_carton,
			COALESCE(si.approved_buy_carton, 0) AS approved_buy_carton,
			COALESCE(si.approved_buy_box, 0) AS approved_buy_box,
			COALESCE(si.approved_buy_pcs, 0) AS approved_buy_pcs,
			(
				COALESCE(si.approved_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.approved_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.approved_buy_pcs, 0)
			) AS approved_qty,
			buyer_notes,
			status
		FROM stock_check_session_items si
		INNER JOIN products p ON p.id = si.product_id
		WHERE si.stock_check_session_id = ? AND si.id = ?
		LIMIT 1
	`,
		input.SessionID,
		input.ItemID,
	).Scan(
		&existing.ProductID,
		&existing.PcsPerBox,
		&existing.PcsPerCarton,
		&existing.ApprovedBuyCarton,
		&existing.ApprovedBuyBox,
		&existing.ApprovedBuyPcs,
		&existing.ApprovedBuyQty,
		&existing.BuyerNotes,
		&existing.Status,
	)
	if err != nil {
		return err
	}

	input.ApprovedBuyQty = float64(input.ApprovedBuyCarton*existing.PcsPerCarton + input.ApprovedBuyBox*existing.PcsPerBox + input.ApprovedBuyPcs)

	_, err = tx.Exec(`
		UPDATE stock_check_session_items
		SET
			approved_buy_carton = ?,
			approved_buy_box = ?,
			approved_buy_pcs = ?,
			buyer_notes = ?,
			status = ?,
			reviewed_by = ?,
			reviewed_at = NOW(),
			updated_by = ?
		WHERE stock_check_session_id = ? AND id = ?
	`,
		input.ApprovedBuyCarton,
		input.ApprovedBuyBox,
		input.ApprovedBuyPcs,
		nullableText(input.BuyerNotes),
		input.Status,
		input.ReviewedBy,
		input.UpdatedBy,
		input.SessionID,
		input.ItemID,
	)
	if err != nil {
		return err
	}

	histories := []struct {
		FieldName    string
		OldValue     string
		NewValue     string
		ChangeReason string
		Notes        string
	}{
		{
			FieldName:    "approved_buy_carton",
			OldValue:     strconv.Itoa(existing.ApprovedBuyCarton),
			NewValue:     strconv.Itoa(input.ApprovedBuyCarton),
			ChangeReason: "review item updated",
			Notes:        "Perubahan final approve carton dari halaman detail stock check session.",
		},
		{
			FieldName:    "approved_buy_box",
			OldValue:     strconv.Itoa(existing.ApprovedBuyBox),
			NewValue:     strconv.Itoa(input.ApprovedBuyBox),
			ChangeReason: "review item updated",
			Notes:        "Perubahan final approve box dari halaman detail stock check session.",
		},
		{
			FieldName:    "approved_buy_pcs",
			OldValue:     strconv.Itoa(existing.ApprovedBuyPcs),
			NewValue:     strconv.Itoa(input.ApprovedBuyPcs),
			ChangeReason: "review item updated",
			Notes:        "Perubahan final approve pcs dari halaman detail stock check session.",
		},
		{
			FieldName:    "buyer_notes",
			OldValue:     formatHistoryText(existing.BuyerNotes),
			NewValue:     strings.TrimSpace(input.BuyerNotes),
			ChangeReason: "review item updated",
			Notes:        "Perubahan buyer notes dari halaman detail stock check session.",
		},
		{
			FieldName:    "status",
			OldValue:     strings.TrimSpace(existing.Status),
			NewValue:     strings.TrimSpace(input.Status),
			ChangeReason: "review item updated",
			Notes:        "Perubahan status approval dari halaman detail stock check session.",
		},
	}

	for _, history := range histories {
		if history.OldValue == history.NewValue {
			continue
		}

		_, err = tx.Exec(`
			INSERT INTO stock_check_session_item_histories (
				stock_check_session_item_id,
				product_id,
				field_name,
				old_value,
				new_value,
				change_reason,
				notes,
				changed_by
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			input.ItemID,
			existing.ProductID,
			history.FieldName,
			nullableHistoryValue(history.OldValue),
			nullableHistoryValue(history.NewValue),
			nullableText(history.ChangeReason),
			nullableText(history.Notes),
			input.ReviewedBy,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (r *StockCheckSessionRepository) GetLatestSubmittedItemsBySupplier(supplierID int) ([]StockCheckLatestSubmittedItem, error) {
	rows, err := r.DB.Query(`
		SELECT
			si.stock_check_session_id,
			si.id,
			(
				COALESCE(si.suggest_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.suggest_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.suggest_buy_pcs, 0)
			) AS suggest_buy_qty,
			COALESCE(si.suggest_buy_carton, 0) AS suggest_buy_carton,
			COALESCE(si.suggest_buy_box, 0) AS suggest_buy_box,
			COALESCE(si.suggest_buy_pcs, 0) AS suggest_buy_pcs,
			(
				COALESCE(si.approved_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.approved_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.approved_buy_pcs, 0)
			) AS approved_qty,
			COALESCE(si.buyer_notes, '') AS buyer_notes
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		INNER JOIN products p ON p.id = si.product_id
		WHERE scs.supplier_id = ?
			AND scs.session_date = (
				SELECT MAX(scs_latest.session_date)
				FROM stock_check_sessions scs_latest
				WHERE scs_latest.supplier_id = ?
			)
			AND si.status = 'submitted'
		ORDER BY scs.id DESC, si.id ASC
	`, supplierID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StockCheckLatestSubmittedItem, 0)
	for rows.Next() {
		var (
			item             StockCheckLatestSubmittedItem
			suggestBuyQty    sql.NullFloat64
			suggestBuyCarton sql.NullInt64
			suggestBuyBox    sql.NullInt64
			suggestBuyPcs    sql.NullInt64
			approvedBuyQty   sql.NullFloat64
		)

		if err := rows.Scan(
			&item.SessionID,
			&item.ItemID,
			&suggestBuyQty,
			&suggestBuyCarton,
			&suggestBuyBox,
			&suggestBuyPcs,
			&approvedBuyQty,
			&item.BuyerNotes,
		); err != nil {
			return nil, err
		}

		if suggestBuyQty.Valid {
			item.SuggestBuyQty = suggestBuyQty.Float64
		}
		if suggestBuyCarton.Valid {
			item.SuggestBuyCarton = int(suggestBuyCarton.Int64)
		}
		if suggestBuyBox.Valid {
			item.SuggestBuyBox = int(suggestBuyBox.Int64)
		}
		if suggestBuyPcs.Valid {
			item.SuggestBuyPcs = int(suggestBuyPcs.Int64)
		}
		if approvedBuyQty.Valid {
			item.ApprovedBuyQty = approvedBuyQty.Float64
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StockCheckSessionRepository) GetSubmittedItemsBySession(sessionID int) ([]StockCheckLatestSubmittedItem, error) {
	rows, err := r.DB.Query(`
		SELECT
			si.stock_check_session_id,
			si.id,
			(
				COALESCE(si.suggest_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.suggest_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.suggest_buy_pcs, 0)
			) AS suggest_buy_qty,
			COALESCE(si.suggest_buy_carton, 0) AS suggest_buy_carton,
			COALESCE(si.suggest_buy_box, 0) AS suggest_buy_box,
			COALESCE(si.suggest_buy_pcs, 0) AS suggest_buy_pcs,
			(
				COALESCE(si.approved_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.approved_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.approved_buy_pcs, 0)
			) AS approved_qty,
			COALESCE(si.buyer_notes, '') AS buyer_notes
		FROM stock_check_session_items si
		INNER JOIN products p ON p.id = si.product_id
		WHERE si.stock_check_session_id = ?
			AND si.status = 'submitted'
		ORDER BY si.id ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StockCheckLatestSubmittedItem, 0)
	for rows.Next() {
		var (
			item             StockCheckLatestSubmittedItem
			suggestBuyQty    sql.NullFloat64
			suggestBuyCarton sql.NullInt64
			suggestBuyBox    sql.NullInt64
			suggestBuyPcs    sql.NullInt64
			approvedBuyQty   sql.NullFloat64
		)

		if err := rows.Scan(
			&item.SessionID,
			&item.ItemID,
			&suggestBuyQty,
			&suggestBuyCarton,
			&suggestBuyBox,
			&suggestBuyPcs,
			&approvedBuyQty,
			&item.BuyerNotes,
		); err != nil {
			return nil, err
		}

		if suggestBuyQty.Valid {
			item.SuggestBuyQty = suggestBuyQty.Float64
		}
		if suggestBuyCarton.Valid {
			item.SuggestBuyCarton = int(suggestBuyCarton.Int64)
		}
		if suggestBuyBox.Valid {
			item.SuggestBuyBox = int(suggestBuyBox.Int64)
		}
		if suggestBuyPcs.Valid {
			item.SuggestBuyPcs = int(suggestBuyPcs.Int64)
		}
		if approvedBuyQty.Valid {
			item.ApprovedBuyQty = approvedBuyQty.Float64
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StockCheckSessionRepository) GetLatestBuyerApproverName(sessionID int) (string, error) {
	var approverName sql.NullString
	err := r.DB.QueryRow(`
		SELECT COALESCE(u.name, '')
		FROM stock_check_session_items si
		INNER JOIN users u ON u.id = si.reviewed_by
		WHERE si.stock_check_session_id = ?
			AND si.reviewed_by IS NOT NULL
			AND si.reviewed_at IS NOT NULL
		ORDER BY si.reviewed_at DESC, si.id DESC
		LIMIT 1
	`, sessionID).Scan(&approverName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(approverName.String), nil
}

func (r *StockCheckSessionRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM stock_check_sessions WHERE id = ?`, id)
	return err
}

func (r *StockCheckSessionRepository) SeedItemsFromSupplier(sessionID int, supplierID int, storeID int, userID int) error {
	_, err := r.DB.Exec(`
		INSERT INTO stock_check_session_items (
			stock_check_session_id,
			product_id,
			suggested_supplier_id,
			status,
			created_by,
			updated_by
		)
		SELECT
			?,
			ps.product_id,
			?,
			'draft',
			?,
			?
		FROM product_suppliers ps
		INNER JOIN products p ON p.id = ps.product_id
		WHERE ps.supplier_id = ?
			AND ps.is_active = 1
			AND COALESCE(p.store_id, 0) = ?
			AND NOT EXISTS (
				SELECT 1
				FROM stock_check_session_items si
				WHERE si.stock_check_session_id = ?
					AND si.product_id = ps.product_id
			)
		GROUP BY ps.product_id
	`,
		sessionID,
		supplierID,
		userID,
		userID,
		supplierID,
		storeID,
		sessionID,
	)

	return err
}

func (r *StockCheckSessionRepository) GetCheckerInputItems(sessionID int, storeID int, supplierID int) ([]models.StockCheckSessionCheckerInputItem, error) {
	rows, err := r.DB.Query(`
		SELECT
			si.id,
			si.product_id,
			p.product_code,
			COALESCE(p.barcode, '') AS barcode,
			COALESCE(p.barcode_box, '') AS barcode_box,
			COALESCE(p.barcode_carton, '') AS barcode_carton,
			p.product_name,
			COALESCE(pc.category_name, 'Tanpa Kategori') AS category_name,
			COALESCE(un.unit_name, '-') AS unit_name,
			COALESCE(p.pcs_per_box, 0) AS pcs_per_box,
			COALESCE(p.box_per_carton, 0) AS box_per_carton,
			COALESCE(p.pcs_per_carton, 0) AS pcs_per_carton,
			COALESCE(si.qty_store_carton, 0) AS qty_store_carton,
			COALESCE(si.qty_store_box, 0) AS qty_store_box,
			COALESCE(si.qty_store_pcs, 0) AS qty_store_pcs,
			(
				COALESCE(si.qty_store_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.qty_store_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.qty_store_pcs, 0)
			) AS qty_store,
			COALESCE(si.qty_warehouse_carton, 0) AS qty_warehouse_carton,
			COALESCE(si.qty_warehouse_box, 0) AS qty_warehouse_box,
			COALESCE(si.qty_warehouse_pcs, 0) AS qty_warehouse_pcs,
			(
				COALESCE(si.qty_warehouse_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.qty_warehouse_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.qty_warehouse_pcs, 0)
			) AS qty_warehouse,
			(
				(
					COALESCE(si.qty_store_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
					COALESCE(si.qty_store_box, 0) * COALESCE(p.pcs_per_box, 0) +
					COALESCE(si.qty_store_pcs, 0)
				) +
				(
					COALESCE(si.qty_warehouse_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
					COALESCE(si.qty_warehouse_box, 0) * COALESCE(p.pcs_per_box, 0) +
					COALESCE(si.qty_warehouse_pcs, 0)
				)
			) AS computed_total_qty,
			COALESCE(si.suggest_buy_carton, 0) AS suggest_buy_carton,
			COALESCE(si.suggest_buy_box, 0) AS suggest_buy_box,
			COALESCE(si.suggest_buy_pcs, 0) AS suggest_buy_pcs,
			(
				COALESCE(si.suggest_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.suggest_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.suggest_buy_pcs, 0)
			) AS suggest_qty,
			CASE WHEN si.store_checked_at IS NULL THEN 0 ELSE 1 END AS store_checked,
			CASE WHEN si.warehouse_checked_at IS NULL THEN 0 ELSE 1 END AS warehouse_checked,
			COALESCE(si.status, 'draft') AS status
		FROM stock_check_session_items si
		INNER JOIN products p ON p.id = si.product_id
		INNER JOIN product_suppliers ps ON ps.product_id = p.id
			AND ps.supplier_id = ?
			AND ps.is_active = 1
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN units un ON un.id = p.unit_id
		WHERE si.stock_check_session_id = ?
			AND COALESCE(p.store_id, 0) = ?
		ORDER BY
			CASE
				WHEN COALESCE(si.updated_at, si.created_at) > COALESCE(si.created_at, si.updated_at) THEN 0
				ELSE 1
			END ASC,
			COALESCE(si.updated_at, si.created_at) DESC,
			p.product_name ASC,
			si.id ASC
	`, supplierID, sessionID, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.StockCheckSessionCheckerInputItem, 0)
	for rows.Next() {
		var (
			item             models.StockCheckSessionCheckerInputItem
			qtyStore         sql.NullFloat64
			qtyWarehouse     sql.NullFloat64
			totalQty         sql.NullFloat64
			suggestQty       sql.NullFloat64
			storeChecked     int
			warehouseChecked int
		)

		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.ProductCode,
			&item.Barcode,
			&item.BarcodeBox,
			&item.BarcodeCarton,
			&item.ProductName,
			&item.CategoryName,
			&item.UnitName,
			&item.PcsPerBox,
			&item.BoxPerCarton,
			&item.PcsPerCarton,
			&item.QtyStoreCarton,
			&item.QtyStoreBox,
			&item.QtyStorePcs,
			&qtyStore,
			&item.QtyWarehouseCarton,
			&item.QtyWarehouseBox,
			&item.QtyWarehousePcs,
			&qtyWarehouse,
			&totalQty,
			&item.SuggestBuyCarton,
			&item.SuggestBuyBox,
			&item.SuggestBuyPcs,
			&suggestQty,
			&storeChecked,
			&warehouseChecked,
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
		if totalQty.Valid {
			item.TotalQty = totalQty.Float64
		}
		if suggestQty.Valid {
			item.SuggestBuyQty = suggestQty.Float64
		}

		item.QtyStoreDisplay = formatStockCheckWholeNumber(item.QtyStore)
		item.QtyWarehouseDisplay = formatStockCheckWholeNumber(item.QtyWarehouse)
		item.TotalQtyDisplay = formatStockCheckWholeNumber(item.TotalQty)
		item.SuggestBuyQtyDisplay = formatStockCheckWholeNumber(item.SuggestBuyQty)
		item.SuggestBuyCarton, item.SuggestBuyBox, item.SuggestBuyPcs, item.SuggestBuyCartonDisplay, item.SuggestBuyBreakdownDisplay = resolveStockCheckSuggestBreakdown(
			item.SuggestBuyQty,
			item.SuggestBuyCarton,
			item.SuggestBuyBox,
			item.SuggestBuyPcs,
			item.PcsPerBox,
			item.PcsPerCarton,
		)
		item.QtyStoreBreakdownDisplay = formatStockCheckUnitBreakdown(item.QtyStoreCarton, item.QtyStoreBox, item.QtyStorePcs)
		item.QtyWarehouseBreakdownDisplay = formatStockCheckUnitBreakdown(item.QtyWarehouseCarton, item.QtyWarehouseBox, item.QtyWarehousePcs)
		item.ConversionDisplay = formatStockCheckConversion(item.PcsPerBox, item.BoxPerCarton, item.PcsPerCarton)
		item.StatusLabel, _ = stockCheckSessionItemStatusMeta(item.Status)
		item.PreferredBarcode = pickPreferredProductBarcode(item.Barcode, item.BarcodeBox, item.BarcodeCarton)
		item.BarcodeSummary = buildProductBarcodeSummary(item.Barcode, item.BarcodeBox, item.BarcodeCarton)
		item.BarcodeSearchText = strings.TrimSpace(strings.Join([]string{item.Barcode, item.BarcodeBox, item.BarcodeCarton}, " "))
		item.HasBarcode = item.PreferredBarcode != ""
		item.StoreChecked = storeChecked > 0
		item.WarehouseChecked = warehouseChecked > 0

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StockCheckSessionRepository) UpdateCheckerItemQtyByBarcode(sessionID int, storeID int, supplierID int, location string, barcode string, qtyCarton int, qtyBox int, qtyPcs int, updatedBy int) (int, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		itemID          int
		pcsPerBox       int
		pcsPerCarton    int
		storeCarton     int
		storeBox        int
		storePcs        int
		warehouseCarton int
		warehouseBox    int
		warehousePcs    int
	)

	err = tx.QueryRow(`
		SELECT
			si.id,
			COALESCE(p.pcs_per_box, 0) AS pcs_per_box,
			COALESCE(p.pcs_per_carton, 0) AS pcs_per_carton,
			COALESCE(si.qty_store_carton, 0) AS qty_store_carton,
			COALESCE(si.qty_store_box, 0) AS qty_store_box,
			COALESCE(si.qty_store_pcs, 0) AS qty_store_pcs,
			COALESCE(si.qty_warehouse_carton, 0) AS qty_warehouse_carton,
			COALESCE(si.qty_warehouse_box, 0) AS qty_warehouse_box,
			COALESCE(si.qty_warehouse_pcs, 0) AS qty_warehouse_pcs
		FROM stock_check_session_items si
		INNER JOIN products p ON p.id = si.product_id
		INNER JOIN product_suppliers ps ON ps.product_id = p.id
			AND ps.supplier_id = ?
			AND ps.is_active = 1
		WHERE si.stock_check_session_id = ?
			AND COALESCE(p.store_id, 0) = ?
			AND (
				LOWER(COALESCE(p.barcode, '')) = LOWER(?)
				OR LOWER(COALESCE(p.barcode_box, '')) = LOWER(?)
				OR LOWER(COALESCE(p.barcode_carton, '')) = LOWER(?)
			)
		LIMIT 1
	`, supplierID, sessionID, storeID, strings.TrimSpace(barcode), strings.TrimSpace(barcode), strings.TrimSpace(barcode)).Scan(
		&itemID,
		&pcsPerBox,
		&pcsPerCarton,
		&storeCarton,
		&storeBox,
		&storePcs,
		&warehouseCarton,
		&warehouseBox,
		&warehousePcs,
	)
	if err != nil {
		return 0, err
	}

	switch location {
	case "warehouse":
		warehouseCarton = qtyCarton
		warehouseBox = qtyBox
		warehousePcs = qtyPcs
	default:
		storeCarton = qtyCarton
		storeBox = qtyBox
		storePcs = qtyPcs
	}

	_, err = tx.Exec(`
		UPDATE stock_check_session_items
		SET
			qty_store_carton = ?,
			qty_store_box = ?,
			qty_store_pcs = ?,
			qty_warehouse_carton = ?,
			qty_warehouse_box = ?,
			qty_warehouse_pcs = ?,
			store_checked_at = CASE WHEN ? = 'warehouse' THEN store_checked_at ELSE NOW() END,
			store_checked_by = CASE WHEN ? = 'warehouse' THEN store_checked_by ELSE ? END,
			warehouse_checked_at = CASE WHEN ? = 'warehouse' THEN NOW() ELSE warehouse_checked_at END,
			warehouse_checked_by = CASE WHEN ? = 'warehouse' THEN ? ELSE warehouse_checked_by END,
			updated_by = ?
		WHERE stock_check_session_id = ? AND id = ?
	`,
		storeCarton,
		storeBox,
		storePcs,
		warehouseCarton,
		warehouseBox,
		warehousePcs,
		location,
		location,
		updatedBy,
		location,
		location,
		updatedBy,
		updatedBy,
		sessionID,
		itemID,
	)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return itemID, nil
}

func (r *StockCheckSessionRepository) UpdateCheckerItemSuggest(sessionID int, itemID int, suggestCarton int, suggestBox int, suggestPcs int, updatedBy int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		UPDATE stock_check_session_items
		SET
			suggest_buy_carton = ?,
			suggest_buy_box = ?,
			suggest_buy_pcs = ?,
			status = CASE
				WHEN status = 'draft' THEN 'submitted'
				ELSE status
			END,
			updated_by = ?
		WHERE stock_check_session_id = ? AND id = ?
	`, suggestCarton, suggestBox, suggestPcs, updatedBy, sessionID, itemID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func buildStockCheckSessionListQuery(filter models.StockCheckSessionListFilter, countOnly bool) (string, []interface{}) {
	var query string
	if countOnly {
		query = `
		SELECT COUNT(*)
		FROM stock_check_sessions scs
		INNER JOIN stores st ON st.store_id = scs.store_id
		INNER JOIN suppliers sp ON sp.id = scs.supplier_id
		LEFT JOIN users u ON u.id = scs.created_by
		`
	} else {
		query = `
		SELECT
			scs.id,
			scs.session_number,
			scs.session_date,
			scs.store_id,
			st.store_name,
			scs.supplier_id,
			COALESCE(sp.supplier_code, '') AS supplier_code,
			sp.supplier_name,
			scs.initiation_type,
			scs.status,
			scs.created_by,
			COALESCE(u.name, '') AS created_by_name,
			COALESCE(scs.notes, '') AS notes,
			scs.created_at
		FROM stock_check_sessions scs
		INNER JOIN stores st ON st.store_id = scs.store_id
		INNER JOIN suppliers sp ON sp.id = scs.supplier_id
		LEFT JOIN users u ON u.id = scs.created_by
		`
	}

	conditions := make([]string, 0, 4)
	args := make([]interface{}, 0, 6)

	if filter.DateFrom != "" {
		conditions = append(conditions, "scs.session_date >= ?")
		args = append(args, filter.DateFrom)
	}
	if filter.DateTo != "" {
		conditions = append(conditions, "scs.session_date <= ?")
		args = append(args, filter.DateTo)
	}
	if filter.StoreID > 0 {
		conditions = append(conditions, "scs.store_id = ?")
		args = append(args, filter.StoreID)
	}
	if filter.SupplierID > 0 {
		conditions = append(conditions, "scs.supplier_id = ?")
		args = append(args, filter.SupplierID)
	}
	if strings.TrimSpace(filter.SupplierName) != "" {
		conditions = append(conditions, "LOWER(sp.supplier_name) LIKE LOWER(?)")
		args = append(args, "%"+strings.TrimSpace(filter.SupplierName)+"%")
	}
	if filter.Status != "" {
		conditions = append(conditions, "scs.status = ?")
		args = append(args, filter.Status)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

func scanStockCheckSession(scanner interface {
	Scan(dest ...interface{}) error
}) (models.StockCheckSession, error) {
	var (
		session     models.StockCheckSession
		sessionDate sql.NullTime
		createdAt   sql.NullTime
	)

	err := scanner.Scan(
		&session.ID,
		&session.SessionNumber,
		&sessionDate,
		&session.StoreID,
		&session.StoreName,
		&session.SupplierID,
		&session.SupplierCode,
		&session.SupplierName,
		&session.InitiationType,
		&session.Status,
		&session.CreatedBy,
		&session.CreatedByName,
		&session.Notes,
		&createdAt,
	)
	if err != nil {
		return session, err
	}

	if sessionDate.Valid {
		session.SessionDate = sessionDate.Time.Format("2006-01-02")
		session.SessionDateDisplay = sessionDate.Time.Format("02 Jan 2006")
		session.SessionDateMonthShort = sessionDate.Time.Format("Jan")
		session.SessionDateDay = sessionDate.Time.Format("02")
		session.SessionDateYear = sessionDate.Time.Format("2006")
	} else {
		session.SessionDate = "-"
		session.SessionDateDisplay = "-"
		session.SessionDateMonthShort = "-"
		session.SessionDateDay = "-"
		session.SessionDateYear = "-"
	}

	if createdAt.Valid {
		session.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		session.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04")
	} else {
		session.CreatedAt = "-"
		session.CreatedAtDisplay = "-"
	}

	if strings.TrimSpace(session.CreatedByName) == "" {
		session.CreatedByName = "System"
	}

	session.CreatedByInitials = strings.ToUpper(helpers.Initials(session.CreatedByName))
	if session.CreatedByInitials == "" {
		session.CreatedByInitials = "SY"
	}
	session.CreatedByAvatarClass = stockCheckSessionAvatarClass(session.CreatedByName)
	session.InitiationTypeLabel, session.InitiationTypeBadgeClass = stockCheckSessionTypeMeta(session.InitiationType)
	session.StatusLabel, session.StatusTextClass, session.StatusDotClass = stockCheckSessionStatusMeta(session.Status)

	return session, nil
}

func stockCheckSessionTypeMeta(value string) (string, string) {
	switch value {
	case "checker_initiative":
		return "Initiative", "bg-slate-200 text-slate-600"
	case "scheduled":
		return "Scheduled", "bg-[#dbeafe] text-[#1d4ed8]"
	default:
		return "Unknown", "bg-slate-100 text-slate-500"
	}
}

func stockCheckSessionStatusMeta(value string) (string, string, string) {
	switch value {
	case "draft":
		return "Draft", "session-status-text-draft", "session-status-dot-draft"
	case "in_progress":
		return "In Progress", "session-status-text-in-progress", "session-status-dot-in-progress"
	case "submitted":
		return "Submitted", "session-status-text-submitted", "session-status-dot-submitted"
	case "reviewed":
		return "Reviewed", "session-status-text-reviewed", "session-status-dot-reviewed"
	case "closed":
		return "Closed", "session-status-text-closed", "session-status-dot-closed"
	case "po":
		return "PO", "session-status-text-po", "session-status-dot-po"
	case "cancelled":
		return "Cancelled", "session-status-text-cancelled", "session-status-dot-cancelled"
	default:
		return "Unknown", "session-status-text-draft", "session-status-dot-draft"
	}
}

func stockCheckSessionAvatarClass(name string) string {
	sum := 0
	for _, ch := range name {
		sum += int(ch)
	}

	switch sum % 4 {
	case 0:
		return "bg-[#dbeafe] text-[#1d4ed8]"
	case 1:
		return "bg-[#ede9fe] text-[#6d28d9]"
	case 2:
		return "bg-[#ffedd5] text-[#ea580c]"
	default:
		return "bg-[#e2e8f0] text-slate-600"
	}
}

func nullableText(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableHistoryValue(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func formatHistoryText(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func formatHistoryFloat(value sql.NullFloat64) string {
	if !value.Valid {
		return ""
	}
	return formatHistoryDecimal(value.Float64)
}

func formatHistoryDecimal(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func buildStockCheckSessionNumber(storeCode string, sessionDate string, sequence int) string {
	return fmt.Sprintf("SCS-%s-%s-%03d", storeCode, strings.ReplaceAll(sessionDate, "-", ""), sequence)
}

func formatStockCheckDecimal(value float64) string {
	return fmt.Sprintf("%0.2f", value)
}

func formatStockCheckWholeNumber(value float64) string {
	return fmt.Sprintf("%.0f", math.Round(value))
}

func formatStockCheckCurrency(value float64) string {
	return fmt.Sprintf("Rp %s", formatStockCheckDecimal(value))
}

func computeStockCheckQtyInPcs(carton int, box int, pcs int, pcsPerBox int, pcsPerCarton int) (float64, error) {
	// Allow input even when conversion is missing by falling back to 1:1 units.
	if pcsPerCarton <= 0 {
		pcsPerCarton = 1
	}
	if pcsPerBox <= 0 {
		pcsPerBox = 1
	}

	total := (carton * pcsPerCarton) + (box * pcsPerBox) + pcs
	return float64(total), nil
}

func formatStockCheckUnitBreakdown(carton int, box int, pcs int) string {
	parts := make([]string, 0, 3)
	if carton > 0 {
		parts = append(parts, fmt.Sprintf("%d ctn", carton))
	}
	if box > 0 {
		parts = append(parts, fmt.Sprintf("%d box", box))
	}
	if pcs > 0 {
		parts = append(parts, fmt.Sprintf("%d pcs", pcs))
	}
	if len(parts) == 0 {
		return "0 pcs"
	}
	return strings.Join(parts, " ")
}

func formatStockCheckUnitBreakdownShort(carton int, box int, pcs int) string {
	return fmt.Sprintf("%dc - %db - %dp", carton, box, pcs)
}

func formatStockCheckConversion(pcsPerBox int, boxPerCarton int, pcsPerCarton int) string {
	parts := make([]string, 0, 3)
	if pcsPerBox > 0 {
		parts = append(parts, fmt.Sprintf("1 box = %d pcs", pcsPerBox))
	}
	if boxPerCarton > 0 && pcsPerCarton > 0 {
		parts = append(parts, fmt.Sprintf("1 carton = %d box = %d pcs", boxPerCarton, pcsPerCarton))
	} else if boxPerCarton > 0 {
		parts = append(parts, fmt.Sprintf("1 carton = %d box", boxPerCarton))
	} else if pcsPerCarton > 0 {
		parts = append(parts, fmt.Sprintf("1 carton = %d pcs", pcsPerCarton))
	}
	if len(parts) == 0 {
		return "Input satuan dasar pcs"
	}
	return strings.Join(parts, " | ")
}

func formatStockCheckSuggestCarton(suggestQty float64, pcsPerCarton int) (int, string) {
	if suggestQty <= 0 {
		return 0, "0 carton"
	}
	if pcsPerCarton <= 0 {
		return int(math.Ceil(suggestQty)), formatStockCheckWholeNumber(suggestQty)
	}

	carton := int(math.Ceil(suggestQty / float64(pcsPerCarton)))
	return carton, fmt.Sprintf("%d carton", carton)
}

func formatStockCheckSuggestBreakdown(suggestQty float64, pcsPerBox int, pcsPerCarton int) (int, int, int, string, string) {
	totalPcs := int(math.Round(suggestQty))
	if totalPcs <= 0 {
		return 0, 0, 0, "0 pcs", "0 pcs"
	}

	carton := 0
	box := 0
	pcs := totalPcs

	if pcsPerCarton > 0 {
		carton = pcs / pcsPerCarton
		pcs = pcs % pcsPerCarton
	}
	if pcsPerBox > 0 {
		box = pcs / pcsPerBox
		pcs = pcs % pcsPerBox
	}

	return carton, box, pcs, formatStockCheckSuggestHeadline(carton, box, pcs), formatStockCheckUnitBreakdown(carton, box, pcs)
}

func resolveStockCheckSuggestBreakdown(suggestQty float64, suggestCarton int, suggestBox int, suggestPcs int, pcsPerBox int, pcsPerCarton int) (int, int, int, string, string) {
	if suggestCarton == 0 && suggestBox == 0 && suggestPcs == 0 && suggestQty > 0 {
		return formatStockCheckSuggestBreakdown(suggestQty, pcsPerBox, pcsPerCarton)
	}

	return suggestCarton, suggestBox, suggestPcs, formatStockCheckSuggestHeadline(suggestCarton, suggestBox, suggestPcs), formatStockCheckUnitBreakdown(suggestCarton, suggestBox, suggestPcs)
}

func formatStockCheckSuggestHeadline(carton int, box int, pcs int) string {
	switch {
	case carton > 0:
		return fmt.Sprintf("%d carton", carton)
	case box > 0:
		return fmt.Sprintf("%d box", box)
	default:
		return fmt.Sprintf("%d pcs", pcs)
	}
}

func pickPreferredProductBarcode(barcode string, barcodeBox string, barcodeCarton string) string {
	for _, value := range []string{barcode, barcodeBox, barcodeCarton} {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func buildProductBarcodeSummary(barcode string, barcodeBox string, barcodeCarton string) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(barcode) != "" {
		parts = append(parts, "Unit: "+strings.TrimSpace(barcode))
	}
	if strings.TrimSpace(barcodeBox) != "" {
		parts = append(parts, "Box: "+strings.TrimSpace(barcodeBox))
	}
	if strings.TrimSpace(barcodeCarton) != "" {
		parts = append(parts, "Carton: "+strings.TrimSpace(barcodeCarton))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " | ")
}

func stockCheckSessionItemStatusMeta(status string) (string, string) {
	switch status {
	case "approved":
		return "Approved", "item-status-approved"
	case "po_created":
		return "PO Created", "item-status-po-created"
	case "reviewed":
		return "Reviewed", "item-status-pending"
	case "submitted":
		return "Submitted", "item-status-pending"
	case "rejected":
		return "Rejected", "item-status-rejected"
	default:
		return "Draft", "item-status-draft"
	}
}

func stockCheckSessionConditionMeta(status string) (string, string, string) {
	switch status {
	case "empty_rack":
		return "Empty Rack", "condition-empty-rack", "buyer-note-amber"
	case "damaged":
		return "Damaged", "condition-danger", "buyer-note-danger"
	case "missing":
		return "Missing", "condition-danger", "buyer-note-danger"
	case "overstock":
		return "Overstock", "condition-info", "buyer-note-info"
	case "other":
		return "Other", "condition-neutral", "buyer-note-default"
	default:
		return "Good", "condition-good", "buyer-note-default"
	}
}

func stockCheckSessionProductAvatarClass(name string) string {
	sum := 0
	for _, ch := range name {
		sum += int(ch)
	}

	switch sum % 4 {
	case 0:
		return "linear-gradient(145deg, #0f2f82 0%, #2149a6 100%)"
	case 1:
		return "linear-gradient(145deg, #ad3f22 0%, #dc6b2e 100%)"
	case 2:
		return "linear-gradient(145deg, #0f6a5b 0%, #1f9f82 100%)"
	default:
		return "linear-gradient(145deg, #3b455c 0%, #66738c 100%)"
	}
}
