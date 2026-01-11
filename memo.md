## migrate のコマンド

### 開発環境

ホスト（ローカル）から全マイグレーション適用:

```
migrate -path db/migrations -database "postgres://dev_user:dev_password@postgres:5432/szer_dev?sslmode=disable" up
```

1 つだけ適用（ステップ適用）:

```
migrate -path db/migrations -database "postgres://dev_user:dev_password@postgres:5432/szer_dev?sslmode=disable" up 1
```

最後の 1 つをロールバック（取り消し）:

```
migrate -path db/migrations -database "postgres://dev_user:dev_password@postgres:5432/szer_dev?sslmode=disable" up 1
```

全てダウン（全ロールバック）:

```
migrate -path db/migrations -database "postgres://dev_user:dev_password@postgres:5432/szer_dev?sslmode=disable" down
```

本番環境

```
railway run migrate -path backend/db/migrations -database "[接続のurl]" up
```
