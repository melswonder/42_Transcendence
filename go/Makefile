# デフォルトのポート番号
port?=8000

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

# 整形
fmt:
	go fmt ./...

# 静的解析
vet:
	go vet ./...

.PHONY: all start dev build fmt vet
