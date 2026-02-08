package handler

import (
	"encoding/json"
	"fmt"
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
// GET /top
// =============================================================================

func TestTop_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	_, err := dbConn.Exec(`INSERT INTO programs (title, video_path, thumbnail_path, description) VALUES ($1, $2, $3, $4)`, "integration-title", "/video/test.mp4", "/thumbnail/test.jpg", "integration-desc")
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.GET("/top", h.Top)

	req, _ := http.NewRequest("GET", "/top", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Programs []map[string]interface{} `json:"programs"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.NotEmpty(t, resp.Programs)
	found := false
	for _, p := range resp.Programs {
		if p["title"] == "integration-title" {
			found = true
		}
	}
	assert.True(t, found, "inserted program should be in /top response")
}

// =============================================================================
// GET /top/liked
// =============================================================================

func TestTopLiked_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	// ユーザー作成
	userID := "topliked-user"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "TopLikedテストユーザー", "topliked@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// 番組2つ作成
	var programID1, programID2 int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"liked-program-1", "/video/liked1.mp4").Scan(&programID1)
	if err != nil {
		t.Fatalf("failed to insert program1: %v", err)
	}
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"liked-program-2", "/video/liked2.mp4").Scan(&programID2)
	if err != nil {
		t.Fatalf("failed to insert program2: %v", err)
	}

	// program1にだけいいね
	_, err = dbConn.Exec(`INSERT INTO likes (user_id, program_id) VALUES ($1, $2)`, userID, programID1)
	if err != nil {
		t.Fatalf("failed to insert like: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.GET("/top/liked", h.TopLiked)

	req, _ := http.NewRequest("GET", "/top/liked", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Programs []map[string]interface{} `json:"programs"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.NotEmpty(t, resp.Programs)
	// 最初の番組がいいねの多い方であることを確認
	assert.Equal(t, "liked-program-1", resp.Programs[0]["title"])
}

// =============================================================================
// GET /top/viewed
// =============================================================================

func TestTopViewed_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	// view_countが異なる番組を2つ作成
	_, err := dbConn.Exec(`INSERT INTO programs (title, video_path, view_count) VALUES ($1, $2, $3)`,
		"viewed-less", "/video/less.mp4", 5)
	if err != nil {
		t.Fatalf("failed to insert program: %v", err)
	}
	_, err = dbConn.Exec(`INSERT INTO programs (title, video_path, view_count) VALUES ($1, $2, $3)`,
		"viewed-more", "/video/more.mp4", 100)
	if err != nil {
		t.Fatalf("failed to insert program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.GET("/top/viewed", h.TopViewed)

	req, _ := http.NewRequest("GET", "/top/viewed", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Programs []map[string]interface{} `json:"programs"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.NotEmpty(t, resp.Programs)
	// 視聴数の多い方が先頭
	assert.Equal(t, "viewed-more", resp.Programs[0]["title"])
}

// =============================================================================
// GET /programs/:id (ProgramDetails)
// =============================================================================

func TestProgramDetails_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, thumbnail_path, description) VALUES ($1, $2, $3, $4) RETURNING id`,
		"detail-test-program", "/video/detail-test.mp4", "/thumbnail/detail-test.jpg", "detail-test-desc",
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.GET("/programs/:id", h.ProgramDetails)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Program map[string]interface{} `json:"program"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	assert.NotNil(t, resp.Program)
	assert.Equal(t, "detail-test-program", resp.Program["title"])
	assert.Equal(t, "detail-test-desc", resp.Program["description"])
	assert.NotNil(t, resp.Program["view_count"])
	assert.NotNil(t, resp.Program["like_count"])
	assert.Equal(t, false, resp.Program["liked"])
	assert.Equal(t, false, resp.Program["is_limited_release"])
	assert.Equal(t, float64(0), resp.Program["price"])
	assert.NotNil(t, resp.Program["category_tags"])
	assert.NotNil(t, resp.Program["performers"])
	assert.Nil(t, resp.Program["watch_history"])
}

func TestProgramDetails_WithTagsAndPerformers_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, description) VALUES ($1, $2, $3) RETURNING id`,
		"tagged-program", "/video/tagged.mp4", "tagged-desc",
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	var tagID int64
	err = dbConn.QueryRow(`INSERT INTO category_tags (name) VALUES ($1) RETURNING id`, "test-tag").Scan(&tagID)
	if err != nil {
		t.Fatalf("failed to insert test tag: %v", err)
	}
	_, err = dbConn.Exec(`INSERT INTO program_category_tags (program_id, tag_id) VALUES ($1, $2)`, programID, tagID)
	if err != nil {
		t.Fatalf("failed to insert program_category_tags: %v", err)
	}

	var performerID int64
	err = dbConn.QueryRow(
		`INSERT INTO performers (first_name, last_name, first_name_kana, last_name_kana) VALUES ($1, $2, $3, $4) RETURNING id`,
		"太郎", "田中", "タロウ", "タナカ",
	).Scan(&performerID)
	if err != nil {
		t.Fatalf("failed to insert test performer: %v", err)
	}
	_, err = dbConn.Exec(`INSERT INTO program_performers (program_id, performer_id) VALUES ($1, $2)`, programID, performerID)
	if err != nil {
		t.Fatalf("failed to insert program_performers: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.GET("/programs/:id", h.ProgramDetails)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Program struct {
			Title        string                   `json:"title"`
			CategoryTags []map[string]interface{} `json:"category_tags"`
			Performers   []map[string]interface{} `json:"performers"`
		} `json:"program"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	assert.Equal(t, "tagged-program", resp.Program.Title)
	assert.Len(t, resp.Program.CategoryTags, 1)
	assert.Equal(t, "test-tag", resp.Program.CategoryTags[0]["name"])
	assert.Len(t, resp.Program.Performers, 1)
	assert.Equal(t, "田中太郎", resp.Program.Performers[0]["full_name"])
	assert.Equal(t, "タナカタロウ", resp.Program.Performers[0]["full_name_kana"])
}

func TestProgramDetails_InvalidID_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.GET("/programs/:id", h.ProgramDetails)

	req, _ := http.NewRequest("GET", "/programs/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "invalid id", resp["error"])
}

func TestProgramDetails_NotFound_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.Use(MockOptionalAuth(""))
	r.GET("/programs/:id", h.ProgramDetails)

	req, _ := http.NewRequest("GET", "/programs/999999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "program not found", resp["error"])
}

func TestProgramDetails_LimitedReleaseAndPrice_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"limited-program", "/video/limited.mp4", true, 500,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.Use(MockOptionalAuth(""))
	r.GET("/programs/:id", h.ProgramDetails)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Program map[string]interface{} `json:"program"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Equal(t, "limited-program", resp.Program["title"])
	assert.Equal(t, true, resp.Program["is_limited_release"])
	assert.Equal(t, float64(500), resp.Program["price"])
}

func TestProgramDetails_ViewCountIncrement_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, view_count) VALUES ($1, $2, $3) RETURNING id`,
		"viewcount-test", "/video/vc.mp4", 0,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.Use(MockOptionalAuth(""))
	r.GET("/programs/:id", h.ProgramDetails)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req2, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var viewCount int
	err = dbConn.QueryRow(`SELECT view_count FROM programs WHERE id = $1`, programID).Scan(&viewCount)
	if err != nil {
		t.Fatalf("failed to query view_count: %v", err)
	}
	assert.Equal(t, 2, viewCount)
}

func TestProgramDetails_LimitedReleasePermission_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "test-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "テストユーザー", "test@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"permtest-limited", "/video/permtest.mp4", true, 100,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.Use(MockOptionalAuth(userID))
	r.GET("/programs/:id", h.ProgramDetails)

	// 閲覧権限なしパターン
	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Program     map[string]interface{} `json:"program"`
		IsPermitted bool                   `json:"is_permitted"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Equal(t, false, resp.IsPermitted)
	assert.Equal(t, "", resp.Program["video_url"])

	// 閲覧権限ありパターン
	_, err = dbConn.Exec(`INSERT INTO permitted_program_users (user_id, program_id) VALUES ($1, $2)`, userID, programID)
	if err != nil {
		t.Fatalf("failed to insert permitted_program_users: %v", err)
	}

	reqAuth, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	reqAuth.Header.Set("Authorization", "Bearer test-user-1")
	wAuth := httptest.NewRecorder()
	r.ServeHTTP(wAuth, reqAuth)
	err = json.Unmarshal(wAuth.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Equal(t, true, resp.IsPermitted)
	assert.NotEmpty(t, resp.Program["video_url"])
}

// =============================================================================
// GET /programs (ListPrograms)
// =============================================================================

func TestListPrograms_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	_, err := dbConn.Exec(`INSERT INTO programs (title, video_path) VALUES ($1, $2)`,
		"list-program-1", "/video/list1.mp4")
	if err != nil {
		t.Fatalf("failed to insert program: %v", err)
	}
	_, err = dbConn.Exec(`INSERT INTO programs (title, video_path) VALUES ($1, $2)`,
		"list-program-2", "/video/list2.mp4")
	if err != nil {
		t.Fatalf("failed to insert program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.GET("/programs", h.ListPrograms)

	req, _ := http.NewRequest("GET", "/programs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Programs []map[string]interface{} `json:"programs"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.GreaterOrEqual(t, len(resp.Programs), 2)
}

func TestListPrograms_WithTitleFilter_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	_, err := dbConn.Exec(`INSERT INTO programs (title, video_path) VALUES ($1, $2)`,
		"unique-title-xyz", "/video/unique.mp4")
	if err != nil {
		t.Fatalf("failed to insert program: %v", err)
	}
	_, err = dbConn.Exec(`INSERT INTO programs (title, video_path) VALUES ($1, $2)`,
		"other-program", "/video/other.mp4")
	if err != nil {
		t.Fatalf("failed to insert program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.Default()
	r.GET("/programs", h.ListPrograms)

	req, _ := http.NewRequest("GET", "/programs?title=unique-title", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Programs []map[string]interface{} `json:"programs"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Len(t, resp.Programs, 1)
	assert.Equal(t, "unique-title-xyz", resp.Programs[0]["title"])
}

// =============================================================================
// POST /programs/:id/purchase (PurchaseProgram)
// =============================================================================

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
	h := NewProgramsHandler(programsUC)
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
	assert.Equal(t, float64(500), resp.Points)

	// DB: 権限付与確認
	var permitted bool
	err = dbConn.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM permitted_program_users WHERE user_id = $1 AND program_id = $2)`,
		userID, programID,
	).Scan(&permitted)
	if err != nil {
		t.Fatalf("failed to query permitted_program_users: %v", err)
	}
	assert.True(t, permitted)

	// DB: ポイント減少確認
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
	h := NewProgramsHandler(programsUC)
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

	// ポイント未変更
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

	_, err = dbConn.Exec(`INSERT INTO permitted_program_users (user_id, program_id) VALUES ($1, $2)`, userID, programID)
	if err != nil {
		t.Fatalf("failed to insert permitted_program_users: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
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

	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"free-program", "/video/free.mp4", false, 0,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
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

	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"limited-free-program", "/video/limfree.mp4", true, 0,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
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
	h := NewProgramsHandler(programsUC)
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
	h := NewProgramsHandler(programsUC)
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
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.POST("/programs/:id/purchase", h.PurchaseProgram)

	req, _ := http.NewRequest("POST", "/programs/1/purchase", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "unauthorized", resp["error"])
}

// =============================================================================
// POST /programs/:id/like (LikeProgram)
// =============================================================================

func TestLikeProgram_Success_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "like-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "いいねユーザー", "like1@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"like-test-program", "/video/like.mp4").Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/like", h.LikeProgram)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/programs/%d/like", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Equal(t, true, resp["liked"])
	assert.Equal(t, float64(1), resp["like_count"])

	// DBでいいねが存在することを確認
	var exists bool
	err = dbConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND program_id = $2)`,
		userID, programID).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to query likes: %v", err)
	}
	assert.True(t, exists)
}

func TestLikeProgram_NotFound_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "like-user-nf"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "いいねNFユーザー", "likenf@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/programs/:id/like", h.LikeProgram)

	req, _ := http.NewRequest("POST", "/programs/999999/like", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLikeProgram_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.POST("/programs/:id/like", h.LikeProgram)

	req, _ := http.NewRequest("POST", "/programs/1/like", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// DELETE /programs/:id/like (UnlikeProgram)
// =============================================================================

func TestUnlikeProgram_Success_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "unlike-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "いいね解除ユーザー", "unlike1@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"unlike-test-program", "/video/unlike.mp4").Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	// 先にいいねしておく
	_, err = dbConn.Exec(`INSERT INTO likes (user_id, program_id) VALUES ($1, $2)`, userID, programID)
	if err != nil {
		t.Fatalf("failed to insert like: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.DELETE("/programs/:id/like", h.UnlikeProgram)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/programs/%d/like", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Equal(t, false, resp["liked"])
	assert.Equal(t, float64(0), resp["like_count"])

	// DBでいいねが消えていることを確認
	var exists bool
	err = dbConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND program_id = $2)`,
		userID, programID).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to query likes: %v", err)
	}
	assert.False(t, exists)
}

func TestUnlikeProgram_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.DELETE("/programs/:id/like", h.UnlikeProgram)

	req, _ := http.NewRequest("DELETE", "/programs/1/like", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// POST /watch-histories (UpsertWatchHistory)
// =============================================================================

func TestUpsertWatchHistory_Insert_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "wh-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "視聴履歴ユーザー", "wh1@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"wh-test-program", "/video/wh.mp4").Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/watch-histories", h.UpsertWatchHistory)

	body := fmt.Sprintf(`{"program_id":%d,"position_seconds":120,"is_completed":false}`, programID)
	req, _ := http.NewRequest("POST", "/watch-histories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		WatchHistory map[string]interface{} `json:"watch_history"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.NotNil(t, resp.WatchHistory)
	assert.Equal(t, float64(120), resp.WatchHistory["position_seconds"])
	assert.Equal(t, false, resp.WatchHistory["is_completed"])
}

func TestUpsertWatchHistory_Update_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "wh-user-update"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "視聴履歴更新ユーザー", "whupdate@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"wh-update-program", "/video/whupdate.mp4").Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	// 初回INSERT
	_, err = dbConn.Exec(`INSERT INTO watch_histories (user_id, program_id, position_seconds, is_completed) VALUES ($1, $2, $3, $4)`,
		userID, programID, 60, false)
	if err != nil {
		t.Fatalf("failed to insert watch_history: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.POST("/watch-histories", h.UpsertWatchHistory)

	// UPDATE: 位置を更新
	body := fmt.Sprintf(`{"program_id":%d,"position_seconds":300,"is_completed":false}`, programID)
	req, _ := http.NewRequest("POST", "/watch-histories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		WatchHistory map[string]interface{} `json:"watch_history"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Equal(t, float64(300), resp.WatchHistory["position_seconds"])
}

func TestUpsertWatchHistory_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.POST("/watch-histories", h.UpsertWatchHistory)

	body := `{"program_id":1,"position_seconds":60,"is_completed":false}`
	req, _ := http.NewRequest("POST", "/watch-histories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpsertWatchHistory_InvalidBody_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth("some-user"))
	r.POST("/watch-histories", h.UpsertWatchHistory)

	req, _ := http.NewRequest("POST", "/watch-histories", strings.NewReader("invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpsertWatchHistory_MissingProgramID_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth("some-user"))
	r.POST("/watch-histories", h.UpsertWatchHistory)

	body := `{"position_seconds":60,"is_completed":false}`
	req, _ := http.NewRequest("POST", "/watch-histories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "program_id is required", resp["error"])
}

// =============================================================================
// GET /me/watching-programs (ListWatchingPrograms)
// =============================================================================

func TestListWatchingPrograms_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "watching-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "視聴中ユーザー", "watching1@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"watching-program", "/video/watching.mp4").Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	// 視聴履歴（未完了）
	_, err = dbConn.Exec(`INSERT INTO watch_histories (user_id, program_id, position_seconds, is_completed) VALUES ($1, $2, $3, $4)`,
		userID, programID, 120, false)
	if err != nil {
		t.Fatalf("failed to insert watch_history: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.GET("/me/watching-programs", h.ListWatchingPrograms)

	req, _ := http.NewRequest("GET", "/me/watching-programs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Programs []map[string]interface{} `json:"programs"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Len(t, resp.Programs, 1)
	assert.Equal(t, "watching-program", resp.Programs[0]["title"])
}

func TestListWatchingPrograms_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.GET("/me/watching-programs", h.ListWatchingPrograms)

	req, _ := http.NewRequest("GET", "/me/watching-programs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// GET /me/liked-programs (ListLikedPrograms)
// =============================================================================

func TestListLikedPrograms_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	userID := "likedlist-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "いいね一覧ユーザー", "likedlist1@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var programID int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path) VALUES ($1, $2) RETURNING id`,
		"liked-list-program", "/video/likedlist.mp4").Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	_, err = dbConn.Exec(`INSERT INTO likes (user_id, program_id) VALUES ($1, $2)`, userID, programID)
	if err != nil {
		t.Fatalf("failed to insert like: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(userID))
	r.GET("/me/liked-programs", h.ListLikedPrograms)

	req, _ := http.NewRequest("GET", "/me/liked-programs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Programs []map[string]interface{} `json:"programs"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Len(t, resp.Programs, 1)
	assert.Equal(t, "liked-list-program", resp.Programs[0]["title"])
}

func TestListLikedPrograms_Unauthorized_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(dbConn, q)
	h := NewProgramsHandler(programsUC)
	r := gin.New()
	r.Use(MockOptionalAuth(""))
	r.GET("/me/liked-programs", h.ListLikedPrograms)

	req, _ := http.NewRequest("GET", "/me/liked-programs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
