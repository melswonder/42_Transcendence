# 42_Transcendence

オンライン対戦型 **Quoridor（コリドール）** プラットフォーム。
Next.js（フロントエンド） + Go（バックエンド） + PostgreSQL の3コンテナ構成。

## 構成

```
42_Transcendence/
├── frontend/     フロントエンド : Next.js (App Router) + TailwindCSS  → :3000
│   └── types/models.ts           DB スキーマに対応する型定義
├── backend/      バックエンド   : Go (net/http)                        → :4000
│   ├── cmd/migrate/  GORM の構造体から DDL を生成（Atlas 用）
│   ├── atlas.hcl     マイグレーション設定
│   └── infrastructure/model.go   GORM の永続化モデル
├── db/postgres/  データベース   : PostgreSQL 16                        → :5432
├── docs/         設計ドキュメント（評価対象。Notion ではなくここが正本）
├── setup/        ローカル開発環境のセットアップスクリプト
├── docker-compose.yml
└── Makefile      docker compose のラッパー
```

各ディレクトリの詳細はそれぞれの README を参照:

- [frontend/README.md](frontend/README.md) — App Router、Server/Client Component、Feature-Sliced Design、型定義の同期、Tailwind、Lint/Format
- [backend/README.md](backend/README.md) — Clean Architecture の考え方、層の追加方法、DB マイグレーション、Lint/Format

## ドキュメント

設計資料は `docs/` に置く。**Notion は下書きであり、評価対象は Git リポジトリ内のこちら。**

| ドキュメント | 内容 |
| --- | --- |
| [docs/database-design.md](docs/database-design.md) | DB 設計の正本。全テーブル・制約・インデックス・トランザクション境界・Redis キー |
| [docs/FSD.md](docs/FSD.md) | フロントエンドの Feature-Sliced Design |

## 技術スタック

### 現在使っているもの

| 領域 | 技術 | 選定理由 |
| --- | --- | --- |
| フロントエンド | **Next.js** (App Router) + TypeScript | 課題が求めるフレームワーク要件を満たす。SSR とファイルベースルーティングでページ追加のコストが低い |
| スタイリング | **TailwindCSS** | 盤面 UI をレスポンシブに組むのに、独自 CSS を書かずクラスだけで完結できる |
| フロント設計 | **Feature-Sliced Design** | 層ごとの依存方向が決まっているため、機能追加で構造が壊れにくい（[docs/FSD.md](docs/FSD.md)） |
| バックエンド | **Go** + `net/http` | 標準ライブラリだけで HTTP と goroutine による並行処理が完結する。WebSocket の多重接続を捌く用途に向く |
| バックエンド設計 | **Clean Architecture** | ゲームルール（domain）を HTTP・DB・WebSocket から独立させ、単体テストできる状態に保つ |
| データベース | **PostgreSQL 16** | 後述 |
| ORM | **GORM** | Go で最も広く使われる ORM。構造体からスキーマを導出でき、`cmd/migrate` の仕組みがそのまま使える |
| マイグレーション | **Atlas** (`ariga.io/atlas`) | GORM の `AutoMigrate` と違い、**バージョン管理された SQL ファイル**として差分を残せる。ロールバック手順を持てる |
| 実行環境 | **Docker Compose** | `cp .env.example .env` の後、単一コマンドで全サービスが起動する |
| ホットリロード | **air**（Go） / Next.js dev server | `backend/` `frontend/` をコンテナにマウントしているため、編集が即反映される |
| Lint / Format | **golangci-lint v2**（Go） / **ESLint + Prettier**（TS） | CI と同じチェックをローカルで実行できる |

### なぜ PostgreSQL を選んだか

このプロジェクトは**アプリケーションのバグやレースコンディションを DB 制約で止める**方針を採っている（[docs/database-design.md](docs/database-design.md) §1）。そのために以下が必要で、いずれも PostgreSQL で使える。

- **partial unique index** — 「進行中の対局に同時参加できない」「退会済みユーザーは handle の一意性判定から外す」といった条件付き一意制約
- **CHECK 制約** — 盤面座標の範囲（0〜8）、`walls_remaining + walls_used = 10` のような複数列にまたがる不変条件
- **`citext`** — 大文字小文字を区別しないメールアドレスの一意性
- **`JSONB`** — `match_actions.payload` に操作内容を型を固定せず保存する
- **`TIMESTAMPTZ`** — タイムゾーン付きの時刻。タイムアウト判定の正本になる
- **トランザクションと行ロック** — `SELECT ... FOR UPDATE` で同一ターンの競合操作を直列化する

MySQL でも大半は代替できるが、partial unique index が無いため上記の二重参加防止をアプリ側検証に逃がすことになる。設計方針と合わないため PostgreSQL を選んだ。

### 今後追加する予定のもの

| 領域 | 技術 | 対応 Issue |
| --- | --- | --- |
| バックエンドフレームワーク | Gin（`net/http` から移行） | [#2](../../issues/2) |
| キャッシュ・presence・Pub/Sub | Redis | [#3](../../issues/3) |
| リアルタイム通信 | WebSocket（サーバー権威型） | [#8](../../issues/8) |
| リバースプロキシ / HTTPS・WSS | Caddy または Nginx | [#15](../../issues/15) |
| API ドキュメント | OpenAPI / Swagger | [#12](../../issues/12) |
| 多言語対応 | i18n（日本語 / 英語 / フランス語） | [#14](../../issues/14) |

## サービス間の関係

```
ブラウザ ──> frontend (:3000) ──> backend (:4000) ──> db (:5432)
              Next.js              Go net/http        PostgreSQL
```

ブラウザから直接叩く API の URL は `NEXT_PUBLIC_API_URL`（既定 `http://localhost:4000`）。
サーバー間通信ではコンテナ名（`http://backend:4000`）で解決できるが、
`NEXT_PUBLIC_` 付きの変数はブラウザで評価されるため `localhost` を使う点に注意。

## はじめかた

### 1. 環境変数

```bash
cp .env.example .env
```

| 変数                  | 用途                                       | 既定値                          |
| --------------------- | ------------------------------------------ | ------------------------------- |
| `POSTGRES_USER`       | DB ユーザー                                | `postgres`                      |
| `POSTGRES_PASSWORD`   | DB パスワード                              | `postgres`                      |
| `POSTGRES_DB`         | DB 名                                      | `transcendence`                 |
| `DATABASE_URL`        | バックエンドの接続文字列                   | `postgresql://postgres:...`     |
| `NEXT_PUBLIC_API_URL` | ブラウザから見たバックエンドの URL         | `http://localhost:4000`         |

### 2. Docker で全部起動（推奨）

```bash
make        # build + up
make logs   # ログを追う
make down   # 停止・削除
```

- http://localhost:3000 … フロントエンド
- http://localhost:4000 … バックエンド（`{"message":"Hello from Go!"}` が返る）

どちらもホットリロードが効く。`./backend` と `./frontend` がコンテナにマウントされていて、
Go は air、Next.js は dev サーバーがファイル変更を検知する。

### 3. ホストで直接動かす場合

Go / Node のツールチェーンをホストに入れる:

```bash
./setup/setup.sh              # 両方
./setup/setup.sh backend      # バックエンド（Go）だけ
./setup/setup.sh frontend     # フロントエンドだけ
```

## データベースとマイグレーション

スキーマの正本は [docs/database-design.md](docs/database-design.md)。それを GORM の構造体へ写したものが `backend/infrastructure/model.go`。

```bash
cd backend
make schema         # GORM の構造体から DDL を生成して表示（確認用）
make migrate-diff name=add_users   # 差分から migration ファイルを生成
make migrate-apply  # DB へ適用
```

**注意: 生成された DDL は完成形ではない。** partial unique index、`CREATE EXTENSION citext`、循環外部キーは GORM のタグでは表現できないため、手書きの SQL migration で補う必要がある。何を補うべきかは `backend/infrastructure/model.go` の各構造体のコメントに書いてある。

詳細は [backend/README.md#データベース--マイグレーション](backend/README.md#データベース--マイグレーション)。

## Makefile（ルート）

| コマンド                 | 内容                                                      |
| ------------------------ | --------------------------------------------------------- |
| `make` / `make all`      | build して起動                                            |
| `make build`             | イメージをビルド                                          |
| `make up` / `make down`  | 起動（バックグラウンド） / 停止・削除                     |
| `make stop` / `make start` | 一時停止 / 再開（コンテナは残す）                       |
| `make restart`           | down してから up                                          |
| `make logs`              | 全コンテナのログを追跡                                    |
| `make status`            | コンテナの状態                                            |
| `make clean`             | コンテナとイメージを削除（ボリュームは残す）              |
| `make fclean`            | イメージ・ボリューム・ネットワークまで削除（DB も消える） |
| `make re`                | fclean してから作り直し                                   |
| `make exec-backend`      | バックエンドのコンテナに入る                              |
| `make exec-frontend`     | フロントエンドのコンテナに入る                            |
| `make exec-postgres`     | DB のコンテナに入る                                       |

## CI

GitHub Actions が変更されたディレクトリに応じて走る。

| ワークフロー                                                             | 対象           | 内容                                                          |
| ------------------------------------------------------------------------ | -------------- | ------------------------------------------------------------- |
| [frontend-lint-format.yml](.github/workflows/frontend-lint-format.yml)     | `frontend/**`   | Prettier で整形チェック、ESLint で静的解析                    |
| [backend-lint-format.yml](.github/workflows/backend-lint-format.yml)       | `backend/**`    | ビルド、golangci-lint で静的解析、gofmt/goimports で整形チェック |
| [claude-auto-review.yaml](.github/workflows/claude-auto-review.yaml)       | PR 全体         | Claude による自動レビュー                                     |

push 前にローカルで CI と同じチェックを通せる:

```bash
cd backend && make ci                   # ビルド + lint + 整形チェック
cd frontend && pnpm format && pnpm lint
```

## 開発の流れ

1. `main` / `develop` から作業ブランチを切る
2. 実装する（`make logs` で動作を確認）
3. push 前にローカルで lint / format を通す（上記）
4. PR を出す → CI と Claude の自動レビューが走る
