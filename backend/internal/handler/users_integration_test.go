package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// GET /me/points (GetPoints)
// =============================================================================

func TestGetPoints_Success_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "getpoints-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "ポイント確認ユーザー", "getpoints@example.com", 500)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	usersUC := usecase.NewUsersUsecase(q)
	h := NewUsersHandler(usersUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.GET("/me/points", h.GetPoints)

	req, _ := http.NewRequest("GET", "/me/points", nil)
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
	assert.Equal(t, float64(500), resp.Points)
}

func TestGetPoints_NotFound_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, q := setupTestDB(t)

	usersUC := usecase.NewUsersUsecase(q)
	h := NewUsersHandler(usersUC)
	r := gin.New()
	r.Use(MockOptionalAuth("nonexistent-user"))
	r.GET("/me/points", h.GetPoints)

	req, _ := http.NewRequest("GET", "/me/points", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "user not found", resp["error"])
}

func TestGetPoints_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, q := setupTestDB(t)

	usersUC := usecase.NewUsersUsecase(q)
	h := NewUsersHandler(usersUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.GET("/me/points", h.GetPoints)

	req, _ := http.NewRequest("GET", "/me/points", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// POST /me/points/add (AddPoints)
// =============================================================================

func TestAddPoints_Success_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "addpoints-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "ポイント追加ユーザー", "addpoints@example.com", 200)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	usersUC := usecase.NewUsersUsecase(q)
	h := NewUsersHandler(usersUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/me/points/add", h.AddPoints)

	body := `{"amount":500}`
	req, _ := http.NewRequest("POST", "/me/points/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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
	// 200 + 500 = 700
	assert.Equal(t, float64(700), resp.Points)

	// DB確認
	var points int32
	err = dbConn.QueryRow(`SELECT points FROM "user" WHERE id = $1`, userID).Scan(&points)
	if err != nil {
		t.Fatalf("failed to query user points: %v", err)
	}
	assert.Equal(t, int32(700), points)
}

func TestAddPoints_InvalidAmount_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "addpoints-user-invalid"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points) VALUES ($1, $2, $3, true, $4)`,
		userID, "不正額ユーザー", "invalidamount@example.com", 200)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	usersUC := usecase.NewUsersUsecase(q)
	h := NewUsersHandler(usersUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/me/points/add", h.AddPoints)

	// 許可されていない金額（100, 500, 1000のみ）
	body := `{"amount":250}`
	req, _ := http.NewRequest("POST", "/me/points/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "invalid amount", resp["error"])
}

func TestAddPoints_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, q := setupTestDB(t)

	usersUC := usecase.NewUsersUsecase(q)
	h := NewUsersHandler(usersUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.POST("/me/points/add", h.AddPoints)

	body := `{"amount":100}`
	req, _ := http.NewRequest("POST", "/me/points/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddPoints_InvalidBody_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, q := setupTestDB(t)

	usersUC := usecase.NewUsersUsecase(q)
	h := NewUsersHandler(usersUC)
	r := gin.New()
	r.Use(MockOptionalAuth("some-user"))
	r.POST("/me/points/add", h.AddPoints)

	req, _ := http.NewRequest("POST", "/me/points/add", strings.NewReader("invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
