package handler

import (
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/chan-shizu/SZer/db"
	"github.com/chan-shizu/SZer/internal/paypay"
	"github.com/chan-shizu/SZer/internal/usecase"
	"github.com/gin-gonic/gin"
)

// PayPayWebhookHandler はPayPay Webhook受信用のハンドラだよ！
type PayPayWebhookHandler struct {
	Q  *db.Queries
	DB *sql.DB
}

func NewPayPayWebhookHandler(db *sql.DB, q *db.Queries) *PayPayWebhookHandler {
	return &PayPayWebhookHandler{Q: q, DB: db}
}

func (h *PayPayWebhookHandler) Handle(c *gin.Context) {
	// IPホワイトリストチェック（署名検証より前に）
	ipWhiteList := os.Getenv("PAYPAY_WEBHOOK_IP_WHITE_LIST")
	if ipWhiteList != "" {
		allowed := false
		remoteIP := c.ClientIP()
		for _, ip := range splitAndTrim(ipWhiteList, ",") {
			if remoteIP == ip {
				allowed = true
				break
			}
		}
		if !allowed {
			log.Printf("[PayPayWebhook] forbidden IP: %s", remoteIP)
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden IP"})
			return
		}
	}

	// PayPayからの署名ヘッダー取得
	signature := c.GetHeader("X-PayPay-Signature")
	if signature == "" {
		log.Printf("[PayPayWebhook] missing signature header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing signature header"})
		return
	}

	// リクエストBody取得
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[PayPayWebhook] failed to read body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	bodyStr := string(bodyBytes)

	// シークレット取得（環境変数から）
	secret := os.Getenv("PAYPAY_WEBHOOK_SECRET")
	if secret == "" {
		log.Printf("[PayPayWebhook] webhook secret not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook secret not set"})
		return
	}

	// 署名検証
	if !paypay.VerifyWebhookSignature(secret, bodyStr, signature) {
		log.Printf("[PayPayWebhook] invalid signature: header=%s body=%s", signature, bodyStr)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// イベントタイプ取得
	eventType := c.GetHeader("X-PayPay-Event-Type")
	if eventType == "" {
		// fallback: bodyからeventTypeをパースしてもOK
		eventType = "unknown"
	}

	// ビジネスロジック呼び出し
	err = usecase.PayPayWebhookEventHandler(c.Request.Context(), h.DB, h.Q, eventType, bodyBytes)
	if err != nil {
		log.Printf("[PayPayWebhook] event handling failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "event handling failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook受信OK"})
}

// --- 末尾にユーティリティ関数 ---
func splitAndTrim(s, sep string) []string {
	var out []string
	for _, v := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
