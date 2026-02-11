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
var ErrNotPurchasable = errors.New("program is not purchasable")
var ErrAlreadyPurchased = errors.New("already purchased")

type PayPayUsecase struct {
	conn    *sql.DB
	q       *db.Queries
	client  *paypay.Client
	cfgErr  error
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

func (p *PayPayUsecase) Checkout(ctx context.Context, userID string, programID int64, redirectBaseURL string) (PayPayCheckoutResult, error) {
	if p.client == nil {
		return PayPayCheckoutResult{}, fmt.Errorf("%w: %v", ErrPayPayNotConfigured, p.cfgErr)
	}

	// 番組情報を取得
	program, err := p.q.GetProgramForPurchase(ctx, programID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PayPayCheckoutResult{}, ErrProgramNotFound
		}
		return PayPayCheckoutResult{}, err
	}

	// 購入可能かチェック
	if !program.IsLimitedRelease || program.Price <= 0 {
		return PayPayCheckoutResult{}, ErrNotPurchasable
	}

	// 既に購入済みかチェック
	permitted, err := p.q.IsUserPermittedForProgram(ctx, db.IsUserPermittedForProgramParams{
		UserID:    userID,
		ProgramID: programID,
	})
	if err != nil {
		return PayPayCheckoutResult{}, err
	}
	if permitted {
		return PayPayCheckoutResult{}, ErrAlreadyPurchased
	}

	merchantPaymentID, err := paypay.RandomMerchantPaymentID()
	if err != nil {
		return PayPayCheckoutResult{}, err
	}

	_, err = p.q.CreatePayPayTopup(ctx, db.CreatePayPayTopupParams{
		UserID:            userID,
		MerchantPaymentID: merchantPaymentID,
		AmountYen:         program.Price,
		ProgramID:         sql.NullInt64{Int64: programID, Valid: true},
	})
	if err != nil {
		return PayPayCheckoutResult{}, err
	}

	redirectURL := fmt.Sprintf("%s/programs/%d/paypay/return?merchantPaymentId=%s", redirectBaseURL, programID, merchantPaymentID)

	var req paypay.CreateCodeRequest
	req.MerchantPaymentID = merchantPaymentID
	req.Amount.Amount = program.Price
	req.Amount.Currency = "JPY"
	req.OrderDescription = "SZer program purchase"
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
	Status    string
	ProgramID int64
	Granted   bool
}

func (p *PayPayUsecase) ConfirmAndGrant(ctx context.Context, userID, merchantPaymentID string) (PayPayConfirmResult, error) {
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

	// ステータス更新
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

	granted := false
	programID := int64(0)
	if topup.ProgramID.Valid {
		programID = topup.ProgramID.Int64
	}

	if status == "COMPLETED" && topup.ProgramID.Valid {
		affected, err := qtx.MarkPayPayTopupCredited(ctx, db.MarkPayPayTopupCreditedParams{
			UserID:            userID,
			MerchantPaymentID: merchantPaymentID,
			PaypayPaymentID:   paypayPaymentID,
		})
		if err != nil {
			return PayPayConfirmResult{}, err
		}

		if affected == 1 {
			// 閲覧権限を付与
			err = qtx.AddPermittedProgramUser(ctx, db.AddPermittedProgramUserParams{
				UserID:    userID,
				ProgramID: topup.ProgramID.Int64,
			})
			if err != nil {
				return PayPayConfirmResult{}, err
			}
			granted = true
		}
	}

	if err := tx.Commit(); err != nil {
		return PayPayConfirmResult{}, err
	}

	return PayPayConfirmResult{Status: status, ProgramID: programID, Granted: granted}, nil
}
