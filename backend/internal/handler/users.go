package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chan-shizu/SZer/internal/middleware"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

type UsersHandler struct {
	users *usecase.UsersUsecase
}

func NewUsersHandler(users *usecase.UsersUsecase) *UsersHandler {
	return &UsersHandler{users: users}
}

type addPointsRequest struct {
	Amount int32 `json:"amount"`
}

func (h *UsersHandler) AddPoints(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
 		log.Printf("[AddPoints] 認証失敗: userID取得できず err=%v", err)
 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
 		return
	}

	var req addPointsRequest
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
 		log.Printf("[AddPoints] 不正リクエスト: bodyデコード失敗 userID=%s, err=%v", userID, err)
 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
 		return
	}

	points, err := h.users.AddPoints(c.Request.Context(), userID, req.Amount)
	if err != nil {
 		if err == usecase.ErrUserNotFound {
 			log.Printf("[AddPoints] データ未発見: user not found userID=%s", userID)
 			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
 			return
 		}
 		if err == usecase.ErrInvalidPointsAmount {
 			log.Printf("[AddPoints] 不正リクエスト: ポイント不正 userID=%s, amount=%d", userID, req.Amount)
 			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
 			return
 		}
 		log.Printf("[AddPoints] サーバーエラー: ポイント追加失敗 userID=%s, amount=%d, err=%v", userID, req.Amount, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add points"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (h *UsersHandler) GetPoints(c *gin.Context) {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
 		log.Printf("[GetPoints] 認証失敗: userID取得できず err=%v", err)
 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
 		return
	}

	points, err := h.users.GetPoints(c.Request.Context(), userID)
	if err != nil {
 		if err == usecase.ErrUserNotFound {
 			log.Printf("[GetPoints] データ未発見: user not found userID=%s", userID)
 			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
 			return
 		}
 		log.Printf("[GetPoints] サーバーエラー: ポイント取得失敗 userID=%s, err=%v", userID, err)
 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get points"})
 		return
	}

	c.JSON(http.StatusOK, gin.H{"points": points})
}
