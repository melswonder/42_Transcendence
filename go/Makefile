# デフォルトのポート番号
port?=4000

all: start

# 通常起動
start:
	PORT=$(port) go run .

# ホットリロード起動（air が必要: go install github.com/air-verse/air@latest）
dev:
	PORT=$(port) air -c .air.toml

# ビルド
build:
	go build -o ./tmp/main .

# 整形（golangci-lint が必要: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2）
fmt:
	golangci-lint fmt

# 整形漏れのチェック（CI と同じ）
fmt-check:
	golangci-lint fmt --diff

# 静的解析（CI と同じ。設定は .golangci.yml）
lint:
	golangci-lint run

# 静的解析（golangci-lint 無しで最低限だけ確認したいとき）
vet:
	go vet ./...

# CI と同じ内容をまとめて実行
ci: build lint fmt-check

.PHONY: all start dev build fmt fmt-check lint vet ci
