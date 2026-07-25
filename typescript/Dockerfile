FROM node:22-alpine

WORKDIR /app

# pnpm を corepack 経由で有効化
RUN corepack enable pnpm

# 依存関係ファイルを先にコピー（レイヤーキャッシュを効かせるため）
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# package.json の packageManager でピン留めした pnpm を用意してから install
RUN corepack install && pnpm install --frozen-lockfile

# アプリケーションコードをコピー
COPY . .

# ポートを公開
EXPOSE 3000

# Next.js の開発サーバーを起動（ホットリロード）
CMD ["pnpm", "dev"]
