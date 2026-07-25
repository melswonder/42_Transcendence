#!/usr/bin/env bash
# Go のインストール（Homebrew） + air（ホットリロード）— macOS 用
set -euo pipefail

GO_REQUIRED_MINOR=23   # go/go.mod の go ディレクティブに合わせる（1.x の x）

# 既存の go が go.mod の要求を満たすか判定する
go_version_ok() {
  command -v go >/dev/null 2>&1 || return 1
  local raw major minor
  raw="$(go version 2>/dev/null | awk '{print $3}')"   # 例: go1.18.1
  raw="${raw#go}"
  major="${raw%%.*}"
  minor="${raw#*.}"; minor="${minor%%.*}"
  [ -n "$major" ] && [ -n "$minor" ] || return 1
  [ "$major" -gt 1 ] && return 0
  [ "$major" -eq 1 ] && [ "$minor" -ge "$GO_REQUIRED_MINOR" ]
}

if go_version_ok; then
  echo "✅ go は既にインストール済み: $(go version)"
else
  if ! command -v brew >/dev/null 2>&1; then
    echo "❌ Homebrew が見つかりません。https://brew.sh/ からインストールしてください"; exit 1
  fi
  if command -v go >/dev/null 2>&1; then
    echo "⚠️  既存の go が古いため更新します: $(go version)（要 go1.${GO_REQUIRED_MINOR} 以上）"
    brew upgrade go || brew install go
  else
    echo "⬇️  brew install go..."
    brew install go
  fi
  go_version_ok || { echo "❌ go1.${GO_REQUIRED_MINOR} 以上にできませんでした: $(go version)"; exit 1; }
fi

# go install で入るバイナリ（air など）用に GOPATH/bin を通す
GO_PATH_BIN="$(go env GOPATH)/bin"
export PATH="$PATH:$GO_PATH_BIN"
if ! grep -qsF 'go env GOPATH' "$HOME/.zshrc" 2>/dev/null; then
  echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> "$HOME/.zshrc"
  echo "📝 ~/.zshrc に PATH を追記しました"
fi

# air（ホットリロード）
# @latest は Go 1.26 以上を要求するため、Go 1.23 対応の最終版に固定（go/Dockerfile と揃える）
echo "⬇️  air をインストール中..."
go install github.com/air-verse/air@v1.61.7

echo "✅ Go セットアップ完了: $(go version)"
