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
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt") VALUES ($1, $2, $3, $4, now(), now()) ON CONFLICT (id) DO NOTHING`, "integration-user-id", "integration-user", "integration@example.com", true)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// テスト用番組をinsert
	var programID int64
	err = dbConn.QueryRow(`INSERT INTO programs (title, video_path, is_limited_release, price) VALUES ($1, $2, $3, $4) RETURNING id`,
		"webhook-test-program", "/video/webhook.mp4", true, 100).Scan(&programID)
	if err != nil {
		t.Fatalf("failed to insert test program: %v", err)
	}

	// テスト用topupをinsert
	_, err = dbConn.Exec(`INSERT INTO paypay_topups (user_id, merchant_payment_id, amount_yen, status, program_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, now(), now()) ON CONFLICT (merchant_payment_id) DO NOTHING`, "integration-user-id", "integration-merchant-id", 100, "CREATED", programID)
	if err != nil {
		t.Fatalf("failed to insert test topup: %v", err)
	}

	handler := NewPayPayWebhookHandler(dbConn, q)
	r := gin.New()
	r.POST("/api/paypay/webhook", handler.Handle)

	// PayPayの実際のWebhookペイロード形式
	body := `{"notification_type":"Transaction","merchant_id":"test-merchant","order_id":"integration-payment-id","merchant_order_id":"integration-merchant-id","order_amount":"100","state":"COMPLETED","paid_at":"2026-02-08T12:00:00Z"}`
	mac := hmac.New(sha256.New, []byte("testsecret"))
	mac.Write([]byte(body))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/paypay/webhook", bytes.NewBuffer([]byte(body)))
	req.Header.Set("X-PayPay-Signature", signature)
	// テスト用IP（ホワイトリスト内）をセット
	req.RemoteAddr = "13.112.237.64:12345"

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// DBの状態を検証: topupのstatus
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

	// 閲覧権限が付与されていることを確認
	var permitted bool
	err = dbConn.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM permitted_program_users WHERE user_id = $1 AND program_id = $2)`,
		"integration-user-id", programID,
	).Scan(&permitted)
	if err != nil {
		t.Fatalf("failed to query permitted_program_users: %v", err)
	}
	if !permitted {
		t.Errorf("expected user to have viewing permission for program %d", programID)
	}
}

func TestPayPayWebhookHandler_ForbiddenIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("PAYPAY_WEBHOOK_SECRET", "testsecret")

	dbConn, q := setupTestDB(t)
	_, err := dbConn.Exec(`INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt") VALUES ($1, $2, $3, $4, now(), now()) ON CONFLICT (id) DO NOTHING`, "forbidden-user-id", "forbidden-user", "forbidden@example.com", true)
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

	// PayPayの実際のWebhookペイロード形式
	body := `{"notification_type":"Transaction","merchant_id":"test-merchant","order_id":"forbidden-payment-id","merchant_order_id":"forbidden-merchant-id","order_amount":"100","state":"COMPLETED","paid_at":"2026-02-08T12:00:00Z"}`
	mac := hmac.New(sha256.New, []byte("testsecret"))
	mac.Write([]byte(body))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/paypay/webhook", bytes.NewBuffer([]byte(body)))
	req.Header.Set("X-PayPay-Signature", signature)
	// ホワイトリスト外のIP
	req.RemoteAddr = "1.2.3.4:12345"

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
