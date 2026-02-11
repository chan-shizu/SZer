package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chan-shizu/SZer/db"
)

// PayPayWebhookEventHandler はPayPay Webhookイベントごとの処理を行う
func PayPayWebhookEventHandler(ctx context.Context, dbConn *sql.DB, q *db.Queries, eventBody []byte) error {
	var payload struct {
		NotificationType string      `json:"notification_type"`
		MerchantID       string      `json:"merchant_id"`
		StoreID          string      `json:"store_id"`
		OrderID          string      `json:"order_id"`
		MerchantOrderID  string      `json:"merchant_order_id"`
		OrderAmount      json.Number `json:"order_amount"`
		State            string      `json:"state"`
		PaidAt           *string     `json:"paid_at"`
		AuthorizedAt     *string     `json:"authorized_at"`
		ExpiresAt        *string     `json:"expires_at"`
	}
	if err := json.Unmarshal(eventBody, &payload); err != nil {
		return fmt.Errorf("invalid webhook body: %w", err)
	}

	log.Printf("[PayPayWebhook] received: notification_type=%s, state=%s, merchant_order_id=%s, order_id=%s",
		payload.NotificationType, payload.State, payload.MerchantOrderID, payload.OrderID)

	if payload.NotificationType != "Transaction" {
		log.Printf("[PayPayWebhook] ignoring notification_type: %s", payload.NotificationType)
		return nil
	}

	if payload.MerchantOrderID == "" {
		return fmt.Errorf("missing merchant_order_id in webhook payload")
	}

	tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := q.WithTx(tx)

	topup, err := qtx.GetPayPayTopupByMerchantPaymentIDForUpdate(ctx, payload.MerchantOrderID)
	if err != nil {
		return fmt.Errorf("topup not found for merchant_order_id=%s: %w", payload.MerchantOrderID, err)
	}

	paypayPaymentID := sql.NullString{String: payload.OrderID, Valid: payload.OrderID != ""}
	_ = qtx.UpdatePayPayTopupStatusByMerchantPaymentID(ctx, db.UpdatePayPayTopupStatusByMerchantPaymentIDParams{
		MerchantPaymentID: payload.MerchantOrderID,
		Status:            payload.State,
		PaypayPaymentID:   paypayPaymentID,
	})

	// 閲覧権限付与 (COMPLETEDの場合)
	if payload.State == "COMPLETED" && topup.ProgramID.Valid {
		affected, err := qtx.MarkPayPayTopupCreditedByMerchantPaymentID(ctx, db.MarkPayPayTopupCreditedByMerchantPaymentIDParams{
			MerchantPaymentID: payload.MerchantOrderID,
			PaypayPaymentID:   paypayPaymentID,
		})
		if err != nil {
			return err
		}
		if affected == 1 {
			err = qtx.AddPermittedProgramUser(ctx, db.AddPermittedProgramUserParams{
				UserID:    topup.UserID,
				ProgramID: topup.ProgramID.Int64,
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
