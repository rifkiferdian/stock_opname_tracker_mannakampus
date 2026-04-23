package repositories

import (
	"database/sql"
	"fmt"
	helpers "gobase-app/helper"
	"gobase-app/models"
	"strings"
)

type StockCheckSessionRepository struct {
	DB *sql.DB
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

func (r *StockCheckSessionRepository) GetSupplierOptions() ([]models.Supplier, error) {
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

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		if err := rows.Scan(&supplier.ID, &supplier.SupplierName); err != nil {
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

func (r *StockCheckSessionRepository) Create(input models.StockCheckSessionCreateInput) error {
	_, err := r.DB.Exec(`
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

	return err
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

func (r *StockCheckSessionRepository) DeleteByID(id int) error {
	_, err := r.DB.Exec(`DELETE FROM stock_check_sessions WHERE id = ?`, id)
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

func buildStockCheckSessionNumber(storeCode string, sessionDate string, sequence int) string {
	return fmt.Sprintf("SCS-%s-%s-%03d", storeCode, strings.ReplaceAll(sessionDate, "-", ""), sequence)
}
