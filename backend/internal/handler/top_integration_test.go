package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)


func TestTop_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbConn, q := setupTestDB(t)

	// テスト用データinsert
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

func TestTopPage_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// テスト用のルーター作成
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello SZer!")
	})

	// テスト用リクエスト
	req, _ := http.NewRequest("GET", "/", nil)
	w := performRequest(r, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "SZer")
}

// performRequestはテスト用のリクエスト実行ヘルパー
func performRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
