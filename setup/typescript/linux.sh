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

# Node.js 20
echo "⬇️  Node.js 20 をインストール中..."
nvm install 20
nvm use 20

# pnpm（corepack 経由）
corepack enable pnpm

# 依存関係インストール（絶対パスで typescript へ）
echo "⬇️  pnpm install（typescript/）..."
cd "$ROOT_DIR/typescript"
pnpm install

echo "✅ フロントエンド セットアップ完了: node $(node -v) / pnpm $(pnpm -v)"
