#!/usr/bin/env bash
# Go のインストール（公式 tarball） + air（ホットリロード）— Linux 用
set -euo pipefail

GO_VERSION=1.23.4

# CPU アーキテクチャを判定
case "$(uname -m)" in
  x86_64)        GOARCH=amd64 ;;
  aarch64|arm64) GOARCH=arm64 ;;
  *) echo "❌ 未対応のアーキテクチャ: $(uname -m)"; exit 1 ;;
esac

if command -v go >/dev/null 2>&1; then
  echo "✅ go は既にインストール済み: $(go version)"
else
  TARBALL="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
  echo "⬇️  ${TARBALL} をダウンロード中..."
  curl -fLO "https://go.dev/dl/${TARBALL}"
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "${TARBALL}"
  rm -f "${TARBALL}"
fi

# 現在のシェルで PATH を通す
export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"

# ~/.bashrc に永続化（未追記なら）
if ! grep -qsF "/usr/local/go/bin" "$HOME/.bashrc" 2>/dev/null; then
  echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> "$HOME/.bashrc"
  echo "📝 ~/.bashrc に PATH を追記しました"
fi

# air（ホットリロード）
echo "⬇️  air をインストール中..."
go install github.com/air-verse/air@latest

echo "✅ Go セットアップ完了: $(go version)"
