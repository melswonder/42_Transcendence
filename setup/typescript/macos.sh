#!/usr/bin/env bash
# フロントエンドのセットアップは macOS / Linux 共通なので linux.sh に委譲する
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec bash "$SCRIPT_DIR/linux.sh"
