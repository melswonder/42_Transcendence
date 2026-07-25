#!/usr/bin/env bash
# Node.js (nvm) + pnpm + 依存関係インストール — Linux / macOS 共通
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# nvm（未インストールなら導入）
if [ ! -s "$HOME/.nvm/nvm.sh" ]; then
  echo "⬇️  nvm をインストール中..."
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
fi

# シェルを再起動する代わりに読み込む
export NVM_DIR="$HOME/.nvm"
# shellcheck disable=SC1091
\. "$NVM_DIR/nvm.sh"

# Node.js 22（pnpm 11 が Node >= 22.13 を要求するため）
echo "⬇️  Node.js 22 をインストール中..."
nvm install 22
nvm use 22

# pnpm（corepack 経由。バージョンは package.json の packageManager に従う）
corepack enable pnpm

# 依存関係インストール（絶対パスで typescript へ）
cd "$ROOT_DIR/typescript"
echo "⬇️  pnpm をインストール中（packageManager のピン留めを適用）..."
corepack install

echo "⬇️  pnpm install（typescript/）..."
pnpm install

echo "✅ フロントエンド セットアップ完了: node $(node -v) / pnpm $(pnpm -v)"
