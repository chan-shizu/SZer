package handler

import (
	"net/http"
	"strconv"

	"errors"

	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	programs      *usecase.ProgramsUsecase
}

func New(programs *usecase.ProgramsUsecase) *Handler {
	return &Handler{programs: programs}
}

func (h *Handler) ProgramDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	program, err := h.programs.GetProgramDetails(c.Request.Context(), id)
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
