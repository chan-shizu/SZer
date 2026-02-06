package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chan-shizu/SZer/db"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)


func TestPayPayWebhookHandler_Integration(t *testing.T) {

	gin.SetMode(gin.TestMode)
	os.Setenv("PAYPAY_WEBHOOK_SECRET", "testsecret")
	
	dbConn, q := setupTestDB(t)
	
	// テスト用ユーザーをinsert
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points, "createdAt", "updatedAt") VALUES ($1, $2, $3, $4, $5, now(), now()) ON CONFLICT (id) DO NOTHING`, "integration-user-id", "integration-user", "integration@example.com", true, 0)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	
	// テスト用topupをinsert
	_, err = dbConn.Exec(`INSERT INTO paypay_topups (user_id, merchant_payment_id, amount_yen, status, created_at, updated_at) VALUES ($1, $2, $3, $4, now(), now()) ON CONFLICT (merchant_payment_id) DO NOTHING`, "integration-user-id", "integration-merchant-id", 200, "CREATED")
	if err != nil {
		t.Fatalf("failed to insert test topup: %v", err)
	}
	
	handler := NewPayPayWebhookHandler(dbConn, q)
	r := gin.New()
	r.POST("/api/paypay/webhook", handler.Handle)
	
	body := `{"merchantPaymentId":"integration-merchant-id","userId":"integration-user-id","paymentId":"integration-payment-id","status":"COMPLETED","amount":{"amount":300,"currency":"JPY"}}`
	mac := hmac.New(sha256.New, []byte("testsecret"))
	mac.Write([]byte(body))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/paypay/webhook", bytes.NewBuffer([]byte(body)))
	req.Header.Set("X-PayPay-Signature", signature)
	req.Header.Set("X-PayPay-Event-Type", "PAYMENT_COMPLETED")
	// テスト用IP（ホワイトリスト内）をセット
	req.RemoteAddr = "13.112.237.64:12345"
	
	r.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	
	// DBの状態を検証（例: topupのstatus, userのpointsなど）
	topup, err := q.GetPayPayTopupForUpdate(req.Context(), db.GetPayPayTopupForUpdateParams{
		UserID:            "integration-user-id",
		MerchantPaymentID: "integration-merchant-id",
	})
	if err != nil {
		t.Fatalf("failed to get topup: %v", err)
	}
	if topup.Status != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %s", topup.Status)
	}

	// ユーザーのポイントも確認
	points, err := q.GetUserPoints(req.Context(), "integration-user-id")
	if err != nil {
		t.Fatalf("failed to get user points: %v", err)
	}
	if points != 100 {
		t.Errorf("expected points == 100, got %d", points)
	}
}

func TestPayPayWebhookHandler_ForbiddenIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("PAYPAY_WEBHOOK_SECRET", "testsecret")

	dbConn, q := setupTestDB(t)
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", points, "createdAt", "updatedAt") VALUES ($1, $2, $3, $4, $5, now(), now()) ON CONFLICT (id) DO NOTHING`, "forbidden-user-id", "forbidden-user", "forbidden@example.com", true, 0)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	_, err = dbConn.Exec(`INSERT INTO paypay_topups (user_id, merchant_payment_id, amount_yen, status, created_at, updated_at) VALUES ($1, $2, $3, $4, now(), now()) ON CONFLICT (merchant_payment_id) DO NOTHING`, "forbidden-user-id", "forbidden-merchant-id", 100, "CREATED")
	if err != nil {
		t.Fatalf("failed to insert test topup: %v", err)
	}

	handler := NewPayPayWebhookHandler(dbConn, q)
	r := gin.New()
	r.POST("/api/paypay/webhook", handler.Handle)

	body := `{"merchantPaymentId":"forbidden-merchant-id","userId":"forbidden-user-id","paymentId":"forbidden-payment-id","status":"COMPLETED","amount":{"amount":100,"currency":"JPY"}}`
	mac := hmac.New(sha256.New, []byte("testsecret"))
	mac.Write([]byte(body))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/paypay/webhook", bytes.NewBuffer([]byte(body)))
	req.Header.Set("X-PayPay-Signature", signature)
	req.Header.Set("X-PayPay-Event-Type", "PAYMENT_COMPLETED")
	// ホワイトリスト外のIP
	req.RemoteAddr = "1.2.3.4:12345"

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
