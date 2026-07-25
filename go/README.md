Go 1.23

# バックエンド（Go 標準ライブラリ net/http）

## 基本的な使い方

- `go run .`      : サーバーを起動
- `go build`      : ビルド
- `go fmt ./...`  : コード整形
- `go vet ./...`  : 静的解析
- `go mod tidy`   : 依存関係を整理

## 起動

```bash
# 通常起動（デフォルト :8000）
make start

# ポート指定
make port=5555 start

# ホットリロード（air が必要）
make dev
```

## air（ホットリロード）のインストール

```bash
go install github.com/air-verse/air@latest
```

## 環境変数

| 変数          | 説明                          | デフォルト |
| ------------- | ----------------------------- | ---------- |
| `PORT`        | リッスンするポート            | `8000`     |
| `DATABASE_URL`| PostgreSQL 接続文字列         | -          |

# アーキテクチャ

```
infrastructure ──> interface(handler) ──> usecase ──> domain
       依存の向き ─────────────────────────────>
                （domain は誰にも依存しない）
```

**大原則: 依存は内側（domain）に向かってのみ流れる。**
外側（DB・HTTP）は内側を知ってよいが、内側は外側を知ってはいけない。

繋ぎ目のキモは「usecase が必要とする機能を usecase 側で interface として宣言し、
実装は infrastructure が提供する」こと（依存性逆転）。
具象の生成と注入（配線）は `main.go`（composition root）で行う。

## 各層の責務

| 層 | ディレクトリ | 責務 | 依存してよい相手 |
| ---- | ---- | ---- | ---- |
| domain | `domain/` | エンティティとビジネスルール | なし（最も内側） |
| usecase | `usecase/` | 操作の流れ。必要な依存を interface で宣言 | domain |
| interface | `interface/handler/` | HTTP ⇔ domain の変換、ルーティング | usecase, domain |
| infrastructure | `infrastructure/` | DB など技術詳細。usecase の interface を実装 | usecase, domain |

## 実装例（Match: 対戦結果）

- `domain/match.go` … `Match` エンティティと `Validate()` / `Winner()`
- `usecase/match_usecase.go` … `MatchRepository` interface（繋ぎ目）と `MatchUsecase`
- `infrastructure/match_repository.go` … `InMemoryMatchRepo`（後で Postgres 実装に差し替え可能）
- `interface/handler/match_handler.go` … `POST /matches`, `GET /matches`
- `main.go` … 上記を生成して注入する配線

### エンドポイント

```bash
# 対戦結果を登録
curl -X POST localhost:8000/matches \
  -d '{"id":"m1","player1":"alice","player2":"bob","score1":11,"score2":7}'

# 一覧取得
curl localhost:8000/matches
```