# SZer プロジェクト

動画配信プラットフォーム「SZer」の開発プロジェクト。
回答は日本語で行うこと。
口調はギャル風でお願いします。

## 技術スタック

- **Frontend**: Next.js 16 (App Router) / React 19 / TypeScript / Tailwind CSS v4 (`frontend/`)
- **Backend**: Go 1.25.5 / Gin / sqlc / Air (hot reload) (`backend/`)
- **DB**: PostgreSQL 16
- **Storage**: MinIO (S3互換)
- **認証**: Better Auth (Frontend側で管理、Backend側でトークン検証)
- **インフラ**: Docker Compose / 開発OS: Windows 11

## ディレクトリ構造

### Backend (`backend/`)

- `cmd/`: エントリポイント
- `internal/handler/`: HTTP ハンドラ (Gin)
- `internal/usecase/`: ビジネスロジック
- `internal/router/`: ルーティング定義
- `db/migrations/`: マイグレーションファイル (DDL)
- `db/queries/`: sqlc用クエリファイル (DML)

### Frontend (`frontend/`)

- `app/`: Next.js のページ・ルーティング定義
- `lib/api/`: バックエンドAPIクライアント
- `lib/auth/`: Better Auth 設定

## 開発環境

- Docker Compose で全サービスを管理: `docker compose up --build`
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
- コマンド実行は基本的にDockerコンテナ内で行う
- Windows環境ではPowerShellのコマンド記法に注意

## 開発コマンド

### Frontend

- 開発サーバー: `npm run dev`
- ビルド: `npm run build`
- Lint: `npm run lint`
- 認証DDL生成: `npm run auth:ddl`

### DB マイグレーション (makefileで管理)

- 全マイグレーション実行: `make migrate-up`
- 1ステップ実行: `make migrate-up1`
- 1ステップ戻す: `make migrate-down1`
- dirty時の再作成: `make migrate-init`
- シード: `make seed`

### sqlc コード生成

- `docker compose exec backend sqlc generate`
- クエリやスキーマを変更したら**必ず**実行すること

## コーディング規約

### Backend (Go)

- DB操作はGoコード内にSQLを書かず、`db/queries/*.sql` に記述して `sqlc generate` で生成
- API設計はRESTfulを意識
- エラーハンドリング: DBエラーやバリデーションエラーを適切にハンドリングし、Ginのレスポンスとして返す
- ファイル記述順: package宣言 → import (標準lib→外部lib) → 型定義 → 関数 (public→private) → ユーティリティ
- import宣言は1箇所のみ、重複禁止

### Frontend (Next.js)

- 基本は Server Components、インタラクションが必要な場合のみ `'use client'`
- API通信は `lib/api/` 配下の関数を使用。コンポーネントに直接 `fetch` を書かない
- スタイリングは Tailwind CSS のユーティリティクラスを使用
- すべての画面でログインが必須

### 共通

- 不要になった関数・変数は積極的に削除してコードを綺麗に保つ
- `.env` ファイルはコミットしない

## 行動指針

- エラーや不具合が発生した場合、自ら原因を調査し、ファイル検索・ログ確認・コマンド実行・修正案提示まで責任を持って進める
- 「調査してください」「確認してください」とユーザーに丸投げしない
- Goファイル編集時は構文エラー(import重複、package宣言の位置等)がないか必ず確認する
