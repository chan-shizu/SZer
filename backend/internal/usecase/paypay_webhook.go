package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chan-shizu/SZer/db"
)

// PayPayWebhookEventHandler はPayPay Webhookイベントごとの処理を行うギャル関数だよ！
// eventBodyはPayPay WebhookのJSONそのまま
func PayPayWebhookEventHandler(ctx context.Context, dbConn *sql.DB, q *db.Queries, eventBody []byte) error {
	// PayPayの実際のWebhookペイロード構造
	var payload struct {
		NotificationType string `json:"notification_type"`
		MerchantID       string `json:"merchant_id"`
		StoreID          string `json:"store_id"`
		OrderID          string `json:"order_id"`          // PayPayの決済ID (= paymentId)
		MerchantOrderID  string `json:"merchant_order_id"` // 加盟店が設定した取引ID (= merchantPaymentId)
		OrderAmount      json.Number `json:"order_amount"`
		State            string `json:"state"`
		PaidAt           *string `json:"paid_at"`
		AuthorizedAt     *string `json:"authorized_at"`
		ExpiresAt        *string `json:"expires_at"`
	}
	if err := json.Unmarshal(eventBody, &payload); err != nil {
		return fmt.Errorf("invalid webhook body: %w", err)
	}

	log.Printf("[PayPayWebhook] received: notification_type=%s, state=%s, merchant_order_id=%s, order_id=%s",
		payload.NotificationType, payload.State, payload.MerchantOrderID, payload.OrderID)

	// Transactionイベントのみ処理
	if payload.NotificationType != "Transaction" {
		log.Printf("[PayPayWebhook] ignoring notification_type: %s", payload.NotificationType)
		return nil
	}

	if payload.MerchantOrderID == "" {
		return fmt.Errorf("missing merchant_order_id in webhook payload")
	}

	// トランザクションで処理
	tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := q.WithTx(tx)

	// merchant_payment_id (= merchant_order_id) でTopupレコードを取得
	topup, err := qtx.GetPayPayTopupByMerchantPaymentIDForUpdate(ctx, payload.MerchantOrderID)
	if err != nil {
		return fmt.Errorf("topup not found for merchant_order_id=%s: %w", payload.MerchantOrderID, err)
	}

	// ステータス更新
	paypayPaymentID := sql.NullString{String: payload.OrderID, Valid: payload.OrderID != ""}
	_ = qtx.UpdatePayPayTopupStatusByMerchantPaymentID(ctx, db.UpdatePayPayTopupStatusByMerchantPaymentIDParams{
		MerchantPaymentID: payload.MerchantOrderID,
		Status:            payload.State,
		PaypayPaymentID:   paypayPaymentID,
	})

	// ポイント付与 (COMPLETEDの場合)
	if payload.State == "COMPLETED" {
		affected, err := qtx.MarkPayPayTopupCreditedByMerchantPaymentID(ctx, db.MarkPayPayTopupCreditedByMerchantPaymentIDParams{
			MerchantPaymentID: payload.MerchantOrderID,
			PaypayPaymentID:   paypayPaymentID,
		})
		if err != nil {
			return err
		}
		if affected == 1 {
			// DBレコードのamount_yenを使用（Webhookの金額は改ざん防止のため使わない）
			_, err = qtx.AddPointsToUser(ctx, db.AddPointsToUserParams{
				ID:     topup.UserID,
				Points: topup.AmountYen,
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
