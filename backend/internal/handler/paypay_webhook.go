package handler

import (
	"database/sql"
	"io"
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
	// IPホワイトリストチェック（PayPay推奨のセキュリティ方式）
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

	// リクエストBody取得
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[PayPayWebhook] failed to read body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	bodyStr := string(bodyBytes)

	// 署名検証（ヘッダーがある場合のみ。PayPayの公式ドキュメントでは署名仕様が未記載のため任意）
	signature := c.GetHeader("X-PayPay-Signature")
	if signature != "" {
		secret := os.Getenv("PAYPAY_WEBHOOK_SECRET")
		if secret != "" {
			if !paypay.VerifyWebhookSignature(secret, bodyStr, signature) {
				log.Printf("[PayPayWebhook] invalid signature")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
				return
			}
		}
	}

	// ビジネスロジック呼び出し（イベントタイプはbody内のnotification_typeで判定）
	err = usecase.PayPayWebhookEventHandler(c.Request.Context(), h.DB, h.Q, bodyBytes)
	if err != nil {
		log.Printf("[PayPayWebhook] event handling failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "event handling failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
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
