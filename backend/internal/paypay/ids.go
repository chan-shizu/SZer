package paypay

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomMerchantPaymentID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// 32 chars hex
	return hex.EncodeToString(b), nil
}
