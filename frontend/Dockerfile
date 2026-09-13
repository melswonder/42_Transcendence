FROM node:22-alpine

WORKDIR /app

RUN corepack enable pnpm

# 依存関係ファイルを先にコピー（レイヤーキャッシュを効かせるため）
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# package.json の packageManager でピン留めした pnpm を用意してから install
RUN corepack install && pnpm install --frozen-lockfile

COPY . .

EXPOSE 3000

CMD ["pnpm", "dev"]
