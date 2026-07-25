#!/usr/bin/env bash
# Go のインストール（Homebrew） + air（ホットリロード）— macOS 用
set -euo pipefail

if command -v go >/dev/null 2>&1; then
  echo "✅ go は既にインストール済み: $(go version)"
else
  if ! command -v brew >/dev/null 2>&1; then
    echo "❌ Homebrew が見つかりません。https://brew.sh/ からインストールしてください"; exit 1
  fi
  echo "⬇️  brew install go..."
  brew install go
fi

# go install で入るバイナリ（air など）用に ~/go/bin を通す
export PATH="$PATH:$HOME/go/bin"
if ! grep -qsF '$HOME/go/bin' "$HOME/.zshrc" 2>/dev/null; then
  echo 'export PATH="$PATH:$HOME/go/bin"' >> "$HOME/.zshrc"
  echo "📝 ~/.zshrc に PATH を追記しました"
fi

# air（ホットリロード）
echo "⬇️  air をインストール中..."
go install github.com/air-verse/air@latest

echo "✅ Go セットアップ完了: $(go version)"
