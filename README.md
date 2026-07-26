# 42_Transcendence

Next.js（フロントエンド） + Go（バックエンド） + PostgreSQL の3コンテナ構成のプロジェクト。

## 構成

```
42_Transcendence/
├── typescript/   フロントエンド : Next.js (App Router) + TailwindCSS  → :3000
├── go/           バックエンド   : Go 標準ライブラリ (net/http)         → :8000
├── db/postgres/  データベース   : PostgreSQL 16                        → :5432
├── setup/        ローカル開発環境のセットアップスクリプト
├── docker-compose.yml
└── Makefile      docker compose のラッパー
```

各ディレクトリの詳細はそれぞれの README を参照:

- [typescript/README.md](typescript/README.md) — Next.js App Router の使い方、Tailwind、Lint/Format
- [go/README.md](go/README.md) — Clean Architecture の考え方、層の追加方法、Lint/Format

## サービス間の関係

```
ブラウザ ──> frontend (:3000) ──> backend (:8000) ──> db (:5432)
              Next.js              Go net/http        PostgreSQL
```

ブラウザから直接叩く API の URL は `NEXT_PUBLIC_API_URL`（既定 `http://localhost:8000`）。
サーバー間通信ではコンテナ名（`http://backend:8000`）で解決できるが、
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
| `NEXT_PUBLIC_API_URL` | ブラウザから見たバックエンドの URL         | `http://localhost:8000`         |

### 2. Docker で全部起動（推奨）

```bash
make        # build + up
make logs   # ログを追う
make down   # 停止・削除
```

- http://localhost:3000 … フロントエンド
- http://localhost:8000 … バックエンド（`{"message":"Hello from Go!"}` が返る）

どちらもホットリロードが効く。`./go` と `./typescript` がコンテナにマウントされていて、
Go は air、Next.js は dev サーバーがファイル変更を検知する。

### 3. ホストで直接動かす場合

Go / Node のツールチェーンをホストに入れる:

```bash
./setup/setup.sh              # 両方
./setup/setup.sh go           # Go だけ
./setup/setup.sh typescript   # フロントだけ
```

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
| `make exec-go`           | バックエンドのコンテナに入る                              |
| `make exec-typescript`   | フロントエンドのコンテナに入る                            |
| `make exec-postgres`     | DB のコンテナに入る                                       |

## CI

GitHub Actions が変更されたディレクトリに応じて走る。

| ワークフロー                                                             | 対象           | 内容                                                          |
| ------------------------------------------------------------------------ | -------------- | ------------------------------------------------------------- |
| [lint-format-check.yml](.github/workflows/lint-format-check.yml)           | `typescript/**` | Prettier で整形チェック、ESLint で静的解析                    |
| [go-lint-format.yml](.github/workflows/go-lint-format.yml)                 | `go/**`         | ビルド、golangci-lint で静的解析、gofmt/goimports で整形チェック |
| [claude-auto-review.yaml](.github/workflows/claude-auto-review.yaml)       | PR 全体         | Claude による自動レビュー                                     |

push 前にローカルで CI と同じチェックを通せる:

```bash
cd go && make ci                        # ビルド + lint + 整形チェック
cd typescript && pnpm format && pnpm lint
```

## 開発の流れ

1. `main` / `develop` から作業ブランチを切る
2. 実装する（`make logs` で動作を確認）
3. push 前にローカルで lint / format を通す（上記）
4. PR を出す → CI と Claude の自動レビューが走る
