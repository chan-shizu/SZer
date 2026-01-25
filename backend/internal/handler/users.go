package handler

import (
	"encoding/json"
	"net/http"

	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type addPointsRequest struct {
	Amount int32 `json:"amount"`
}

func (h *Handler) AddPoints(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req addPointsRequest
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	points, err := h.users.AddPoints(c.Request.Context(), userID, req.Amount)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err == usecase.ErrInvalidPointsAmount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add points"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (h *Handler) GetPoints(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	points, err := h.users.GetPoints(c.Request.Context(), userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get points"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"points": points})
}
