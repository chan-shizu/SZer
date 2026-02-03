package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chan-shizu/SZer/db"
	"github.com/chan-shizu/SZer/internal/paypay"
)

var ErrPayPayNotConfigured = errors.New("paypay not configured")
var ErrPayPayTopupNotFound = errors.New("paypay topup not found")

type PayPayUsecase struct {
	conn    *sql.DB
	q       *db.Queries
	client  *paypay.Client
	cfgErr  error
	baseURL string
}

func NewPayPayUsecase(conn *sql.DB, q *db.Queries) *PayPayUsecase {
	cfg, err := paypay.LoadConfigFromEnv()
	if err != nil {
		return &PayPayUsecase{conn: conn, q: q, cfgErr: err}
	}

	return &PayPayUsecase{conn: conn, q: q, client: paypay.NewClient(cfg)}
}

type PayPayCheckoutResult struct {
	MerchantPaymentID string
	URL               string
	Deeplink          string
}

func (p *PayPayUsecase) Checkout(ctx context.Context, userID string, amountYen int32, redirectBaseURL string) (PayPayCheckoutResult, error) {
	if p.client == nil {
		return PayPayCheckoutResult{}, fmt.Errorf("%w: %v", ErrPayPayNotConfigured, p.cfgErr)
	}
	if amountYen <= 0 {
		return PayPayCheckoutResult{}, ErrInvalidPointsAmount
	}
	
	// UIをシンプルにするため、当面は既存と同じ金額に制限
	switch amountYen {
	case 100, 500, 1000:
	default:
		return PayPayCheckoutResult{}, ErrInvalidPointsAmount
	}
	
	merchantPaymentID, err := paypay.RandomMerchantPaymentID()
	if err != nil {
		return PayPayCheckoutResult{}, err
	}
	
	_, err = p.q.CreatePayPayTopup(ctx, db.CreatePayPayTopupParams{
		UserID:            userID,
		MerchantPaymentID: merchantPaymentID,
		AmountYen:         amountYen,
	})
	if err != nil {
		return PayPayCheckoutResult{}, err
	}
	
	redirectURL := fmt.Sprintf("%s/mypage/points/paypay/return?merchantPaymentId=%s", redirectBaseURL, merchantPaymentID)

	var req paypay.CreateCodeRequest
	req.MerchantPaymentID = merchantPaymentID
	req.Amount.Amount = amountYen
	req.Amount.Currency = "JPY"
	req.OrderDescription = "SZer points"
	req.CodeType = "ORDER_QR"
	req.RedirectURL = redirectURL
	req.RedirectType = "WEB_LINK"
	
	resp, err := p.client.CreateCode(ctx, req)
	if err != nil {
		_ = p.q.UpdatePayPayTopupStatus(ctx, db.UpdatePayPayTopupStatusParams{
			UserID:            userID,
			MerchantPaymentID: merchantPaymentID,
			Status:            "FAILED",
			PaypayPaymentID:   sql.NullString{},
		})
		return PayPayCheckoutResult{}, err
	}

	codeID := resp.Data.CodeID
	if codeID != "" {
		_ = p.q.SetPayPayTopupCode(ctx, db.SetPayPayTopupCodeParams{
			UserID:            userID,
			MerchantPaymentID: merchantPaymentID,
			PaypayCodeID:      sql.NullString{String: codeID, Valid: true},
		})
	}

	return PayPayCheckoutResult{
		MerchantPaymentID: merchantPaymentID,
		URL:               resp.Data.URL,
		Deeplink:          resp.Data.Deeplink,
	}, nil
}

type PayPayConfirmResult struct {
	Status  string
	Points  int32
	Credited bool
}

func (p *PayPayUsecase) ConfirmAndCredit(ctx context.Context, userID, merchantPaymentID string) (PayPayConfirmResult, error) {
	if p.client == nil {
		return PayPayConfirmResult{}, fmt.Errorf("%w: %v", ErrPayPayNotConfigured, p.cfgErr)
	}
	if merchantPaymentID == "" {
		return PayPayConfirmResult{}, errors.New("merchantPaymentId is required")
	}

	payment, err := p.client.GetPaymentDetails(ctx, merchantPaymentID)
	if err != nil {
		return PayPayConfirmResult{}, err
	}

	status := payment.Data.Status
	paymentID := payment.Data.PaymentID

	tx, err := p.conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return PayPayConfirmResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := p.q.WithTx(tx)

	topup, err := qtx.GetPayPayTopupForUpdate(ctx, db.GetPayPayTopupForUpdateParams{
		UserID:            userID,
		MerchantPaymentID: merchantPaymentID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PayPayConfirmResult{}, ErrPayPayTopupNotFound
		}
		return PayPayConfirmResult{}, err
	}

	// まずステータスだけ更新（決済未完でも追跡できるように）
	var paypayPaymentID sql.NullString
	if paymentID != "" {
		paypayPaymentID = sql.NullString{String: paymentID, Valid: true}
	}
	_ = qtx.UpdatePayPayTopupStatus(ctx, db.UpdatePayPayTopupStatusParams{
		UserID:            userID,
		MerchantPaymentID: merchantPaymentID,
		Status:            status,
		PaypayPaymentID:   paypayPaymentID,
	})

	credited := false
	points := int32(0)

	if status == "COMPLETED" {
		affected, err := qtx.MarkPayPayTopupCredited(ctx, db.MarkPayPayTopupCreditedParams{
			UserID:            userID,
			MerchantPaymentID: merchantPaymentID,
			PaypayPaymentID:   paypayPaymentID,
		})
		if err != nil {
			return PayPayConfirmResult{}, err
		}

		if affected == 1 {
			credited = true
			points, err = qtx.AddPointsToUser(ctx, db.AddPointsToUserParams{ID: userID, Points: topup.AmountYen})
			if err != nil {
				return PayPayConfirmResult{}, err
			}
		} else {
			// すでに付与済み
			points, err = qtx.GetUserPoints(ctx, userID)
			if err != nil {
				return PayPayConfirmResult{}, err
			}
		}
	} else {
		points, err = qtx.GetUserPoints(ctx, userID)
		if err != nil {
			return PayPayConfirmResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return PayPayConfirmResult{}, err
	}

	return PayPayConfirmResult{Status: status, Points: points, Credited: credited}, nil
}
