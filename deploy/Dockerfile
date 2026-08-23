# 無料ホスティング（Render 等）用の全部入りイメージ。
# ビルドは リポジトリのルートを context にする:
#   docker build -f deploy/Dockerfile --build-arg NEXT_PUBLIC_API_URL=https://<app>.onrender.com/api .

# --- frontend -------------------------------------------------------------
FROM node:22-alpine AS fe
WORKDIR /app
RUN corepack enable pnpm
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN corepack install && pnpm install --frozen-lockfile
COPY frontend/ .
# NEXT_PUBLIC_* はビルド時にバンドルへ焼き込まれる
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL NEXT_STANDALONE=1
RUN pnpm build

# --- backend --------------------------------------------------------------
FROM golang:1.26-alpine AS be
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /out/serv ./cmd/serv

# --- runtime --------------------------------------------------------------
FROM node:22-alpine
# ルータ(caddy) と migration 用の atlas
ARG ATLAS_VERSION=v0.38.0
RUN apk add --no-cache caddy curl \
	&& case "$(uname -m)" in \
		x86_64) atlas_arch=amd64 ;; \
		aarch64) atlas_arch=arm64 ;; \
		*) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;; \
	esac \
	&& curl -fsSL -o /usr/local/bin/atlas \
		"https://release.ariga.io/atlas/atlas-linux-${atlas_arch}-${ATLAS_VERSION}" \
	&& chmod +x /usr/local/bin/atlas

WORKDIR /app
COPY --from=fe /app/.next/standalone ./web
COPY --from=fe /app/.next/static ./web/.next/static
COPY --from=fe /app/public ./web/public
COPY --from=be /out/serv /usr/local/bin/serv
COPY backend/migrations ./migrations
COPY deploy/Caddyfile /etc/caddy/Caddyfile
COPY deploy/start.sh /start.sh

ENV NODE_ENV=production MEDIA_DIR=/app/uploads
CMD ["/bin/sh", "/start.sh"]
