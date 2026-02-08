package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// テスト用ヘルパー: 認証済みユーザー付きルーターを作成
func setupPurchaseRouter(t *testing.T, userID string) (*gin.Engine, *Handler) {
	dbConn, q := setupTestDB(t)
	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)
	return r, h
}

func TestPurchaseProgram_Success_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "purchase-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "購入テストユーザー", "purchase1@example.com", 1000)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"purchase-test", "/video/purchase.mp4", true, 500,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/programs/%d/purchase", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Points float64 `json:"points"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// 1000 - 500 = 500
	assert.Equal(t, float64(500), resp.Points)

	// DBで権限が付与されていることを確認
	var permitted bool
	err = dbConn.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM permitted_program_users WHERE user_id = $1 AND program_id = $2)`,
		userID, programID,
	).Scan(&permitted)
	if err != nil {
		t.Fatalf("failed to query permitted_program_users: %v", err)
	}
	assert.True(t, permitted)

	// DBでポイントが減っていることを確認
	var points int32
	err = dbConn.QueryRow(`SELECT points FROM "user" WHERE id = $1`, userID).Scan(&points)
	if err != nil {
		t.Fatalf("failed to query user points: %v", err)
	}
	assert.Equal(t, int32(500), points)
}

func TestPurchaseProgram_InsufficientPoints_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "purchase-user-poor"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "ポイント不足ユーザー", "poor@example.com", 100)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"expensive-program", "/video/expensive.mp4", true, 500,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/programs/%d/purchase", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "insufficient points", resp["error"])

	// ポイントが減っていないことを確認
	var points int32
	err = dbConn.QueryRow(`SELECT points FROM "user" WHERE id = $1`, userID).Scan(&points)
	if err != nil {
		t.Fatalf("failed to query user points: %v", err)
	}
	assert.Equal(t, int32(100), points)
}

func TestPurchaseProgram_AlreadyPurchased_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "purchase-user-dup"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "重複購入ユーザー", "dup@example.com", 1000)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"dup-program", "/video/dup.mp4", true, 300,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	// 既に権限付与済み
	_, err = dbConn.Exec(`INSERT INTO permitted_program_users (user_id, program_id) VALUES ($1, $2)`, userID, programID)
	if err != nil {
		t.Fatalf("failed to insert permitted_program_users: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/programs/%d/purchase", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "already purchased", resp["error"])

	// ポイントが減っていないことを確認
	var points int32
	err = dbConn.QueryRow(`SELECT points FROM "user" WHERE id = $1`, userID).Scan(&points)
	if err != nil {
		t.Fatalf("failed to query user points: %v", err)
	}
	assert.Equal(t, int32(1000), points)
}

func TestPurchaseProgram_NotPurchasable_FreeProgram_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "purchase-user-free"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "無料番組ユーザー", "free@example.com", 1000)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// 非限定公開の番組（購入不可）
	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"free-program", "/video/free.mp4", false, 0,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/programs/%d/purchase", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "program is not purchasable", resp["error"])
}

func TestPurchaseProgram_NotPurchasable_LimitedButFree_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "purchase-user-limfree"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "限定無料ユーザー", "limfree@example.com", 1000)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// 限定公開だけど価格0（招待制）
	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"limited-free-program", "/video/limfree.mp4", true, 0,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/programs/%d/purchase", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "program is not purchasable", resp["error"])
}

func TestPurchaseProgram_NotFound_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "purchase-user-nf"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "存在しない番組ユーザー", "nf@example.com", 1000)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", "/programs/999999/purchase", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "program not found", resp["error"])
}

func TestPurchaseProgram_InvalidID_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth("some-user"))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", "/programs/abc/purchase", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "invalid id", resp["error"])
}

func TestPurchaseProgram_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	usersUC := usecase.NewUsersUsecase(q)
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, usersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth("")) // 未認証
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", "/programs/1/purchase", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "unauthorized", resp["error"])
}
