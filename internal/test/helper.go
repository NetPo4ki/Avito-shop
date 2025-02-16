package test

import (
	"avito-shop/internal/config"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	dbConfig := config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5433",
		User:     "postgres",
		Password: "postgres",
		DBName:   "avito_shop_test",
		SSLMode:  "disable",
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	_, err = db.Exec(`
		DROP TABLE IF EXISTS coin_transactions;
		DROP TABLE IF EXISTS user_inventory;
		DROP TABLE IF EXISTS merchandise;
		DROP TABLE IF EXISTS users;
	`)
	if err != nil {
		t.Fatalf("Failed to drop existing tables: %v", err)
	}

	schema, err := os.ReadFile("../../migrations/001_initial_schema.sql")
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to execute schema: %v", err)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close test database connection: %v", err)
		}
	}
}

func ClearTestDB(t *testing.T, db *sql.DB) {
	tables := []string{"coin_transactions", "user_inventory", "merchandise", "users"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Fatalf("Failed to clear table %s: %v", table, err)
		}
	}
}
