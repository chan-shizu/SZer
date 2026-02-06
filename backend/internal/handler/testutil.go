package handler

import (
	"database/sql"
	"testing"

	"github.com/chan-shizu/SZer/db"
	_ "github.com/lib/pq"
)

// setupTestDBはintegration test用のテストDB接続を返す共通関数だよ！
func setupTestDB(t *testing.T) (*sql.DB, *db.Queries) {
	dsn := "postgres://test_user:test_pass@postgres-test:5432/test_db?sslmode=disable"
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL must be set for integration tests")
	}
	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	return dbConn, db.New(dbConn)
}
