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

func TestProgramDetails_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	// テスト用データinsert
	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, thumbnail_path, description) VALUES ($1, $2, $3, $4) RETURNING id`,
		"detail-test-program", "/video/detail-test.mp4", "/thumbnail/detail-test.jpg", "detail-test-desc",
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(q)
	dummyUsersUC := &usecase.UsersUsecase{}
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, dummyUsersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth("")) // 未認証（user_idなし）
	r.GET("/programs/:id", h.ProgramDetails)

	// リクエスト実行
	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// ステータスコード確認
	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスボディ確認
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
	// 未認証なのでwatch_historyはnull
	assert.Nil(t, resp.Program["watch_history"])
}

func TestProgramDetails_WithTagsAndPerformers_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	// 番組insert
	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, description) VALUES ($1, $2, $3) RETURNING id`,
		"tagged-program", "/video/tagged.mp4", "tagged-desc",
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	// カテゴリタグinsert
	var tagID int64
	err = dbConn.QueryRow(
		`INSERT INTO category_tags (name) VALUES ($1) RETURNING id`,
		"test-tag",
	).Scan(&tagID)
	if err != nil {
		t.Fatalf("failed to insert test tag: %v", err)
	}

	// 番組×タグ紐付け
	_, err = dbConn.Exec(
		`INSERT INTO program_category_tags (program_id, tag_id) VALUES ($1, $2)`,
		programID, tagID,
	)
	if err != nil {
		t.Fatalf("failed to insert program_category_tags: %v", err)
	}

	// 出演者insert
	var performerID int64
	err = dbConn.QueryRow(
		`INSERT INTO performers (first_name, last_name, first_name_kana, last_name_kana) VALUES ($1, $2, $3, $4) RETURNING id`,
		"太郎", "田中", "タロウ", "タナカ",
	).Scan(&performerID)
	if err != nil {
		t.Fatalf("failed to insert test performer: %v", err)
	}

	// 番組×出演者紐付け
	_, err = dbConn.Exec(
		`INSERT INTO program_performers (program_id, performer_id) VALUES ($1, $2)`,
		programID, performerID,
	)
	if err != nil {
		t.Fatalf("failed to insert program_performers: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(q)
	dummyUsersUC := &usecase.UsersUsecase{}
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, dummyUsersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth("")) // 未認証（user_idなし）
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

	// カテゴリタグの確認
	assert.Len(t, resp.Program.CategoryTags, 1)
	assert.Equal(t, "test-tag", resp.Program.CategoryTags[0]["name"])

	// 出演者の確認
	assert.Len(t, resp.Program.Performers, 1)
	assert.Equal(t, "田中太郎", resp.Program.Performers[0]["full_name"])
	assert.Equal(t, "タナカタロウ", resp.Program.Performers[0]["full_name_kana"])
}

func TestProgramDetails_InvalidID_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(q)
	dummyUsersUC := &usecase.UsersUsecase{}
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, dummyUsersUC, dummyPayPayUC)
	r := gin.New()
	r.Use(MockOptionalAuth("")) // 未認証（user_idなし）
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
	_, q := setupTestDB(t)

	programsUC := usecase.NewProgramsUsecase(q)
	dummyUsersUC := &usecase.UsersUsecase{}
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, dummyUsersUC, dummyPayPayUC)
	r := gin.Default()
	r.Use(MockOptionalAuth("")) // 未認証（user_idなし）
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

	// 限定公開＋有料の番組insert
	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"limited-program", "/video/limited.mp4", true, 500,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(q)
	dummyUsersUC := &usecase.UsersUsecase{}
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, dummyUsersUC, dummyPayPayUC)
	r := gin.Default()
	r.Use(MockOptionalAuth("")) // 未認証（user_idなし）
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

	// view_count=0で番組insert
	var programID int64
	err := dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, view_count) VALUES ($1, $2, $3) RETURNING id`,
		"viewcount-test", "/video/vc.mp4", 0,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(q)
	dummyUsersUC := &usecase.UsersUsecase{}
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, dummyUsersUC, dummyPayPayUC)
	r := gin.Default()
	r.Use(MockOptionalAuth("")) // 未認証（user_idなし）
	r.GET("/programs/:id", h.ProgramDetails)

	// 1回目のリクエスト
	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2回目のリクエスト
	req2, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// DBで view_count を確認（2回アクセスしたので2になるはず）
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

	// テストユーザー
	userID := "test-user-1"
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified") VALUES ($1, $2, $3, true)`,
		userID, "テストユーザー", "test@example.com")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// 限定公開番組insert
	var programID int64
	err = dbConn.QueryRow(
		`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"permtest-limited", "/video/permtest.mp4", true, 100,
	).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	programsUC := usecase.NewProgramsUsecase(q)
	dummyUsersUC := &usecase.UsersUsecase{}
	dummyPayPayUC := &usecase.PayPayUsecase{}
	h := NewHandler(programsUC, dummyUsersUC, dummyPayPayUC)
	r := gin.Default()
	r.Use(MockOptionalAuth(userID)) // 認証ユーザーとしてuserIDをセット
	r.GET("/programs/:id", h.ProgramDetails)

	// --- 閲覧権限なしパターン ---
	req, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Program map[string]interface{} `json:"program"`
		IsPermitted bool `json:"is_permitted"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assert.Equal(t, false, resp.IsPermitted)
	assert.Equal(t, "", resp.Program["video_url"])

	// --- 閲覧権限ありパターン ---
	_, err = dbConn.Exec(`INSERT INTO permitted_program_users (user_id, program_id) VALUES ($1, $2)`, userID, programID)
	if err != nil {
		t.Fatalf("failed to insert permitted_program_users: %v", err)
	}

	// 認証ユーザーとしてアクセス（ヘッダーで認証情報をセット）
	reqAuth, _ := http.NewRequest("GET", fmt.Sprintf("/programs/%d", programID), nil)
	reqAuth.Header.Set("Authorization", "Bearer test-user-1") // テストユーザーIDをセット
	wAuth := httptest.NewRecorder()
	r.ServeHTTP(wAuth, reqAuth)
	err = json.Unmarshal(wAuth.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// 閲覧権限ありなら動画URLが返ることをassert
	assert.Equal(t, true, resp.IsPermitted)
	assert.NotEmpty(t, resp.Program["video_url"])
}
