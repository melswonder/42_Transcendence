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

# GORMの構造体からDDLを生成して標準出力へ（確認用）
schema:
	go run ./cmd/migrate

# 差分からmigrationを生成（atlas と Docker が必要。DBの事前起動は不要）
# 使い方: make migrate-diff name=add_users
migrate-diff:
	atlas migrate diff $(name) --env gorm

# migrationをDBへ適用
migrate-apply:
	atlas migrate apply --env gorm --url "$(DATABASE_URL)"

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

.PHONY: all start dev build schema migrate-diff migrate-apply fmt fmt-check lint vet ci
