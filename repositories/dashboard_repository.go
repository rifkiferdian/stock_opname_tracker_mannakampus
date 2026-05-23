package repositories

import (
	"database/sql"
	"fmt"
	"gobase-app/models"
	"math"
	"time"
)

type DashboardRepository struct {
	DB *sql.DB
}

func (r *DashboardRepository) GetBuyerDashboardMetrics(userID int, currentDate string) (models.BuyerDashboardMetricsSnapshot, error) {
	query := `
		SELECT
			COUNT(DISTINCT CASE WHEN si.status IN ('draft', 'submitted', 'reviewed') THEN scs.id END) AS review_session_count,
			COALESCE(SUM(CASE WHEN si.status IN ('draft', 'submitted', 'reviewed') THEN 1 ELSE 0 END), 0) AS pending_sku_count,
			COALESCE(SUM(CASE WHEN si.status IN ('approved', 'po_created') THEN 1 ELSE 0 END), 0) AS approved_sku_count,
			COUNT(DISTINCT CASE WHEN si.status IN ('draft', 'submitted', 'reviewed') THEN COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id) END) AS supplier_follow_up_count,
			COUNT(DISTINCT CASE WHEN scs.session_date = ? THEN scs.id END) AS today_session_count,
			(
				SELECT COUNT(1)
				FROM products p
				WHERE p.is_active = 1
			) AS active_product_count,
			(
				SELECT COUNT(1)
				FROM suppliers sp
				WHERE sp.is_active = 1
			) AS active_supplier_count,
			COALESCE(SUM(
				CASE
					WHEN si.status IN ('draft', 'submitted', 'reviewed') THEN (
						COALESCE(si.suggest_buy_carton, 0) * COALESCE((SELECT pconv.pcs_per_carton FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
						COALESCE(si.suggest_buy_box, 0) * COALESCE((SELECT pconv.pcs_per_box FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
						COALESCE(si.suggest_buy_pcs, 0)
					) * COALESCE(
						(
							SELECT ps.last_price
							FROM product_suppliers ps
							WHERE ps.product_id = si.product_id
								AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
								AND ps.is_active = 1
							ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
							LIMIT 1
						),
						(
							SELECT ps.last_price
							FROM product_suppliers ps
							WHERE ps.product_id = si.product_id
								AND ps.is_active = 1
							ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
							LIMIT 1
						),
						0
					)
					ELSE 0
				END
			), 0) AS pending_value,
			COALESCE(SUM(
				CASE
					WHEN si.status IN ('approved', 'po_created') THEN (
						COALESCE(si.approved_buy_carton, 0) * COALESCE((SELECT pconv.pcs_per_carton FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
						COALESCE(si.approved_buy_box, 0) * COALESCE((SELECT pconv.pcs_per_box FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
						COALESCE(si.approved_buy_pcs, 0)
					) * COALESCE(
						(
							SELECT ps.last_price
							FROM product_suppliers ps
							WHERE ps.product_id = si.product_id
								AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
								AND ps.is_active = 1
							ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
							LIMIT 1
						),
						(
							SELECT ps.last_price
							FROM product_suppliers ps
							WHERE ps.product_id = si.product_id
								AND ps.is_active = 1
							ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
							LIMIT 1
						),
						0
					)
					ELSE 0
				END
			), 0) AS approved_value
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
	`
	args := make([]interface{}, 0, 2)
	args = append(args, currentDate)
	if userID > 0 {
		query += " INNER JOIN user_stores us ON us.store_id = scs.store_id AND us.user_id = ?"
		args = append(args, userID)
	}

	row := r.DB.QueryRow(query, args...)

	var (
		metrics       models.BuyerDashboardMetricsSnapshot
		pendingValue  sql.NullFloat64
		approvedValue sql.NullFloat64
	)

	if err := row.Scan(
		&metrics.ReviewSessionCount,
		&metrics.PendingSKUCount,
		&metrics.ApprovedSKUCount,
		&metrics.SupplierFollowUpCount,
		&metrics.TodaySessionCount,
		&metrics.ActiveProductCount,
		&metrics.ActiveSupplierCount,
		&pendingValue,
		&approvedValue,
	); err != nil {
		return metrics, err
	}

	if pendingValue.Valid {
		metrics.PendingValue = pendingValue.Float64
	}
	if approvedValue.Valid {
		metrics.ApprovedValue = approvedValue.Float64
	}

	metrics.PendingValueDisplay = formatStockCheckCurrency(metrics.PendingValue)
	metrics.ApprovedValueDisplay = formatStockCheckCurrency(metrics.ApprovedValue)

	return metrics, nil
}

func (r *DashboardRepository) GetBuyerDashboardPipeline(userID int) (models.BuyerDashboardPipelineSnapshot, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN si.status = 'draft' THEN 1 ELSE 0 END), 0) AS draft_count,
			COALESCE(SUM(CASE WHEN si.status IN ('submitted', 'reviewed') THEN 1 ELSE 0 END), 0) AS waiting_count,
			COALESCE(SUM(CASE WHEN si.status IN ('approved', 'po_created') THEN 1 ELSE 0 END), 0) AS approved_count,
			COALESCE(SUM(CASE WHEN si.status = 'rejected' THEN 1 ELSE 0 END), 0) AS rejected_count
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
	`
	args := make([]interface{}, 0, 1)
	if userID > 0 {
		query += " INNER JOIN user_stores us ON us.store_id = scs.store_id AND us.user_id = ?"
		args = append(args, userID)
	}

	row := r.DB.QueryRow(query, args...)

	var snapshot models.BuyerDashboardPipelineSnapshot
	err := row.Scan(
		&snapshot.DraftCount,
		&snapshot.WaitingCount,
		&snapshot.ApprovedCount,
		&snapshot.RejectedCount,
	)

	return snapshot, err
}

func (r *DashboardRepository) GetBuyerMonthlyPOCounts(userID int, fromDate time.Time) (map[string]int, error) {
	query := `
		SELECT
			DATE_FORMAT(scs.session_date, '%Y-%m') AS month_key,
			COALESCE(SUM(CASE WHEN si.status IN ('approved', 'po_created') THEN 1 ELSE 0 END), 0) AS po_count
		FROM stock_check_sessions scs
		LEFT JOIN stock_check_session_items si ON si.stock_check_session_id = scs.id
	`
	args := make([]interface{}, 0, 2)
	if userID > 0 {
		query += " INNER JOIN user_stores us ON us.store_id = scs.store_id AND us.user_id = ?"
		args = append(args, userID)
	}
	query += `
		WHERE scs.session_date >= ?
		GROUP BY DATE_FORMAT(scs.session_date, '%Y-%m')
		ORDER BY month_key ASC
	`
	args = append(args, fromDate.Format("2006-01-02"))

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var monthKey string
		var poCount int
		if err := rows.Scan(&monthKey, &poCount); err != nil {
			return nil, err
		}
		counts[monthKey] = poCount
	}

	return counts, rows.Err()
}

func (r *DashboardRepository) GetBuyerPriorityItems(userID int, limit int) ([]models.BuyerDashboardPriorityItem, error) {
	query := `
		SELECT
			scs.id,
			scs.session_number,
			scs.session_date,
			st.store_name,
			COALESCE(sel.supplier_name, sessup.supplier_name, '') AS supplier_name,
			p.product_code,
			p.product_name,
			COALESCE(si.checker_notes, '') AS checker_notes,
			si.condition_status,
			si.status,
			(
				COALESCE(si.qty_store_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.qty_store_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.qty_store_pcs, 0) +
				COALESCE(si.qty_warehouse_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.qty_warehouse_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.qty_warehouse_pcs, 0)
			) AS physical_qty,
			0 AS system_qty,
			(
				COALESCE(si.suggest_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.suggest_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.suggest_buy_pcs, 0)
			) AS suggest_qty,
			COALESCE(
				(
					SELECT ps.last_price
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
						AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
						AND ps.is_active = 1
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				(
					SELECT ps.last_price
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
						AND ps.is_active = 1
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				0
			) AS unit_price,
			COALESCE(
				(
					SELECT ps.lead_time_days
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
						AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
						AND ps.is_active = 1
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				(
					SELECT ps.lead_time_days
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
						AND ps.is_active = 1
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				0
			) AS lead_time_days
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		INNER JOIN stores st ON st.store_id = scs.store_id
		INNER JOIN products p ON p.id = si.product_id
		LEFT JOIN suppliers sessup ON sessup.id = scs.supplier_id
		LEFT JOIN suppliers sel ON sel.id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
	`
	args := make([]interface{}, 0, 2)
	if userID > 0 {
		query += " INNER JOIN user_stores us ON us.store_id = scs.store_id AND us.user_id = ?"
		args = append(args, userID)
	}
	query += `
		WHERE si.status IN ('draft', 'submitted', 'reviewed')
		ORDER BY
			CASE si.condition_status
				WHEN 'empty_rack' THEN 1
				WHEN 'missing' THEN 2
				WHEN 'damaged' THEN 3
				WHEN 'overstock' THEN 4
				ELSE 5
			END,
			((
				COALESCE(si.suggest_buy_carton, 0) * COALESCE(p.pcs_per_carton, 0) +
				COALESCE(si.suggest_buy_box, 0) * COALESCE(p.pcs_per_box, 0) +
				COALESCE(si.suggest_buy_pcs, 0)
			) * COALESCE(
				(
					SELECT ps.last_price
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
						AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
						AND ps.is_active = 1
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				(
					SELECT ps.last_price
					FROM product_suppliers ps
					WHERE ps.product_id = si.product_id
						AND ps.is_active = 1
					ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
					LIMIT 1
				),
				0
			)) DESC,
			scs.session_date DESC,
			si.id ASC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.BuyerDashboardPriorityItem, 0)
	for rows.Next() {
		var (
			item        models.BuyerDashboardPriorityItem
			sessionDate sql.NullTime
			physicalQty sql.NullFloat64
			systemQty   sql.NullFloat64
			suggestQty  sql.NullFloat64
			unitPrice   sql.NullFloat64
			condition   string
			status      string
		)

		if err := rows.Scan(
			&item.SessionID,
			&item.SessionNumber,
			&sessionDate,
			&item.StoreName,
			&item.SupplierName,
			&item.ProductCode,
			&item.ProductName,
			&item.CheckerNotes,
			&condition,
			&status,
			&physicalQty,
			&systemQty,
			&suggestQty,
			&unitPrice,
			&item.LeadTimeDays,
		); err != nil {
			return nil, err
		}

		if sessionDate.Valid {
			item.SessionDateDisplay = sessionDate.Time.Format("02 Jan 2006")
		} else {
			item.SessionDateDisplay = "-"
		}

		physicalValue := 0.0
		systemValue := 0.0
		suggestValue := 0.0
		priceValue := 0.0

		if physicalQty.Valid {
			physicalValue = physicalQty.Float64
		}
		if systemQty.Valid {
			systemValue = systemQty.Float64
		}
		if suggestQty.Valid {
			suggestValue = suggestQty.Float64
		}
		if unitPrice.Valid {
			priceValue = unitPrice.Float64
		}

		item.PhysicalQtyDisplay = formatStockCheckWholeNumber(physicalValue)
		item.SystemQtyDisplay = formatStockCheckWholeNumber(systemValue)
		item.SuggestBuyQtyDisplay = formatStockCheckWholeNumber(suggestValue)
		item.EstimatedValueDisplay = formatStockCheckCurrency(suggestValue * priceValue)
		item.ConditionLabel, item.ConditionBadgeClass = dashboardBuyerConditionMeta(condition)
		item.StatusLabel, item.StatusBadgeClass = dashboardBuyerItemStatusMeta(status)
		item.DetailURL = fmt.Sprintf("/stock-check-sessions/%d", item.SessionID)

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *DashboardRepository) GetBuyerWeeklySuppliers(userID int, weekStart time.Time, weekEnd time.Time, limit int) ([]models.BuyerDashboardWeeklySupplier, error) {
	query := `
		SELECT
			sp.id,
			COALESCE(sp.supplier_code, '') AS supplier_code,
			sp.supplier_name,
			MIN(scs.session_date) AS first_session_date,
			MAX(scs.session_date) AS latest_session_date,
			COUNT(DISTINCT scs.id) AS session_count,
			COALESCE(GROUP_CONCAT(DISTINCT st.store_name ORDER BY st.store_name ASC SEPARATOR ', '), '') AS store_names,
			COALESCE(SUM(CASE WHEN si.status IN ('draft', 'submitted', 'reviewed') THEN 1 ELSE 0 END), 0) AS pending_items,
			COALESCE(SUM(CASE WHEN si.status IN ('approved', 'po_created') THEN 1 ELSE 0 END), 0) AS approved_items
		FROM stock_check_sessions scs
		INNER JOIN suppliers sp ON sp.id = scs.supplier_id
		INNER JOIN stores st ON st.store_id = scs.store_id
		LEFT JOIN stock_check_session_items si ON si.stock_check_session_id = scs.id
	`
	args := make([]interface{}, 0, 4)
	if userID > 0 {
		query += " INNER JOIN user_stores us ON us.store_id = scs.store_id AND us.user_id = ?"
		args = append(args, userID)
	}
	query += `
		WHERE scs.session_date >= ? AND scs.session_date <= ?
		GROUP BY sp.id, sp.supplier_code, sp.supplier_name
		ORDER BY latest_session_date DESC, session_count DESC, sp.supplier_name ASC
		LIMIT ?
	`
	args = append(args, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"), limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.BuyerDashboardWeeklySupplier, 0)
	for rows.Next() {
		var (
			item          models.BuyerDashboardWeeklySupplier
			firstSession  sql.NullTime
			latestSession sql.NullTime
			storeNames    sql.NullString
			supplierCode  sql.NullString
		)

		if err := rows.Scan(
			&item.SupplierID,
			&supplierCode,
			&item.SupplierName,
			&firstSession,
			&latestSession,
			&item.SessionCount,
			&storeNames,
			&item.PendingItems,
			&item.ApprovedItems,
		); err != nil {
			return nil, err
		}

		if supplierCode.Valid {
			item.SupplierCode = supplierCode.String
		}
		if storeNames.Valid && storeNames.String != "" {
			item.StoreNames = storeNames.String
		} else {
			item.StoreNames = "-"
		}
		if latestSession.Valid {
			item.LatestSessionDateDisplay = latestSession.Time.Format("02 Jan 2006")
		} else if firstSession.Valid {
			item.LatestSessionDateDisplay = firstSession.Time.Format("02 Jan 2006")
		} else {
			item.LatestSessionDateDisplay = "-"
		}

		item.SupplierURL = fmt.Sprintf("/suppliers/%d", item.SupplierID)
		item.SessionListURL = fmt.Sprintf("/stock-check-sessions?supplier_id=%d&date_from=%s&date_to=%s", item.SupplierID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *DashboardRepository) GetBuyerSupplierQueues(userID int, limit int) ([]models.BuyerDashboardSupplierQueue, error) {
	query := `
		SELECT
			COALESCE(sel.id, sessup.id, 0) AS supplier_id,
			COALESCE(sel.supplier_code, sessup.supplier_code, '') AS supplier_code,
			COALESCE(sel.supplier_name, sessup.supplier_name, '') AS supplier_name,
			COALESCE(sel.payment_term_days, sessup.payment_term_days, 0) AS payment_term_days,
			COUNT(*) AS pending_items,
			COUNT(DISTINCT scs.id) AS open_sessions,
			COALESCE(SUM(CASE WHEN si.condition_status IN ('empty_rack', 'missing', 'damaged') THEN 1 ELSE 0 END), 0) AS critical_items,
			COALESCE(SUM(
				(
					COALESCE(si.suggest_buy_carton, 0) * COALESCE((SELECT pconv.pcs_per_carton FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
					COALESCE(si.suggest_buy_box, 0) * COALESCE((SELECT pconv.pcs_per_box FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
					COALESCE(si.suggest_buy_pcs, 0)
				) * COALESCE(
					(
						SELECT ps.last_price
						FROM product_suppliers ps
						WHERE ps.product_id = si.product_id
							AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
							AND ps.is_active = 1
						ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
						LIMIT 1
					),
					(
						SELECT ps.last_price
						FROM product_suppliers ps
						WHERE ps.product_id = si.product_id
							AND ps.is_active = 1
						ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
						LIMIT 1
					),
					0
				)
			), 0) AS estimated_value,
			COALESCE(AVG(NULLIF(
				COALESCE(
					(
						SELECT ps.lead_time_days
						FROM product_suppliers ps
						WHERE ps.product_id = si.product_id
							AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
							AND ps.is_active = 1
						ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
						LIMIT 1
					),
					(
						SELECT ps.lead_time_days
						FROM product_suppliers ps
						WHERE ps.product_id = si.product_id
							AND ps.is_active = 1
						ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
						LIMIT 1
					),
					0
				),
				0
			)), 0) AS avg_lead_time_days
		FROM stock_check_session_items si
		INNER JOIN stock_check_sessions scs ON scs.id = si.stock_check_session_id
		LEFT JOIN suppliers sessup ON sessup.id = scs.supplier_id
		LEFT JOIN suppliers sel ON sel.id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
	`
	args := make([]interface{}, 0, 2)
	if userID > 0 {
		query += " INNER JOIN user_stores us ON us.store_id = scs.store_id AND us.user_id = ?"
		args = append(args, userID)
	}
	query += `
		WHERE si.status IN ('draft', 'submitted', 'reviewed')
		GROUP BY COALESCE(sel.id, sessup.id, 0), COALESCE(sel.supplier_code, sessup.supplier_code, ''), COALESCE(sel.supplier_name, sessup.supplier_name, ''), COALESCE(sel.payment_term_days, sessup.payment_term_days, 0)
		ORDER BY critical_items DESC, estimated_value DESC, pending_items DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queues := make([]models.BuyerDashboardSupplierQueue, 0)
	for rows.Next() {
		var (
			item           models.BuyerDashboardSupplierQueue
			estimatedValue sql.NullFloat64
			avgLeadTime    sql.NullFloat64
		)

		if err := rows.Scan(
			&item.SupplierID,
			&item.SupplierCode,
			&item.SupplierName,
			&item.PaymentTermDays,
			&item.PendingItems,
			&item.OpenSessions,
			&item.CriticalItems,
			&estimatedValue,
			&avgLeadTime,
		); err != nil {
			return nil, err
		}

		if estimatedValue.Valid {
			item.EstimatedValueDisplay = formatStockCheckCurrency(estimatedValue.Float64)
		} else {
			item.EstimatedValueDisplay = formatStockCheckCurrency(0)
		}
		if avgLeadTime.Valid {
			item.AverageLeadTimeDays = int(math.Round(avgLeadTime.Float64))
		}

		item.SupplierURL = fmt.Sprintf("/suppliers/%d", item.SupplierID)
		item.SessionListURL = fmt.Sprintf("/stock-check-sessions?supplier_id=%d", item.SupplierID)
		queues = append(queues, item)
	}

	return queues, rows.Err()
}

func (r *DashboardRepository) GetBuyerSessionQueues(userID int, limit int) ([]models.BuyerDashboardSessionQueue, error) {
	query := `
		SELECT
			scs.id,
			scs.session_number,
			scs.session_date,
			st.store_name,
			sp.supplier_name,
			scs.status,
			COALESCE(SUM(CASE WHEN si.status IN ('draft', 'submitted', 'reviewed') THEN 1 ELSE 0 END), 0) AS pending_items,
			COALESCE(SUM(CASE WHEN si.status IN ('approved', 'po_created') THEN 1 ELSE 0 END), 0) AS approved_items,
			COALESCE(SUM(CASE WHEN si.status = 'rejected' THEN 1 ELSE 0 END), 0) AS rejected_items,
			COALESCE(SUM(
				CASE
					WHEN si.status IN ('draft', 'submitted', 'reviewed') THEN (
						COALESCE(si.suggest_buy_carton, 0) * COALESCE((SELECT pconv.pcs_per_carton FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
						COALESCE(si.suggest_buy_box, 0) * COALESCE((SELECT pconv.pcs_per_box FROM products pconv WHERE pconv.id = si.product_id LIMIT 1), 0) +
						COALESCE(si.suggest_buy_pcs, 0)
					) * COALESCE(
						(
							SELECT ps.last_price
							FROM product_suppliers ps
							WHERE ps.product_id = si.product_id
								AND ps.supplier_id = COALESCE(si.approved_supplier_id, si.suggested_supplier_id, scs.supplier_id)
								AND ps.is_active = 1
							ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
							LIMIT 1
						),
						(
							SELECT ps.last_price
							FROM product_suppliers ps
							WHERE ps.product_id = si.product_id
								AND ps.is_active = 1
							ORDER BY ps.is_primary DESC, ps.priority_no ASC, ps.id ASC
							LIMIT 1
						),
						0
					)
					ELSE 0
				END
			), 0) AS suggested_value
		FROM stock_check_sessions scs
		INNER JOIN stores st ON st.store_id = scs.store_id
		INNER JOIN suppliers sp ON sp.id = scs.supplier_id
		LEFT JOIN stock_check_session_items si ON si.stock_check_session_id = scs.id
	`
	args := make([]interface{}, 0, 2)
	if userID > 0 {
		query += " INNER JOIN user_stores us ON us.store_id = scs.store_id AND us.user_id = ?"
		args = append(args, userID)
	}
	query += `
		GROUP BY scs.id, scs.session_number, scs.session_date, st.store_name, sp.supplier_name, scs.status
		HAVING pending_items > 0 OR scs.status IN ('draft', 'in_progress', 'submitted', 'reviewed')
		ORDER BY
			CASE scs.status
				WHEN 'submitted' THEN 1
				WHEN 'reviewed' THEN 2
				WHEN 'in_progress' THEN 3
				WHEN 'draft' THEN 4
				ELSE 5
			END,
			suggested_value DESC,
			scs.session_date DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]models.BuyerDashboardSessionQueue, 0)
	for rows.Next() {
		var (
			item           models.BuyerDashboardSessionQueue
			sessionDate    sql.NullTime
			suggestedValue sql.NullFloat64
			status         string
		)

		if err := rows.Scan(
			&item.SessionID,
			&item.SessionNumber,
			&sessionDate,
			&item.StoreName,
			&item.SupplierName,
			&status,
			&item.PendingItems,
			&item.ApprovedItems,
			&item.RejectedItems,
			&suggestedValue,
		); err != nil {
			return nil, err
		}

		if sessionDate.Valid {
			item.SessionDateDisplay = sessionDate.Time.Format("02 Jan 2006")
		} else {
			item.SessionDateDisplay = "-"
		}
		if suggestedValue.Valid {
			item.SuggestedValueDisplay = formatStockCheckCurrency(suggestedValue.Float64)
		} else {
			item.SuggestedValueDisplay = formatStockCheckCurrency(0)
		}

		item.StatusLabel, item.StatusTextClass = dashboardBuyerSessionStatusMeta(status)
		item.DetailURL = fmt.Sprintf("/stock-check-sessions/%d", item.SessionID)
		sessions = append(sessions, item)
	}

	return sessions, rows.Err()
}

func dashboardBuyerConditionMeta(status string) (string, string) {
	switch status {
	case "empty_rack":
		return "Empty Rack", "bg-[#fff7ed] text-[#c2410c]"
	case "damaged":
		return "Damaged", "bg-[#fff1f2] text-[#be123c]"
	case "missing":
		return "Missing", "bg-[#fff1f2] text-[#be123c]"
	case "overstock":
		return "Overstock", "bg-[#eff6ff] text-[#1d4ed8]"
	case "other":
		return "Other", "bg-slate-100 text-slate-600"
	default:
		return "Good", "bg-[#ecfdf5] text-[#059669]"
	}
}

func dashboardBuyerItemStatusMeta(status string) (string, string) {
	switch status {
	case "approved":
		return "Approved", "bg-[#ecfdf5] text-[#059669]"
	case "po_created":
		return "PO Created", "bg-[#dcfce7] text-[#15803d]"
	case "reviewed":
		return "Reviewed", "bg-[#eff6ff] text-[#1d4ed8]"
	case "submitted":
		return "Submitted", "bg-[#fffbeb] text-[#b45309]"
	case "rejected":
		return "Rejected", "bg-[#fff1f2] text-[#be123c]"
	default:
		return "Draft", "bg-slate-100 text-slate-600"
	}
}

func dashboardBuyerSessionStatusMeta(status string) (string, string) {
	switch status {
	case "submitted":
		return "Submitted", "text-[#b45309]"
	case "reviewed":
		return "Reviewed", "text-[#1d4ed8]"
	case "in_progress":
		return "In Progress", "text-[#475569]"
	case "closed":
		return "Closed", "text-[#059669]"
	case "cancelled":
		return "Cancelled", "text-[#be123c]"
	default:
		return "Draft", "text-slate-600"
	}
}
