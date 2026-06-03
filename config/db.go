package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() {

	var err error

	// enable parseTime so DATETIME/TIMESTAMP scan into time.Time instead of []byte
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = DB.Ping()
	if err != nil {
		panic(err)
	}

	if err := ensureAppSchema(DB); err != nil {
		panic(err)
	}

	fmt.Println("Database connected successfully")
}

func ensureAppSchema(db *sql.DB) error {
	if err := ensureSupplierProductGroupingSchema(db); err != nil {
		return err
	}
	return nil
}

func ensureSupplierProductGroupingSchema(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS supplier_product_groups (
			id INT NOT NULL AUTO_INCREMENT,
			supplier_id INT NOT NULL,
			group_name VARCHAR(150) NOT NULL,
			description TEXT NULL,
			sort_order INT NOT NULL DEFAULT 0,
			is_active TINYINT(1) NOT NULL DEFAULT 1,
			created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_supplier_product_groups_supplier (supplier_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
	`); err != nil {
		return err
	}

	if !columnExists(db, "product_suppliers", "supplier_product_group_id") {
		if _, err := db.Exec(`
			ALTER TABLE product_suppliers
			ADD COLUMN supplier_product_group_id INT NULL AFTER supplier_id
		`); err != nil {
			return err
		}
	}

	if !indexExists(db, "supplier_product_groups", "idx_supplier_product_groups_supplier_name") {
		if _, err := db.Exec(`
			ALTER TABLE supplier_product_groups
			ADD UNIQUE KEY idx_supplier_product_groups_supplier_name (supplier_id, group_name)
		`); err != nil {
			return err
		}
	}

	if !indexExists(db, "product_suppliers", "idx_product_suppliers_group") {
		if _, err := db.Exec(`
			ALTER TABLE product_suppliers
			ADD KEY idx_product_suppliers_group (supplier_product_group_id)
		`); err != nil {
			return err
		}
	}

	if !foreignKeyExists(db, "supplier_product_groups", "fk_supplier_product_groups_supplier") {
		if _, err := db.Exec(`
			ALTER TABLE supplier_product_groups
			ADD CONSTRAINT fk_supplier_product_groups_supplier
			FOREIGN KEY (supplier_id) REFERENCES suppliers (id)
			ON DELETE CASCADE ON UPDATE CASCADE
		`); err != nil {
			return err
		}
	}

	if !foreignKeyExists(db, "product_suppliers", "fk_product_suppliers_group") {
		if _, err := db.Exec(`
			ALTER TABLE product_suppliers
			ADD CONSTRAINT fk_product_suppliers_group
			FOREIGN KEY (supplier_product_group_id) REFERENCES supplier_product_groups (id)
			ON DELETE SET NULL ON UPDATE CASCADE
		`); err != nil {
			return err
		}
	}

	return nil
}

func columnExists(db *sql.DB, tableName string, columnName string) bool {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count)
	return err == nil && count > 0
}

func indexExists(db *sql.DB, tableName string, indexName string) bool {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND INDEX_NAME = ?
	`, tableName, indexName).Scan(&count)
	return err == nil && count > 0
}

func foreignKeyExists(db *sql.DB, tableName string, constraintName string) bool {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.TABLE_CONSTRAINTS
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND CONSTRAINT_NAME = ?
			AND CONSTRAINT_TYPE = 'FOREIGN KEY'
	`, tableName, constraintName).Scan(&count)
	return err == nil && count > 0
}
