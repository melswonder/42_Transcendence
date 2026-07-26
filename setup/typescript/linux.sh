#!/usr/bin/env bash
# Node.js (nvm) + pnpm + 依存関係インストール — Linux / macOS 共通
# nvm の内部は set -u に対応していないため -u は付けない
set -eo pipefail

trap 'echo "❌ フロントエンドのセットアップに失敗しました (linux.sh:$LINENO)" >&2' ERR

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# nvm（未インストールなら導入）
if [ ! -s "$HOME/.nvm/nvm.sh" ]; then
  echo "⬇️  nvm をインストール中..."
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
fi

# シェルを再起動する代わりに読み込む
export NVM_DIR="$HOME/.nvm"
# nvm.sh は読み込みの最後に nvm_auto を実行する。default エイリアスが未設定だと
# ここが非ゼロ（3）で返るため、そのままだと set -e で無言のまま終了してしまう。
# shellcheck disable=SC1091
\. "$NVM_DIR/nvm.sh" || true

if ! command -v nvm >/dev/null 2>&1; then
  echo "❌ nvm を読み込めませんでした: $NVM_DIR/nvm.sh" >&2
  exit 1
fi

# Node.js 22（pnpm 11 が Node >= 22.13 を要求するため）
echo "⬇️  Node.js 22 をインストール中..."
nvm install 22
nvm use 22
# default を 22 にしておかないと、次に開くシェルは nvm のフォールバックで
# システムの古い node（42 のマシンでは v12）に戻ってしまう
nvm alias default 22

# pnpm（corepack 経由。バージョンは package.json の packageManager に従う）
corepack enable pnpm

# 依存関係インストール（絶対パスで typescript へ）
cd "$ROOT_DIR/typescript"
echo "⬇️  pnpm をインストール中（packageManager のピン留めを適用）..."
corepack install

# pnpm のストアはデフォルトだと「リポジトリがあるボリュームの直下」
# （42 のマシンでは /sgoinfre/.pnpm-store）に作られ、権限がなく失敗する。
# リポジトリ直下に置けば同一ファイルシステムなのでハードリンクも効く。
STORE_DIR="${PNPM_STORE_DIR:-$ROOT_DIR/.pnpm-store}"
mkdir -p "$STORE_DIR"

echo "⬇️  pnpm install（typescript/ / store: $STORE_DIR）..."
pnpm install --store-dir "$STORE_DIR"

echo "✅ フロントエンド セットアップ完了: node $(node -v) / pnpm $(pnpm -v)"
