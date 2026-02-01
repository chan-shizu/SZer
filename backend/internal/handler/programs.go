package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"errors"

	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	programs *usecase.ProgramsUsecase
	users    *usecase.UsersUsecase
}

type upsertWatchHistoryRequest struct {
	ProgramID       int64 `json:"program_id"`
	PositionSeconds int32 `json:"position_seconds"`
	IsCompleted     bool  `json:"is_completed"`
}

func New(programs *usecase.ProgramsUsecase, users *usecase.UsersUsecase) *Handler {
	return &Handler{programs: programs, users: users}
}

func (h *Handler) ProgramDetails(c *gin.Context) {
	// ログインしていない場合はuserIDを空文字にする
	userID, _ := middleware.UserIDFromContext(c)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
 		log.Printf("[ProgramDetails] BadRequest: invalid id. userID=%s, idStr=%s, err=%v", userID, idStr, err)
 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
 		return
	}

	program, err := h.programs.GetProgramDetails(c.Request.Context(), userID, id)
	if err != nil {
 		if errors.Is(err, usecase.ErrProgramNotFound) {
 			log.Printf("[ProgramDetails] NotFound: program not found. userID=%s, id=%d", userID, id)
 			c.JSON(http.StatusNotFound, gin.H{"error": "program not found"})
 			return
 		}
 		log.Printf("[ProgramDetails] InternalServerError: failed to get program. userID=%s, id=%d, err=%v", userID, id, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get program"})
 		return
	}

	// 視聴回数インクリメント
	err = h.programs.IncrementViewCount(c.Request.Context(), id)
	if err != nil {
 		log.Printf("[ProgramDetails] InternalServerError: failed to increment view count. userID=%s, id=%d, err=%v", userID, id, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to increment view count"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{
		"program": program,
	})
}

func (h *Handler) LikeProgram(c *gin.Context) {
 	userID, err := middleware.UserIDFromContext(c)
 	if err != nil {
 		log.Printf("[LikeProgram] Unauthorized: userID取得失敗. err=%v", err)
 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
 		return
 	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
 		log.Printf("[LikeProgram] BadRequest: invalid id. userID=%s, idStr=%s, err=%v", userID, idStr, err)
 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
 		return
	}

	liked, likeCount, err := h.programs.LikeProgram(c.Request.Context(), userID, id)
	if err != nil {
 		if errors.Is(err, usecase.ErrProgramNotFound) {
 			log.Printf("[LikeProgram] NotFound: program not found. userID=%s, id=%d", userID, id)
			c.JSON(http.StatusNotFound, gin.H{"error": "program not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like program"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"liked":      liked,
		"like_count": likeCount,
	})
}

func (h *Handler) UnlikeProgram(c *gin.Context) {
 	userID, err := middleware.UserIDFromContext(c)
 	if err != nil {
 		log.Printf("[UnlikeProgram] 認証失敗: userID取得できず err=%v", err)
 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
 		return
 	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
 		log.Printf("[UnlikeProgram] 不正リクエスト: id変換失敗 userID=%s, idStr=%s, err=%v", userID, idStr, err)
 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
 		return
	}

	liked, likeCount, err := h.programs.UnlikeProgram(c.Request.Context(), userID, id)
	if err != nil {
 		if errors.Is(err, usecase.ErrProgramNotFound) {
 			log.Printf("[UnlikeProgram] データ未発見: program not found userID=%s, id=%d", userID, id)
 			c.JSON(http.StatusNotFound, gin.H{"error": "program not found"})
 			return
 		}
 		log.Printf("[UnlikeProgram] サーバーエラー: programのunlike失敗 userID=%s, id=%d, err=%v", userID, id, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike program"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{
		"liked":      liked,
		"like_count": likeCount,
	})
}

func (h *Handler) UpsertWatchHistory(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req upsertWatchHistoryRequest
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

 	if req.ProgramID == 0 {
 		log.Printf("[UpsertWatchHistory] BadRequest: program_id is required. userID=%s, req=%+v", userID, req)
 		c.JSON(http.StatusBadRequest, gin.H{"error": "program_id is required"})
 		return
 	}
 	if req.PositionSeconds < 0 {
 		log.Printf("[UpsertWatchHistory] BadRequest: position_seconds must be >= 0. userID=%s, req=%+v", userID, req)
 		c.JSON(http.StatusBadRequest, gin.H{"error": "position_seconds must be >= 0"})
 		return
 	}

	wh, err := h.programs.UpsertWatchHistory(c.Request.Context(), userID, req.ProgramID, req.PositionSeconds, req.IsCompleted)
	if err != nil {
 		log.Printf("[UpsertWatchHistory] InternalServerError: failed to upsert watch history. userID=%s, req=%+v, err=%v", userID, req, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert watch history"})
 		return
	}
	c.JSON(http.StatusOK, gin.H{"watch_history": wh})
}

func (h *Handler) ListPrograms(c *gin.Context) {
	title := c.Query("title")
	tagIDsStr := c.QueryArray("tag_ids")
	tagIDs := make([]int64, 0, len(tagIDsStr))
	for _, v := range tagIDsStr {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
 			log.Printf("[ListPrograms] 不正リクエスト: tag_ids変換失敗 tag_idsStr=%v, err=%v", tagIDsStr, err)
 			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag_ids"})
 			return
		}
		tagIDs = append(tagIDs, id)
	}

	programs, err := h.programs.ListPrograms(c.Request.Context(), title, tagIDs)
	if err != nil {
 		log.Printf("[ListPrograms] サーバーエラー: program一覧取得失敗 title=%s, tagIDs=%v, err=%v", title, tagIDs, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list programs"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{
		"programs": programs,
	})
}

func (h *Handler) Top(c *gin.Context) {
	programs, err := h.programs.ListTopPrograms(c.Request.Context())
	if err != nil {
 		log.Printf("[Top] サーバーエラー: トップprogram取得失敗 err=%v", err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get top programs"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{
		"programs": programs,
	})
}

func (h *Handler) TopLiked(c *gin.Context) {
	programs, err := h.programs.ListTopLikedPrograms(c.Request.Context())
	if err != nil {
 		log.Printf("[TopLiked] サーバーエラー: いいね多いprogram取得失敗 err=%v", err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get top liked programs"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{
		"programs": programs,
	})
}

func (h *Handler) TopViewed(c *gin.Context) {
	programs, err := h.programs.ListTopViewedPrograms(c.Request.Context())
	if err != nil {
 		log.Printf("[TopViewed] サーバーエラー: 視聴多いprogram取得失敗 err=%v", err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get top viewed programs"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{
		"programs": programs,
	})
}

func (h *Handler) ListWatchingPrograms(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
 		log.Printf("[ListWatchingPrograms] 認証失敗: userID取得できず err=%v", err)
 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
 		return
	}

	programs, err := h.programs.ListWatchingPrograms(c.Request.Context(), userID)
	if err != nil {
 		log.Printf("[ListWatchingPrograms] サーバーエラー: 視聴中program一覧取得失敗 userID=%s, err=%v", userID, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list watching programs"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{"programs": programs})
}

func (h *Handler) ListLikedPrograms(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
 		log.Printf("[ListLikedPrograms] 認証失敗: userID取得できず err=%v", err)
 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
 		return
	}

	programs, err := h.programs.ListLikedPrograms(c.Request.Context(), userID)
	if err != nil {
 		log.Printf("[ListLikedPrograms] サーバーエラー: いいねしたprogram一覧取得失敗 userID=%s, err=%v", userID, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list liked programs"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{"programs": programs})
}
