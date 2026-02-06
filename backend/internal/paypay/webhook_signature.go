package paypay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// VerifyWebhookSignature はPayPay Webhookの署名検証を行うギャル関数だよ！
func VerifyWebhookSignature(secret, body, signature string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	calculated := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(calculated), []byte(signature))
}
