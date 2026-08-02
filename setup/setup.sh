#!/usr/bin/env bash
# ------------------------------------------------------------
# 42_Transcendence ローカル開発環境セットアップ
#   OS（macOS / Linux）を自動判定し、Go と TypeScript の
#   両方（またはどちらか一方）をセットアップする。
#
# 使い方:
#   ./setup.sh              # バックエンドとフロントエンド両方
#   ./setup.sh backend      # バックエンド（Go）だけ
#   ./setup.sh frontend     # フロントエンド（TypeScript）だけ
#   ./setup.sh -h           # ヘルプ
# ------------------------------------------------------------
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: ./setup.sh [target]

  target:
    all          バックエンドとフロントエンドの両方をセットアップ（デフォルト）
    backend      バックエンド（Go）のみ（go でも可）
    frontend     フロントエンド（TypeScript）のみ（ts / typescript でも可）
    -h, --help   このヘルプを表示
EOF
}

# --- OS 判定 ---
case "$(uname -s)" in
  Darwin) OS=macos ;;
  Linux)  OS=linux ;;
  *) echo "❌ 未対応のOSです: $(uname -s)"; exit 1 ;;
esac

TARGET="${1:-all}"

# --- .env を用意 ---
setup_env() {
  if [ -f "$ROOT_DIR/.env" ]; then
    echo "✅ .env は既に存在します"
  elif [ -f "$ROOT_DIR/.env.example" ]; then
    cp "$ROOT_DIR/.env.example" "$ROOT_DIR/.env"
    echo "✅ .env を .env.example から作成しました（必要に応じて編集してください）"
  else
    echo "⚠️  .env.example が見つかりません。スキップします"
  fi
}

run_backend() {
  echo "=============== バックエンド（Go）セットアップ ($OS) ==============="
  bash "$SCRIPT_DIR/backend/$OS.sh"
}

run_frontend() {
  echo "========= フロントエンド（TypeScript）セットアップ ($OS) ========="
  bash "$SCRIPT_DIR/frontend/$OS.sh"
}

echo "🖥  OS: $OS / 対象: $TARGET"

case "$TARGET" in
  -h|--help) usage; exit 0 ;;
  backend|go)
    setup_env
    run_backend
    ;;
  frontend|typescript|ts)
    setup_env
    run_frontend
    ;;
  all)
    setup_env
    run_backend
    run_frontend
    ;;
  *)
    echo "❌ 不明な引数: $TARGET"; echo; usage; exit 1 ;;
esac

echo
echo "🎉 セットアップ完了！"
echo "   新しいシェルを開くか、以下を実行して PATH を反映してください:"
# OS ではなくログインシェルで判定する（Linux でも zsh のことがある）
case "$(basename "${SHELL:-bash}")" in
  zsh) echo "     source ~/.zshrc" ;;
  *)   echo "     source ~/.bashrc" ;;
esac
echo "   Docker でまとめて起動する場合は、リポジトリ直下で 'make' を実行してください。"
