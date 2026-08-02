# バックエンド 言語Go

Go 1.24 / 外部フレームワークなし 設計は**Clean Architecture**。

## 起動する

```bash
# Docker（ルートから）
make up
make exec-go

# ホストで直接
make start              # :4000
make port=5555 start
make dev                # ホットリロード（air が必要）

# 動作確認
curl localhost:4000     # {"message":"Hello from Go!"}
```

air のインストール:

```bash
go install github.com/air-verse/air@v1.61.7
```

> `@latest` は Go 1.26 以上を要求するため、このプロジェクト（Go 1.24）では v1.61.7 に固定する。

## アーキテクチャ

### 大原則: 依存は内側にだけ流れる

```
外側    handler                    infrastructure
        HTTP の入出力               DB・外部サービス
           │                          │
           │ 呼ぶ                      │ interface を実装する
           ▼                          ▼
内側    usecase   操作の流れ + repository interface の宣言
           │
           ▼
内側    domain    エンティティとルール。何にも依存しない
```

外側（DB・HTTP といった技術の詳細）は内側を知ってよいが、
**内側は外側を絶対に知ってはいけない**。`domain` は誰にも依存しない。

`handler` と `infrastructure` は**互いに依存しない兄弟**で、どちらも外側にいる。
両者を繋いで組み立てるのは `main.go`（composition root）の仕事。

なぜそうするのか:

- DB を PostgreSQL から別のものに変えても、`domain` と `usecase` は書き換えなくていい
- HTTP を WebSocket や gRPC に変えても、ビジネスロジックはそのまま使える
- `usecase` のテストで本物の DB を用意しなくていい（偽物の実装を差し込めばいい）

> Clean Architecture の原典では handler の層を "interface adapters" と呼ぶが、
> `interface` は Go のキーワードで import パスも冗長になるため、ここでは `handler/` としている。

### 各層の責務

| 層             | ディレクトリ      | 責務                                         | 依存してよい相手 |
| -------------- | ----------------- | -------------------------------------------- | ---------------- |
| domain         | `domain/`         | エンティティとビジネスルール                 | なし（最も内側） |
| usecase        | `usecase/`        | 操作の流れ。必要な依存を interface で宣言    | domain           |
| handler        | `handler/`        | HTTP ⇔ domain の変換、ルーティング           | usecase, domain  |
| infrastructure | `infrastructure/` | DB など技術詳細。usecase の interface を実装 | usecase, domain  |

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

4層すべてが揃った最小の実例が `*/greeting.go`（`GET /` が JSON を返すだけ）。まずこれを読むとよい。

### 機能を追加する手順

**内側から外側へ**作っていく。

1. `domain/xxx.go` — エンティティとビジネスルール（HTTP も DB も出てこない）
2. `usecase/xxx.go` — 操作の流れ。必要な依存を `XxxRepository` interface として宣言する
3. `infrastructure/xxx.go` — 2 の interface を実装。最初はメモリ実装で作り、後から PostgreSQL 実装へ差し替えればいい（`usecase` 側の変更は不要）
4. `handler/xxx.go` — JSON をデコードして `usecase` に渡し、結果をエンコードして返すだけ
5. `main.go` — 具象を作って内側へ注入し、ルーティングに登録する

レビュー観点は [pull_request_template.md](../.github/pull_request_template.md) にある。

## データベース / マイグレーション

スキーマの正本は [docs/database-design.md](../docs/database-design.md)。そこから下流へこう流れる。

```
docs/database-design.md          設計の正本（人が読む）
        │  写す
        ▼
infrastructure/model.go          GORM の構造体
        │  go run ./cmd/migrate
        ▼
DDL（CREATE TABLE ...）          あるべきスキーマ
        │  atlas migrate diff
        ▼
migrations/*.sql                 バージョン管理された移行手順
        │  atlas migrate apply
        ▼
PostgreSQL
```

```bash
make schema                        # DDL を表示（DB 不要）
make migrate-diff name=add_users   # 差分から migrations/*.sql を生成
make migrate-apply                 # DB へ適用

curl -sSf https://atlasgo.sh | sh  # Atlas のインストール
```

押さえておくこと:

- **GORM の `AutoMigrate` は使わない。** 起動時にテーブルを勝手に書き換えるため記録が残らず、ロールバックもできない。代わりに Atlas で差分を SQL ファイルとして残す
- **`make schema` の出力は完成形ではない。** partial unique index・`CREATE EXTENSION citext`・複数列 CHECK・循環 FK は GORM のタグで表現できないため、`migrations/` に手書き SQL として足す。何を足すべきかは `infrastructure/model.go` の各構造体の doc コメントに列挙してある
- **GORM モデルは `domain` ではなく `infrastructure` に置く。** GORM のタグは DB の都合であり、`domain` に書くと「何にも依存しない」原則が崩れるため。`domain` のエンティティとは別物として扱い、変換は repository の責務
- `cmd/migrate` は DDL を標準出力へ吐くだけの部品で、単体では何も migrate しない。設定の詳細は [atlas.hcl](atlas.hcl) の冒頭コメントを参照

## 依存管理（go.mod / go.sum）

| ファイル | 役割 | 例え |
| --- | --- | --- |
| `go.mod` | モジュール名、必要な Go バージョン、依存パッケージとそのバージョンの**宣言** | 買い物リスト |
| `go.sum` | 実際にダウンロードした各モジュールの**ハッシュ台帳**。改ざん検知に使う | レシート＋封印シール |

**どちらも Git にコミットする。** これがあるおかげで、誰がいつビルドしても同じ依存が入る（再現性）。

`go.mod` の `// indirect` は「自分のコードが直接 import していない」という印。`gorm.io/gorm` が indirect にいるのは、コード中で `import "gorm.io/gorm"` を書かず GORM タグ（文字列）だけを使っているため。

| コマンド | 内容 |
| --- | --- |
| `go get X` | X を追加・更新 |
| `go mod tidy` | 実際の import と突き合わせて過不足を修正。**依存を足したら必ず実行する** |
| `go mod verify` | 手元のキャッシュが `go.sum` と一致するか検証 |

> `go.sum` が 1700 行を超えているのは、`atlas-provider-gorm` が MySQL / PostgreSQL / SQLite / SQL Server / Cloud Spanner の全ドライバを引き連れてくるため。これらは `cmd/migrate` のためだけで、アプリ本体は使わない。

## Lint / Format

整形と静的解析は **golangci-lint v2** に集約している（有効な linter は [.golangci.yml](.golangci.yml)）。

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
```

| コマンド         | 内容                                              |
| ---------------- | ------------------------------------------------- |
| `make lint`      | 静的解析（CI と同じ）                             |
| `make fmt`       | 整形（gofmt + goimports）                         |
| `make fmt-check` | 整形漏れの検出（差分表示のみ）                     |
| `make ci`        | `build` + `lint` + `fmt-check`                    |
| `make vet`       | golangci-lint 無しで最低限だけ見たいとき          |
| `make build`     | `./tmp/main` にビルド                             |

1箇所だけ例外にしたいときは、その行の上に理由付きで書く:

```go
//nolint:errcheck // ベストエフォートなので失敗しても続行する
```

CI は [.github/workflows/go-lint-format.yml](../.github/workflows/go-lint-format.yml)。
`go/**` に変更がある PR / push で走るので、push 前に `make ci` を通しておくと落ちない。

## 環境変数

| 変数           | 説明                  | デフォルト |
| -------------- | --------------------- | ---------- |
| `PORT`         | リッスンするポート    | `4000`     |
| `DATABASE_URL` | PostgreSQL 接続文字列 | -          |

Docker で起動する場合はルートの `.env` から渡される。
