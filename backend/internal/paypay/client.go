package paypay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	MerchantID string
}

func LoadConfigFromEnv() (Config, error) {
	baseURL := strings.TrimSpace(os.Getenv("PAYPAY_API_URL"))
	apiKey := strings.TrimSpace(os.Getenv("PAYPAY_API_KEY"))
	apiSecret := strings.TrimSpace(os.Getenv("PAYPAY_API_SECRET"))
	merchantID := strings.TrimSpace(os.Getenv("PAYPAY_MERCHANT_ID"))

	if baseURL == "" {
		baseURL = "https://stg-api.sandbox.paypay.ne.jp"
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	if apiKey == "" || apiSecret == "" {
		return Config{}, errors.New("PAYPAY_API_KEY and PAYPAY_API_SECRET are required")
	}

	return Config{BaseURL: baseURL, APIKey: apiKey, APISecret: apiSecret, MerchantID: merchantID}, nil
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CreateCodeRequest struct {
	MerchantPaymentID string `json:"merchantPaymentId"`
	Amount            struct {
		Amount   int32  `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`
	OrderDescription string `json:"orderDescription,omitempty"`
	CodeType         string `json:"codeType,omitempty"`
	RedirectURL      string `json:"redirectUrl,omitempty"`
	RedirectType     string `json:"redirectType,omitempty"`
}

type ResultInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	CodeID  string `json:"codeId"`
}

type CreateCodeResponse struct {
	ResultInfo ResultInfo `json:"resultInfo"`
	Data       struct {
		CodeID            string `json:"codeId"`
		URL              string `json:"url"`
		Deeplink         string `json:"deeplink"`
		MerchantPaymentID string `json:"merchantPaymentId"`
	} `json:"data"`
}

type GetPaymentDetailsResponse struct {
	ResultInfo ResultInfo `json:"resultInfo"`
	Data       struct {
		Status            string `json:"status"`
		PaymentID         string `json:"paymentId"`
		MerchantPaymentID string `json:"merchantPaymentId"`
	} `json:"data"`
}

func (c *Client) CreateCode(ctx context.Context, req CreateCodeRequest) (CreateCodeResponse, error) {
	var res CreateCodeResponse
	b, err := json.Marshal(req)
	if err != nil {
		return res, err
	}

	err = c.doJSON(ctx, http.MethodPost, "/v2/codes", url.Values{}, "application/json", b, &res)
	return res, err
}

func (c *Client) GetPaymentDetails(ctx context.Context, merchantPaymentID string) (GetPaymentDetailsResponse, error) {
	var res GetPaymentDetailsResponse
	path := "/v2/codes/payments/" + url.PathEscape(merchantPaymentID)
	err := c.doJSON(ctx, http.MethodGet, path, url.Values{}, "application/json", nil, &res)
	return res, err
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, contentType string, body []byte, out any) error {
	fullURL := c.cfg.BaseURL + path
	if len(query) > 0 {
		fullURL = fullURL + "?" + query.Encode()
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")

	if strings.TrimSpace(c.cfg.MerchantID) != "" {
		req.Header.Set("X-ASSUME-MERCHANT", c.cfg.MerchantID)
	}
	
	pathWithQuery := u.EscapedPath()
	if u.RawQuery != "" {
		pathWithQuery = pathWithQuery + "?" + u.RawQuery
	}
	
	auth, err := buildAuthHeader(c.cfg.APIKey, c.cfg.APISecret, method, pathWithQuery, contentType, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// PayPay often returns JSON, but keep raw body for debugging.
		return fmt.Errorf("paypay api error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if out == nil {
		return nil
	}
	if len(respBody) == 0 {
		return errors.New("empty response body")
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return err
	}
	return nil
}

func buildAuthHeader(apiKey, apiSecret, method, pathWithQuery, contentType string, body []byte) (string, error) {
       nonce, err := randomHex(16)
       if err != nil {
	       return "", err
       }
       epoch := fmt.Sprintf("%d", time.Now().Unix())

       // PayPay公式仕様: body/content-typeが空の場合は"empty"をセット
       var hash, ct string
       if body == nil || len(body) == 0 {
	       hash = "empty"
	       ct = "empty"
       } else {
	       md := md5.New()
	       md.Write([]byte(contentType))
	       md.Write(body)
	       hash = base64.StdEncoding.EncodeToString(md.Sum(nil))
	       ct = contentType
       }

       // 署名文字列: {Request URI}\n{HTTPメソッド}\n{Nonce}\n{Epoch}\n{Content-Type}\n{Hash}
       signatureString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", pathWithQuery, method, nonce, epoch, ct, hash)

       mac := hmac.New(sha256.New, []byte(apiSecret))
       _, _ = mac.Write([]byte(signatureString))
       macData := base64.StdEncoding.EncodeToString(mac.Sum(nil))

       // Authorizationヘッダ: hmac OPA-Auth:{APIKey}:{macData}:{Nonce}:{Epoch}:{Hash}
       return fmt.Sprintf("hmac OPA-Auth:%s:%s:%s:%s:%s", apiKey, macData, nonce, epoch, hash), nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
