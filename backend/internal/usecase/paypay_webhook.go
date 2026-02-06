package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chan-shizu/SZer/db"
)

// PayPayWebhookEventHandler はPayPay Webhookイベントごとの処理を行うギャル関数だよ！
// eventBodyはPayPay WebhookのJSONそのまま
func PayPayWebhookEventHandler(ctx context.Context, dbConn *sql.DB, q *db.Queries, eventType string, eventBody []byte) error {
	if eventType == "PAYMENT_COMPLETED" {
		// Webhookのbodyから必要な情報をパース
		var payload struct {
			MerchantPaymentID string `json:"merchantPaymentId"`
			UserID            string `json:"userId"`
			PaymentID         string `json:"paymentId"`
			Status            string `json:"status"`
			Amount            struct {
				Amount   int32  `json:"amount"`
				Currency string `json:"currency"`
			} `json:"amount"`
		}
		if err := json.Unmarshal(eventBody, &payload); err != nil {
			return fmt.Errorf("invalid webhook body: %w", err)
		}
		if payload.MerchantPaymentID == "" || payload.UserID == "" {
			return errors.New("missing merchantPaymentId or userId")
		}

		// DB接続取得（グローバルなdbConn/qを使う or DIする設計に後で修正可）

		// トランザクションで処理
		tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		qtx := q.WithTx(tx)

		// ステータス更新
		paypayPaymentID := sql.NullString{String: payload.PaymentID, Valid: payload.PaymentID != ""}
		_ = qtx.UpdatePayPayTopupStatus(ctx, db.UpdatePayPayTopupStatusParams{
			UserID:            payload.UserID,
			MerchantPaymentID: payload.MerchantPaymentID,
			Status:            payload.Status,
			PaypayPaymentID:   paypayPaymentID,
		})

		// ポイント付与
		if payload.Status == "COMPLETED" {
			affected, err := qtx.MarkPayPayTopupCredited(ctx, db.MarkPayPayTopupCreditedParams{
				UserID:            payload.UserID,
				MerchantPaymentID: payload.MerchantPaymentID,
				PaypayPaymentID:   paypayPaymentID,
			})
			if err != nil {
				return err
			}
			if affected == 1 {
				_, err = qtx.AddPointsToUser(ctx, db.AddPointsToUserParams{
					ID:     payload.UserID,
					Points: payload.Amount.Amount,
				})
				if err != nil {
					return err
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}
	// 他イベントは未対応
	return nil
}
