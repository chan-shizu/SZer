package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"errors"

	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	programs      *usecase.ProgramsUsecase
}

type upsertWatchHistoryRequest struct {
	ProgramID       int64 `json:"program_id"`
	PositionSeconds int32 `json:"position_seconds"`
	IsCompleted     bool  `json:"is_completed"`
}

func New(programs *usecase.ProgramsUsecase) *Handler {
	return &Handler{programs: programs}
}

func (h *Handler) ProgramDetails(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	program, err := h.programs.GetProgramDetails(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, usecase.ErrProgramNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "program not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get program"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"program": program,
	})
}

func (h *Handler) LikeProgram(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	liked, likeCount, err := h.programs.LikeProgram(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, usecase.ErrProgramNotFound) {
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	liked, likeCount, err := h.programs.UnlikeProgram(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, usecase.ErrProgramNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "program not found"})
			return
		}
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
	if err != nil {
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

	if req.ProgramID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "program_id is required"})
		return
	}
	if req.PositionSeconds < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "position_seconds must be >= 0"})
		return
	}

	wh, err := h.programs.UpsertWatchHistory(c.Request.Context(), userID, req.ProgramID, req.PositionSeconds, req.IsCompleted)
	if err != nil {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag_ids"})
			return
		}
		tagIDs = append(tagIDs, id)
	}

	programs, err := h.programs.ListPrograms(c.Request.Context(), title, tagIDs)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get top programs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"programs": programs,
	})
}
