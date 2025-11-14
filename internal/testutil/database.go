package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetupTestDB creates and returns a test database connection
// It reads the TEST_DATABASE_URL environment variable
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL environment variable not set")
	}

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run migrations
	RunMigrations(t, db.DB)

	return db
}

// RunMigrations runs all up migrations on the test database
func RunMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatalf("failed to create migration driver: %v", err)
	}

	// Find migrations directory relative to project root
	migrationsPath := findMigrationsPath()
	
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		t.Fatalf("failed to create migration instance: %v", err)
	}

	// Run migrations up
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}
}

// findMigrationsPath finds the migrations directory from the project root
func findMigrationsPath() string {
	// Try current directory first
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}
	
	// Try parent directory (for tests in internal/)
	if _, err := os.Stat("../migrations"); err == nil {
		return "../migrations"
	}
	
	// Try two levels up (for tests in internal/handlers/)
	if _, err := os.Stat("../../migrations"); err == nil {
		return "../../migrations"
	}
	
	// Default to absolute path
	return "/Users/sunfmin/Developments/apidemo1/migrations"
}

// BeginTestTransaction starts a new transaction for test isolation
// The transaction will be rolled back automatically via t.Cleanup
func BeginTestTransaction(t *testing.T, db *sqlx.DB) *sqlx.Tx {
	t.Helper()

	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Ensure rollback on test completion
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Errorf("failed to rollback transaction: %v", err)
		}
	})

	return tx
}

// TruncateTables truncates all tables in the test database
// Use this sparingly - transaction rollback is preferred
func TruncateTables(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

