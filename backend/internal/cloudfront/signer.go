package cloudfront

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	cfsign "github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
)

// VideoURLSigner はCloudFront署名付きURLを生成する
type VideoURLSigner struct {
	domain string
	signer *cfsign.URLSigner
}

// NewVideoURLSigner は環境変数から設定を読み込んでSignerを初期化する。
// CloudFront設定が不完全な場合はnilを返す（開発環境用フォールバック）。
func NewVideoURLSigner() (*VideoURLSigner, error) {
	domain := strings.TrimSpace(os.Getenv("CLOUDFRONT_DOMAIN"))
	keyPairID := strings.TrimSpace(os.Getenv("CLOUDFRONT_KEY_PAIR_ID"))
	privateKeyPEM := os.Getenv("CLOUD_FRONT_SECRET_KEY")

	if domain == "" || keyPairID == "" || privateKeyPEM == "" {
		return nil, nil
	}

	privKey, err := parseRSAPrivateKey([]byte(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CloudFront private key: %w", err)
	}

	signer := cfsign.NewURLSigner(keyPairID, privKey)
	log.Printf("[CloudFront] 署名付きURLモード: ON (domain=%s, keyPairID=%s)", domain, keyPairID)

	return &VideoURLSigner{
		domain: domain,
		signer: signer,
	}, nil
}

// SignURL は動画パスから署名付きCloudFront URLを生成する。
func (s *VideoURLSigner) SignURL(videoPath string, expiry time.Duration) (string, error) {
	path := strings.TrimLeft(videoPath, "/")
	rawURL := fmt.Sprintf("https://%s/video/%s", s.domain, path)

	expiresAt := time.Now().Add(expiry)
	signedURL, err := s.signer.Sign(rawURL, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to sign CloudFront URL: %w", err)
	}
	return signedURL, nil
}

func parseRSAPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no PEM block found in private key")
	}

	// PKCS#8形式を試す
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("parsed key is not RSA")
		}
		return rsaKey, nil
	}

	// PKCS#1形式も試す
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
