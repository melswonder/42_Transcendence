FROM golang:1.23-alpine

WORKDIR /app

# ホットリロード用ツール air をインストール
# @latest は Go 1.26 以上を要求するため、Go 1.23 対応の最終版に固定する
RUN go install github.com/air-verse/air@v1.61.7

# 依存関係ファイルをコピー
COPY go.mod go.sum* ./

# 依存関係をダウンロード
RUN go mod download

# アプリケーションコードをコピー
COPY . .

# ポートを公開
EXPOSE 8000

# air でホットリロード起動（開発モード）
CMD ["air", "-c", ".air.toml"]
