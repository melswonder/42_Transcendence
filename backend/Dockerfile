# backend/go.mod の go ディレクティブと同じ 1.26 系にする。
# ここがズレると go.mod を満たすために toolchain を毎ビルド落としに行くことになる。
FROM golang:1.26-alpine

WORKDIR /app

# ホットリロード用ツール air をインストール
RUN go install github.com/air-verse/air@latest

# migration を適用するために入れる
# alpine の apk には無いのでリリースバイナリを直接取得する。
# バージョンは開発者のローカル（make migrate-diff を叩く側）と揃えること。
ARG ATLAS_VERSION=v0.38.0
RUN apk add --no-cache curl \
	&& case "$(uname -m)" in \
		x86_64) atlas_arch=amd64 ;; \
		aarch64) atlas_arch=arm64 ;; \
		*) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;; \
	esac \
	&& curl -fsSL -o /usr/local/bin/atlas \
		"https://release.ariga.io/atlas/atlas-linux-${atlas_arch}-${ATLAS_VERSION}" \
	&& chmod +x /usr/local/bin/atlas

# 依存関係ファイルをコピー
COPY go.mod go.sum* ./

# 依存関係をダウンロード
RUN go mod download

# アプリケーションコードをコピー
COPY . .

# ポートを公開
EXPOSE 4000

# entrypoint は /app の外に置く。/app は docker-compose でホストの ./backend を
# マウントして上書きされるため、そこに置くとイメージ内の実体が隠れてしまう。
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["docker-entrypoint.sh"]

# air でホットリロード起動（開発モード）
CMD ["air", "-c", ".air.toml"]
