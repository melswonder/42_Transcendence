# swaggo

Goのコメント（アノテーション）から OpenAPI 2.0 の定義を生成するツール。

## セットアップ

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

`$(go env GOPATH)/bin` に PATH を通しておくこと。

## このリポジトリでの使い方 backendディレクトリで

```bash
make swagger        # 定義を生成
make swagger-serve  # 生成してから確認用サーバーを起動
```

起動後 http://localhost:4000/swagger/index.html で Swagger UI が開く。

## コマンド仕様

### `swag init` — 定義を生成する

```bash
swag init -g cmd/swaggo/main.go -o docs/swagger
```

`swagger.json` / `swagger.yaml` / `docs.go` の3つを出力する。

主なオプション:

| オプション | 既定値 | 説明 |
|---|---|---|
| `-g`, `--generalInfo` | `main.go` | `@title` などの全体情報を書いたファイル |
| `-d`, `--dir` | `./` | アノテーションを走査するディレクトリ（カンマ区切り） |
| `-o`, `--output` | `./docs` | 出力先ディレクトリ |
| `--ot`, `--outputTypes` | `go,json,yaml` | 出力する形式 |
| `--parseDependency` | false | `go.mod` の依存パッケージの型も解決する |
| `--parseInternal` | false | `internal/` 配下も走査する |
| `--exclude` | - | 走査から除外するパス（カンマ区切り） |
| `--instanceName` | - | 複数の定義を同居させるときの識別名 |

`-g` は実質必須。既定値がリポジトリ直下の `main.go` なので、エントリポイントが別の場所にある場合は指定しないと空の定義ができる。

`-o` に指定したディレクトリ名がそのまま `docs.go` のパッケージ名になる（`-o docs/swagger` → `package swagger`）。変えたい場合のみ `--packageName` を使う。

### `swag fmt` — アノテーションを整形する

```bash
swag fmt -g cmd/swaggo/main.go
```

`@Summary` などの縦位置を揃えるだけ。定義は生成しない。

## 生成物を有効にする

`docs.go` の `init()` が定義を登録するので、**エントリポイントで blank import しないと UI が定義を読めない**（`doc.json` が404になる）。

```go
import (
	_ "transcendence-backend/docs/swagger"  // -o に指定したパスと一致させる
)
```

## アノテーション

全体情報は `main` 関数の直前に置く。

```go
// @title           Sample API
// @version         1.0
// @description     swaggoの動作確認用APIです。
// @host            localhost:4000
// @BasePath        /api/v1
func main() {
```

各エンドポイントはハンドラ関数の直前に置く。

```go
// @Summary      挨拶を取得
// @Description  Simple hello world endpoint
// @Tags         example
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /hello [get]
func HelloHandler(c *gin.Context) {
```

`@Router` のパスは `@BasePath` からの相対。上の例は実際には `/api/v1/hello` になる。

## 注意点

- `docs/swagger/` は生成物。手で編集しても次の `swag init` で消える
- `swag` CLI と `go.mod` の `github.com/swaggo/swag` はバージョンを揃える。ずれると生成された `docs.go` がビルドできない（例: CLI v1.16 が出力する `LeftDelim` はライブラリ v1.8 に存在しない）
