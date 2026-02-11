-- userテーブルからpointsカラム削除
ALTER TABLE "user" DROP COLUMN IF EXISTS points;

-- paypay_topupsにprogram_idカラム追加（番組と紐づけ）
ALTER TABLE paypay_topups ADD COLUMN program_id BIGINT REFERENCES programs(id);
