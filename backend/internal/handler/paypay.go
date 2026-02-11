package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type PayPayHandler struct {
	paypay *usecase.PayPayUsecase
}

func NewPayPayHandler(paypay *usecase.PayPayUsecase) *PayPayHandler {
	return &PayPayHandler{paypay: paypay}
}

type payPayCheckoutRequest struct {
	ProgramID int64 `json:"program_id"`
}

func (h *PayPayHandler) PayPayCheckout(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req payPayCheckoutRequest
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

	redirectBase := strings.TrimRight(middleware.FrontendBaseURL(), "/")

	res, err := h.paypay.Checkout(c.Request.Context(), userID, req.ProgramID, redirectBase)
	if err != nil {
		if errors.Is(err, usecase.ErrProgramNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "program not found"})
			return
		}
		if errors.Is(err, usecase.ErrNotPurchasable) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "program is not purchasable"})
			return
		}
		if errors.Is(err, usecase.ErrAlreadyPurchased) {
			c.JSON(http.StatusConflict, gin.H{"error": "already purchased"})
			return
		}
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create paypay checkout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"merchant_payment_id": res.MerchantPaymentID,
		"url":                 res.URL,
		"deeplink":            res.Deeplink,
	})
}

func (h *PayPayHandler) PayPayGetPayment(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	merchantPaymentID := strings.TrimSpace(c.Param("merchantPaymentId"))
	result, err := h.paypay.ConfirmAndGrant(c.Request.Context(), userID, merchantPaymentID)
	if err != nil {
		if errors.Is(err, usecase.ErrPayPayTopupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm paypay payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     result.Status,
		"granted":    result.Granted,
		"program_id": result.ProgramID,
	})
}
