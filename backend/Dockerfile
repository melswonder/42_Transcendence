# backend/go.mod の go ディレクティブと同じ 1.26 系にする。
# ここがズレると go.mod を満たすために toolchain を毎ビルド落としに行くことになる。
FROM golang:1.26-alpine

WORKDIR /app

# ホットリロード用ツール air をインストール
RUN go install github.com/air-verse/air@latest

# 依存関係ファイルをコピー
COPY go.mod go.sum* ./

# 依存関係をダウンロード
RUN go mod download

# アプリケーションコードをコピー
COPY . .

# ポートを公開
EXPOSE 4000

# air でホットリロード起動（開発モード）
CMD ["air", "-c", ".air.toml"]
