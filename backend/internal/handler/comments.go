package handler

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/chan-shizu/SZer/db"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type CommentsHandler struct {
	uc *usecase.CommentsUsecase
}

func NewCommentsHandler(q *db.Queries) *CommentsHandler {
	return &CommentsHandler{uc: usecase.NewCommentsUsecase(q)}
}

// GET /programs/:id/comments
func (h *CommentsHandler) ListComments(c *gin.Context) {
	programID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("[コメント一覧取得] programIDパースエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid program id"})
		return
	}
	comments, err := h.uc.ListCommentsByProgramID(c, programID)
	if err != nil {
		log.Printf("[コメント一覧取得] programID=%d err=%v", programID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get comments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// POST /programs/:id/comments
func (h *CommentsHandler) PostComment(c *gin.Context) {
	programID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("[コメント投稿] programIDパースエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid program id"})
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		log.Printf("[コメント投稿] JSONバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content"})
		return
	}
	       var userID string
	       if v, ok := c.Get("user_id"); ok {
		       if s, ok := v.(string); ok {
			       userID = s
		       }
	       }
	       comment, err := h.uc.PostComment(c, programID, userID, req.Content)
	       if err != nil {
		       log.Printf("[コメント投稿:usecase失敗] programID=%d userID=%s content=%s err=%v", programID, userID, req.Content, err)
		       c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		       return
	       }
	       c.JSON(http.StatusOK, gin.H{"comment": comment})
}

func sqlNullString(s string) sql.NullString {
       if s == "" {
	       return sql.NullString{Valid: false}
       }
       return sql.NullString{String: s, Valid: true}
}
