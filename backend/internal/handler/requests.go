package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type RequestsHandler struct {
	requests *usecase.RequestsUsecase
}

func NewRequestsHandler(requests *usecase.RequestsUsecase) *RequestsHandler {
	return &RequestsHandler{requests: requests}
}

type createRequestBody struct {
	Content string `json:"content"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Note    string `json:"note"`
}

func (h *RequestsHandler) CreateRequest(c *gin.Context) {
	// OptionalAuthなのでエラーでも続行（未ログインの場合userIDは空文字）
	userID, _ := middleware.UserIDFromContext(c)

	var req createRequestBody
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	result, err := h.requests.CreateRequest(c.Request.Context(), userID, req.Content, req.Name, req.Contact, req.Note)
	if err != nil {
		if errors.Is(err, usecase.ErrRequestContentRequired) || errors.Is(err, usecase.ErrRequestNameRequired) || errors.Is(err, usecase.ErrRequestContactRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"request": result})
}
