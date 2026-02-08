-- PayPay topups (user purchases points via PayPay)

-- name: CreatePayPayTopup :one
INSERT INTO paypay_topups (
  user_id,
  merchant_payment_id,
  amount_yen,
  status
) VALUES (
  $1,
  $2,
  $3,
  'CREATED'
)
RETURNING *;

-- name: SetPayPayTopupCode :exec
UPDATE paypay_topups
SET paypay_code_id = $3,
    updated_at = now()
WHERE user_id = $1
  AND merchant_payment_id = $2;

-- name: GetPayPayTopupForUpdate :one
SELECT *
FROM paypay_topups
WHERE user_id = $1
  AND merchant_payment_id = $2
FOR UPDATE;

-- name: UpdatePayPayTopupStatus :exec
UPDATE paypay_topups
SET status = $3,
    paypay_payment_id = COALESCE($4, paypay_payment_id),
    updated_at = now()
WHERE user_id = $1
  AND merchant_payment_id = $2;

-- name: MarkPayPayTopupCredited :execrows
UPDATE paypay_topups
SET credited_at = now(),
    status = 'COMPLETED',
    paypay_payment_id = $3,
    updated_at = now()
WHERE user_id = $1
  AND merchant_payment_id = $2
  AND credited_at IS NULL;

-- Webhook用クエリ (merchant_payment_idのみで検索、userIdはWebhookに含まれないため)

-- name: GetPayPayTopupByMerchantPaymentIDForUpdate :one
SELECT *
FROM paypay_topups
WHERE merchant_payment_id = $1
FOR UPDATE;

-- name: UpdatePayPayTopupStatusByMerchantPaymentID :exec
UPDATE paypay_topups
SET status = $2,
    paypay_payment_id = COALESCE($3, paypay_payment_id),
    updated_at = now()
WHERE merchant_payment_id = $1;

-- name: MarkPayPayTopupCreditedByMerchantPaymentID :execrows
UPDATE paypay_topups
SET credited_at = now(),
    status = 'COMPLETED',
    paypay_payment_id = $2,
    updated_at = now()
WHERE merchant_payment_id = $1
  AND credited_at IS NULL;
