package handler

import (
	"database/sql"
	"testing"

	"github.com/chan-shizu/SZer/db"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// cleanupProgramDetailsTestData はテストデータを削除する（integration test用共通）
func cleanupProgramDetailsTestData(t *testing.T, dbConn *sql.DB) {
	tables := []string{
		"permitted_program_users",
		"program_category_tags",
		"program_performers",
		"likes",
		"watch_histories",
		"programs",
		"category_tags",
		"performers",
		"user",
	}
	for _, table := range tables {
		// テーブル名を必ずダブルクォートで囲む（PostgreSQL予約語対策）
		_, err := dbConn.Exec(`DELETE FROM "` + table + `"`)
		if err != nil {
			t.Logf("cleanup warning: failed to delete from %s: %v", table, err)
		}
	}
}


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
		
	// テストデータをクリーンアップ
	t.Cleanup(func() { cleanupProgramDetailsTestData(t, dbConn) })

	return dbConn, db.New(dbConn)
}

// テスト用OptionalAuthモック: 引数で指定したidでuser_idをセット
func MockOptionalAuth(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}