#!/usr/bin/env bash
# Go のインストール（公式 tarball） + air（ホットリロード）— Linux 用
set -euo pipefail

GO_VERSION=1.23.4
GO_REQUIRED_MINOR=23   # backend/go.mod の go ディレクティブに合わせる（1.x の x）

# CPU アーキテクチャを判定
case "$(uname -m)" in
  x86_64)        GOARCH=amd64 ;;
  aarch64|arm64) GOARCH=arm64 ;;
  *) echo "❌ 未対応のアーキテクチャ: $(uname -m)"; exit 1 ;;
esac

# 既存の go が go.mod の要求を満たすか判定する
# （"go があるか" だけを見ると、古い go が居座っていても素通りしてしまう）
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
  GO_BIN_DIR="$(cd "$(dirname "$(command -v go)")" && pwd)"
else
  if command -v go >/dev/null 2>&1; then
    echo "⚠️  既存の go が古いため入れ直します: $(go version) → go${GO_VERSION}"
  fi

  # 書き込み先を決める。sudo が使えない環境（42 のマシン等）では $HOME に入れる
  if sudo -n true 2>/dev/null; then
    PREFIX=/usr/local
    SUDO=sudo
  else
    PREFIX="$HOME/.local"
    SUDO=""
    mkdir -p "$PREFIX"
    echo "ℹ️  sudo が使えないため $PREFIX/go にインストールします"
  fi

  TARBALL="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "$TMP_DIR"' EXIT
  echo "⬇️  ${TARBALL} をダウンロード中..."
  curl -fL -o "$TMP_DIR/$TARBALL" "https://go.dev/dl/${TARBALL}"
  $SUDO rm -rf "$PREFIX/go"
  $SUDO tar -C "$PREFIX" -xzf "$TMP_DIR/$TARBALL"
  GO_BIN_DIR="$PREFIX/go/bin"
fi

# 現在のシェルで PATH を通す
# 新しい go を「先頭」に置くのが重要（/usr/bin/go など古いものに負けないように）
export PATH="$GO_BIN_DIR:$PATH:$("$GO_BIN_DIR/go" env GOPATH)/bin"

# シェル設定ファイルに永続化（未追記なら）
# ログインシェルが zsh のこともあるので $SHELL を見て振り分ける
case "$(basename "${SHELL:-bash}")" in
  zsh) RC_FILE="$HOME/.zshrc" ;;
  *)   RC_FILE="$HOME/.bashrc" ;;
esac
if ! grep -qsF "$GO_BIN_DIR" "$RC_FILE" 2>/dev/null; then
  echo "export PATH=\"$GO_BIN_DIR:\$PATH:\$(go env GOPATH)/bin\"" >> "$RC_FILE"
  echo "📝 $RC_FILE に PATH を追記しました"
fi

# air（ホットリロード）
# @latest は Go 1.26 以上を要求するため、Go 1.23 対応の最終版に固定（backend/Dockerfile と揃える）
echo "⬇️  air をインストール中..."
go install github.com/air-verse/air@v1.61.7

echo "✅ Go セットアップ完了: $(go version)"
