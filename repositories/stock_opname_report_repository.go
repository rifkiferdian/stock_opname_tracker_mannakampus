package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type StockOpnameReportRepository struct {
	DB *sql.DB
}

type StockOpnameReportRecord struct {
	ItemID              int
	ProductID           int
	ProductCode         string
	Barcode             string
	ProductName         string
	Brand               string
	CategoryName        string
	DefaultLeadTimeDays int
	PcsPerBox           int
	BoxPerCarton        int
	PcsPerCarton        int
	SessionID           int
	SessionDate         time.Time
	QtyStoreCarton      int
	QtyStoreBox         int
	QtyStorePcs         int
	QtyStore            float64
	QtyWarehouseCarton  int
	QtyWarehouseBox     int
	QtyWarehousePcs     int
	QtyWarehouse        float64
	SystemQtyStore      float64
	SystemQtyWarehouse  float64
	SuggestBuyCarton    int
	SuggestBuyBox       int
	SuggestBuyPcs       int
	SuggestBuyQty       float64
	ApprovedBuyQty      float64
	Status              string
	ConditionStatus     string
	CheckerNotes        string
	BuyerNotes          string
}

type StockOpnameProductMonthlyPORecord struct {
	ProductID int
	MonthKey  string
	POQty     float64
}

func (r *StockOpnameReportRepository) GetDistinctSessionDates(supplierID int, limit int) ([]time.Time, error) {
	query := `
		SELECT DISTINCT scs.session_date
		FROM stock_check_sessions scs
	`
	args := []interface{}{supplierID}
	query += `
		WHERE scs.supplier_id = ?
	`
	query += `
		ORDER BY scs.session_date DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0, limit)
	for rows.Next() {
		var sessionDate sql.NullTime
		if err := rows.Scan(&sessionDate); err != nil {
			return nil, err
		}
		if sessionDate.Valid {
			dates = append(dates, sessionDate.Time)
		}
	}

	return dates, rows.Err()
}

func (r *StockOpnameReportRepository) GetStatusOptions(supplierID int, currentDate time.Time) ([]string, error) {
	rows, err := r.DB.Query(`
		SELECT DISTINCT COALESCE(si.status, '') AS status
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		WHERE scs.supplier_id = ?
			AND scs.session_date = ?
			AND COALESCE(si.status, '') <> ''
		ORDER BY status ASC
	`, supplierID, currentDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statuses := make([]string, 0)
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	return statuses, rows.Err()
}

func (r *StockOpnameReportRepository) CountSessions(supplierID int) (int, error) {
	query := `
		SELECT COUNT(DISTINCT scs.id)
		FROM stock_check_sessions scs
	`
	args := []interface{}{supplierID}
	query += `
		WHERE scs.supplier_id = ?
	`

	var total int
	err := r.DB.QueryRow(query, args...).Scan(&total)
	return total, err
}

func (r *StockOpnameReportRepository) GetReportRecords(supplierID int, status string, currentDate time.Time, dates []time.Time) ([]StockOpnameReportRecord, error) {
	if len(dates) == 0 {
		return []StockOpnameReportRecord{}, nil
	}

	placeholders := make([]string, 0, len(dates))
	args := make([]interface{}, 0, len(dates)+2)
	args = append(args, supplierID)
	for _, date := range dates {
		placeholders = append(placeholders, "?")
		args = append(args, date.Format("2006-01-02"))
	}

	query := fmt.Sprintf(`
		SELECT
			si.id,
			p.id,
			p.product_code,
			COALESCE(p.barcode, '') AS barcode,
			p.product_name,
			COALESCE(p.brand, '') AS brand,
			COALESCE(pc.category_name, 'Tanpa Kategori') AS category_name,
			COALESCE(p.default_lead_time_days, 0) AS default_lead_time_days,
			COALESCE(p.pcs_per_box, 0) AS pcs_per_box,
			COALESCE(p.box_per_carton, 0) AS box_per_carton,
			COALESCE(p.pcs_per_carton, 0) AS pcs_per_carton,
			scs.id AS session_id,
			scs.session_date,
			COALESCE(si.qty_store_carton, 0) AS qty_store_carton,
			COALESCE(si.qty_store_box, 0) AS qty_store_box,
			COALESCE(si.qty_store_pcs, 0) AS qty_store_pcs,
			COALESCE(si.qty_store, 0) AS qty_store,
			COALESCE(si.qty_warehouse_carton, 0) AS qty_warehouse_carton,
			COALESCE(si.qty_warehouse_box, 0) AS qty_warehouse_box,
			COALESCE(si.qty_warehouse_pcs, 0) AS qty_warehouse_pcs,
			COALESCE(si.qty_warehouse, 0) AS qty_warehouse,
			COALESCE(si.system_qty_store, 0) AS system_qty_store,
			COALESCE(si.system_qty_warehouse, 0) AS system_qty_warehouse,
			COALESCE(si.suggest_buy_carton, 0) AS suggest_buy_carton,
			COALESCE(si.suggest_buy_box, 0) AS suggest_buy_box,
			COALESCE(si.suggest_buy_pcs, 0) AS suggest_buy_pcs,
			(
				COALESCE(si.suggest_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.suggest_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.suggest_buy_pcs, 0)
			) AS suggest_qty,
			COALESCE(si.approved_buy_qty, 0) AS approved_buy_qty,
			COALESCE(si.status, '') AS status,
			COALESCE(si.condition_status, '') AS condition_status,
			COALESCE(si.checker_notes, '') AS checker_notes,
			COALESCE(si.buyer_notes, '') AS buyer_notes
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		INNER JOIN products p ON p.id = si.product_id
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		WHERE scs.supplier_id = ?
			AND scs.session_date IN (%s)
	`, strings.Join(placeholders, ","))

	if strings.TrimSpace(status) != "" {
		query += `
			AND EXISTS (
				SELECT 1
				FROM stock_check_session_items si_current
				INNER JOIN stock_check_sessions scs_current ON scs_current.id = si_current.stock_check_session_id
				WHERE scs_current.supplier_id = ?
					AND scs_current.session_date = ?
					AND si_current.product_id = p.id
					AND si_current.status = ?
					AND scs_current.id = (
						SELECT MAX(scs_latest.id)
						FROM stock_check_session_items si_latest
						INNER JOIN stock_check_sessions scs_latest ON scs_latest.id = si_latest.stock_check_session_id
						WHERE scs_latest.supplier_id = scs_current.supplier_id
							AND scs_latest.session_date = scs_current.session_date
							AND si_latest.product_id = si_current.product_id
					)
			)
		`
		args = append(args, supplierID, currentDate.Format("2006-01-02"), status)
	}

	query += `
		ORDER BY
			p.product_name ASC,
			scs.session_date DESC,
			scs.id DESC,
			si.id DESC
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]StockOpnameReportRecord, 0)
	for rows.Next() {
		var (
			record             StockOpnameReportRecord
			sessionDate        sql.NullTime
			qtyStore           sql.NullFloat64
			qtyWarehouse       sql.NullFloat64
			systemQtyStore     sql.NullFloat64
			systemQtyWarehouse sql.NullFloat64
			suggestBuyQty      sql.NullFloat64
			approvedBuyQty     sql.NullFloat64
		)

		if err := rows.Scan(
			&record.ItemID,
			&record.ProductID,
			&record.ProductCode,
			&record.Barcode,
			&record.ProductName,
			&record.Brand,
			&record.CategoryName,
			&record.DefaultLeadTimeDays,
			&record.PcsPerBox,
			&record.BoxPerCarton,
			&record.PcsPerCarton,
			&record.SessionID,
			&sessionDate,
			&record.QtyStoreCarton,
			&record.QtyStoreBox,
			&record.QtyStorePcs,
			&qtyStore,
			&record.QtyWarehouseCarton,
			&record.QtyWarehouseBox,
			&record.QtyWarehousePcs,
			&qtyWarehouse,
			&systemQtyStore,
			&systemQtyWarehouse,
			&record.SuggestBuyCarton,
			&record.SuggestBuyBox,
			&record.SuggestBuyPcs,
			&suggestBuyQty,
			&approvedBuyQty,
			&record.Status,
			&record.ConditionStatus,
			&record.CheckerNotes,
			&record.BuyerNotes,
		); err != nil {
			return nil, err
		}

		if sessionDate.Valid {
			record.SessionDate = sessionDate.Time
		}
		if qtyStore.Valid {
			record.QtyStore = qtyStore.Float64
		}
		if qtyWarehouse.Valid {
			record.QtyWarehouse = qtyWarehouse.Float64
		}
		if systemQtyStore.Valid {
			record.SystemQtyStore = systemQtyStore.Float64
		}
		if systemQtyWarehouse.Valid {
			record.SystemQtyWarehouse = systemQtyWarehouse.Float64
		}
		if suggestBuyQty.Valid {
			record.SuggestBuyQty = suggestBuyQty.Float64
		}
		if approvedBuyQty.Valid {
			record.ApprovedBuyQty = approvedBuyQty.Float64
		}

		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *StockOpnameReportRepository) GetMonthlyApprovalCounts(supplierID int, fromDate time.Time) (map[string]float64, error) {
	query := `
		SELECT
			DATE_FORMAT(scs.session_date, '%Y-%m') AS month_key,
			COALESCE(SUM(
				CASE
					WHEN si.status IN ('approved', 'po_created') THEN COALESCE(si.approved_buy_qty, 0)
					ELSE 0
				END
			), 0) AS approval_qty
		FROM stock_check_sessions scs
		LEFT JOIN stock_check_session_items si ON si.stock_check_session_id = scs.id
		WHERE scs.supplier_id = ?
			AND scs.session_date >= ?
	`
	args := []interface{}{supplierID, fromDate.Format("2006-01-02")}
	query += `
		GROUP BY DATE_FORMAT(scs.session_date, '%Y-%m')
		ORDER BY month_key ASC
	`
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]float64)
	for rows.Next() {
		var monthKey string
		var approvalQty sql.NullFloat64
		if err := rows.Scan(&monthKey, &approvalQty); err != nil {
			return nil, err
		}
		if approvalQty.Valid {
			counts[monthKey] = approvalQty.Float64
		} else {
			counts[monthKey] = 0
		}
	}

	return counts, rows.Err()
}

func (r *StockOpnameReportRepository) GetProductMonthlyPORecords(supplierID int, status string, currentDate time.Time, fromDate time.Time) ([]StockOpnameProductMonthlyPORecord, error) {
	query := `
		SELECT
			base.product_id,
			base.month_key,
			COALESCE(SUM(base.session_po_qty), 0) AS po_qty
		FROM (
			SELECT
				si.product_id,
				scs.id AS session_id,
				DATE_FORMAT(scs.session_date, '%Y-%m') AS month_key,
				COALESCE(SUM(COALESCE(si.approved_buy_qty, 0)), 0) AS session_po_qty
			FROM stock_check_sessions scs
			INNER JOIN stock_check_session_items si ON si.stock_check_session_id = scs.id
			WHERE scs.supplier_id = ?
				AND scs.session_date >= ?
	`
	args := []interface{}{supplierID, fromDate.Format("2006-01-02")}

	if strings.TrimSpace(status) != "" {
		query += `
			AND EXISTS (
				SELECT 1
				FROM stock_check_session_items si_current
				INNER JOIN stock_check_sessions scs_current ON scs_current.id = si_current.stock_check_session_id
				WHERE scs_current.supplier_id = ?
					AND scs_current.session_date = ?
					AND si_current.product_id = si.product_id
					AND si_current.status = ?
					AND scs_current.id = (
						SELECT MAX(scs_latest.id)
						FROM stock_check_session_items si_latest
						INNER JOIN stock_check_sessions scs_latest ON scs_latest.id = si_latest.stock_check_session_id
						WHERE scs_latest.supplier_id = scs_current.supplier_id
							AND scs_latest.session_date = scs_current.session_date
							AND si_latest.product_id = si_current.product_id
					)
			)
		`
		args = append(args, supplierID, currentDate.Format("2006-01-02"), status)
	}

	query += `
			GROUP BY si.product_id, scs.id, DATE_FORMAT(scs.session_date, '%Y-%m')
		) base
		GROUP BY base.product_id, base.month_key
		ORDER BY base.month_key ASC, base.product_id ASC
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]StockOpnameProductMonthlyPORecord, 0)
	for rows.Next() {
		var (
			record StockOpnameProductMonthlyPORecord
			poQty  sql.NullFloat64
		)

		if err := rows.Scan(&record.ProductID, &record.MonthKey, &poQty); err != nil {
			return nil, err
		}
		if poQty.Valid {
			record.POQty = poQty.Float64
		}

		records = append(records, record)
	}

	return records, rows.Err()
}
