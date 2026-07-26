# バックエンド（Go 標準ライブラリ net/http）

Go 1.23 / 外部フレームワークなし。設計は **Clean Architecture**。

## 目次

- [起動する](#起動する)
- [アーキテクチャ](#アーキテクチャ)
- [今のコードを読んでみる（Greeting の例）](#今のコードを読んでみるgreeting-の例)
- [機能を追加する手順](#機能を追加する手順)
- [Lint / Format](#lint--format)
- [環境変数](#環境変数)

## 起動する

### Docker（ルートから）

```bash
make up          # 全サービス起動
make exec-go     # コンテナに入る
```

air によるホットリロードが効くので、`go/` 以下を編集すれば自動で再ビルドされる。

### ホストで直接

```bash
make start              # デフォルト :8000
make port=5555 start    # ポート指定
make dev                # ホットリロード（air が必要）
```

air のインストール:

```bash
go install github.com/air-verse/air@v1.61.7
```

> `@latest` は Go 1.26 以上を要求するため、このプロジェクト（Go 1.23）では v1.61.7 に固定する。

### 動作確認

```bash
curl localhost:8000
# {"message":"Hello from Go!"}
```

## アーキテクチャ

### 大原則: 依存は内側にだけ流れる

```
infrastructure ──> interface(handler) ──> usecase ──> domain
     (外側)                                            (内側)

     依存の向き ───────────────────────────────────>
```

外側（DB・HTTP といった技術の詳細）は内側を知ってよいが、
**内側は外側を絶対に知ってはいけない**。`domain` は誰にも依存しない。

なぜそうするのか:

- DB を PostgreSQL から別のものに変えても、`domain` と `usecase` は書き換えなくていい
- HTTP を WebSocket や gRPC に変えても、ビジネスロジックはそのまま使える
- `usecase` のテストで本物の DB を用意しなくていい（偽物の実装を差し込めばいい）

### 各層の責務

| 層             | ディレクトリ         | 責務                                       | 依存してよい相手   |
| -------------- | -------------------- | ------------------------------------------ | ------------------ |
| domain         | `domain/`            | エンティティとビジネスルール               | なし（最も内側）   |
| usecase        | `usecase/`           | 操作の流れ。必要な依存を interface で宣言  | domain             |
| interface      | `interface/handler/` | HTTP ⇔ domain の変換、ルーティング         | usecase, domain    |
| infrastructure | `infrastructure/`    | DB など技術詳細。usecase の interface を実装 | usecase, domain    |

### 繋ぎ目のキモ: 依存性逆転

矢印は内向きなのに、実際には `usecase` が DB を使いたい。この矛盾をどう解くか。

答えは「**usecase が必要とする機能を、usecase 側で interface として宣言する**」こと。

```
      usecase パッケージ
      ┌──────────────────────────────┐
      │ type GreetingRepository      │  ← 「こういう機能が欲しい」と宣言（抽象）
      │     interface { Get() ... }  │
      │                              │
      │ GreetingUsecase は上の       │
      │ interface だけを見て動く     │
      └──────────────────────────────┘
                    ▲
                    │ 実装する（依存の向きは外→内のまま）
      ┌──────────────────────────────┐
      │ infrastructure.GreetingRepo  │  ← 実際に DB を叩く（具象）
      └──────────────────────────────┘
```

`usecase` は「Get() できる何か」しか知らない。それが PostgreSQL なのかメモリなのかは知らないし、
知る必要もない。具象を作って注入する配線は `main.go`（composition root）だけが担当する。

## 今のコードを読んでみる（Greeting の例）

現状は `GET /` が JSON を返すだけだが、4層すべてが揃っている最小の見本になっている。

**1. `domain/greeting.go` — 何にも依存しない**

```go
type Greeting struct {
	Message string
}
```

**2. `usecase/greeting.go` — 欲しい機能を interface で宣言し、それだけを使う**

```go
// 繋ぎ目。実装は infrastructure が提供する。
type GreetingRepository interface {
	Get() domain.Greeting
}

type GreetingUsecase struct {
	repo GreetingRepository // 具象ではなく interface を持つ
}

func (u *GreetingUsecase) Hello() domain.Greeting {
	return u.repo.Get()
}
```

**3. `infrastructure/greeting.go` — interface を実装する**

```go
// コンパイル時に「ちゃんと実装できているか」を検査する定型句
var _ usecase.GreetingRepository = (*GreetingRepo)(nil)

func (r *GreetingRepo) Get() domain.Greeting {
	return domain.Greeting{Message: "Hello from Go!"}
}
```

**4. `interface/handler/greeting.go` — HTTP ⇔ domain の変換だけ**

処理そのものは書かず、`usecase` を呼んで結果を JSON にするだけ。

**5. `main.go` — 配線（composition root）**

```go
repo := infrastructure.NewGreetingRepo() // 具象を作り
uc := usecase.NewGreetingUsecase(repo)   // 内側へ注入していく
h := handler.NewGreetingHandler(uc)

mux.HandleFunc("/", h.Hello)
```

`main.go` は CORS ミドルウェアも持っている。許可オリジンは `allowedOrigins` で管理。

## 機能を追加する手順

例として「対戦結果（Match）」を足す場合、**内側から外側へ**作っていく。

**1. `domain/match.go` — エンティティとルール**

```go
package domain

type Match struct {
	ID      string
	Player1 string
	Player2 string
	Score1  int
	Score2  int
}

// ビジネスルールは domain に置く（HTTP も DB も出てこない）
func (m Match) Winner() string { ... }
func (m Match) Validate() error { ... }
```

**2. `usecase/match.go` — 流れと、必要な依存の宣言**

```go
package usecase

type MatchRepository interface {  // ← 繋ぎ目をここで宣言
	Save(m domain.Match) error
	FindAll() ([]domain.Match, error)
}

type MatchUsecase struct{ repo MatchRepository }
```

**3. `infrastructure/match.go` — 実装**

最初はメモリ実装で作り、後から PostgreSQL 実装に差し替えればいい。
`usecase` 側は一切変更が要らない。

**4. `interface/handler/match.go` — HTTP の入出力**

JSON をデコードして `usecase` に渡し、結果をエンコードして返すだけ。

**5. `main.go` に配線を追加**

```go
matchRepo := infrastructure.NewInMemoryMatchRepo()
matchUC := usecase.NewMatchUsecase(matchRepo)
matchHandler := handler.NewMatchHandler(matchUC)
mux.HandleFunc("/matches", matchHandler.Handle)
```

### レビューで見られるポイント

- `domain` が外側（`net/http`、DB ドライバ、`usecase`）を import していないか
- `usecase` が具象型ではなく interface に依存しているか
- error を握りつぶしていないか（`_ = doSomething()` になっていないか）
- goroutine のリーク・競合状態がないか

## Lint / Format

整形と静的解析は **golangci-lint v2** に集約している（設定は [.golangci.yml](.golangci.yml)）。

```bash
# インストール
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
```

| コマンド         | 内容                                              |
| ---------------- | ------------------------------------------------- |
| `make lint`      | 静的解析（CI と同じ）                             |
| `make fmt`       | 整形（gofmt + goimports）                         |
| `make fmt-check` | 整形漏れの検出（CI と同じ。差分を表示するだけ）   |
| `make ci`        | `build` + `lint` + `fmt-check` をまとめて         |
| `make vet`       | golangci-lint 無しで最低限だけ見たいとき          |
| `make build`     | `./tmp/main` にビルド                             |

有効にしている linter:

- **standard セット** … errcheck（error の無視）、govet、ineffassign、staticcheck、unused
- **bodyclose / sqlclosecheck** … HTTP レスポンスや DB rows の Close 漏れ
- **errorlint** … `errors.Is` / `errors.As` / `%w` の使い方
- **misspell** … スペルミス
- **revive** … 一般的なスタイル（doc コメント必須系は無効化済み）

1箇所だけ例外にしたいときは、その行の上に理由付きで書く:

```go
//nolint:errcheck // ベストエフォートなので失敗しても続行する
```

CI は [.github/workflows/go-lint-format.yml](../.github/workflows/go-lint-format.yml)。
`go/**` に変更がある PR / push で走るので、push 前に `make ci` を通しておくと落ちない。

## 環境変数

| 変数           | 説明                  | デフォルト |
| -------------- | --------------------- | ---------- |
| `PORT`         | リッスンするポート    | `8000`     |
| `DATABASE_URL` | PostgreSQL 接続文字列 | -          |

Docker で起動する場合はルートの `.env` から渡される。
