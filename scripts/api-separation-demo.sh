#!/bin/sh
# 通常 API と Public API の認証分離を実演するスクリプト。
#
# 証明すること:
#   1. 通常 API は認証なしでは叩けない (401)
#   2. 通常 API は API キー (Bearer) でも叩けない (401)  ← セッション Cookie 専用
#   3. Public API はセッション Cookie では叩けない (401)   ← API キー専用
#   4. Public API は API キーでなら叩ける (200)
#   5. read スコープのキーでは write 系を叩けない (403)
#
# 使い方: sh scripts/api-separation-demo.sh [BASE_URL]
#   既定は https://localhost:8443（ローカルの Caddy 経由）

BASE="${1:-https://localhost:8443}"
JAR="$(mktemp)"
SUFFIX="$(date +%s)"
PASS=0
FAIL=0

check() { # check <説明> <期待コード> <実際コード>
	if [ "$2" = "$3" ]; then
		printf 'PASS  %-52s -> %s\n' "$1" "$3"
		PASS=$((PASS + 1))
	else
		printf 'FAIL  %-52s -> %s (expected %s)\n' "$1" "$3" "$2"
		FAIL=$((FAIL + 1))
	fi
}

code() { curl -sk -o /dev/null -w '%{http_code}' "$@"; }

echo "== 対象: $BASE"
echo

echo "-- 1. 認証なしで通常 API"
check "GET /users/me (認証なし)" 401 "$(code "$BASE/users/me")"
check "GET /friends (認証なし)" 401 "$(code "$BASE/friends")"

echo
echo "-- 準備: デモユーザー登録 (セッション Cookie 取得)"
REG=$(code -c "$JAR" -X POST -H "Content-Type: application/json" \
	-d "{\"email\":\"apidemo$SUFFIX@example.com\",\"password\":\"demopass123\",\"display_name\":\"API Demo\",\"handle\":\"apidemo$SUFFIX\"}" \
	"$BASE/auth/register")
check "POST /auth/register" 201 "$REG"
check "GET /users/me (Cookie)" 200 "$(code -b "$JAR" "$BASE/users/me")"

echo
echo "-- 準備: API キーを発行 (write / read)"
KEY_W=$(curl -sk -b "$JAR" -X POST -H "Content-Type: application/json" \
	-d '{"name":"demo write key","scopes":["read","write"]}' "$BASE/apikeys" \
	| sed -n 's/.*"api_key":"\([^"]*\)".*/\1/p')
KEY_R=$(curl -sk -b "$JAR" -X POST -H "Content-Type: application/json" \
	-d '{"name":"demo read key","scopes":["read"]}' "$BASE/apikeys" \
	| sed -n 's/.*"api_key":"\([^"]*\)".*/\1/p')
[ -n "$KEY_W" ] && echo "PASS  API キー発行 (raw はこのレスポンスの一度きり)" || { echo "FAIL  API キー発行"; FAIL=$((FAIL + 1)); }

echo
echo "-- 2. API キーで通常 API は叩けない（セッション Cookie 専用）"
check "GET /users/me (Bearer キー)" 401 "$(code -H "Authorization: Bearer $KEY_W" "$BASE/users/me")"
check "GET /friends (Bearer キー)" 401 "$(code -H "Authorization: Bearer $KEY_W" "$BASE/friends")"

echo
echo "-- 3. セッション Cookie で Public API は叩けない（API キー専用）"
check "GET /v1/profile (Cookie)" 401 "$(code -b "$JAR" "$BASE/v1/profile")"
check "GET /v1/leaderboard (Cookie)" 401 "$(code -b "$JAR" "$BASE/v1/leaderboard")"

echo
echo "-- 4. API キーで Public API は叩ける"
check "GET /v1/profile (Bearer キー)" 200 "$(code -H "Authorization: Bearer $KEY_W" "$BASE/v1/profile")"
check "GET /v1/leaderboard (Bearer キー)" 200 "$(code -H "Authorization: Bearer $KEY_W" "$BASE/v1/leaderboard")"

echo
echo "-- 5. スコープ: read キーでは write 系は叩けない"
check "PUT /v1/profile (read キー)" 403 "$(code -X PUT -H "Authorization: Bearer $KEY_R" -H "Content-Type: application/json" -d '{"display_name":"x"}' "$BASE/v1/profile")"
check "PUT /v1/profile (write キー)" 200 "$(code -X PUT -H "Authorization: Bearer $KEY_W" -H "Content-Type: application/json" -d '{"display_name":"API Demo"}' "$BASE/v1/profile")"

echo
echo "== 結果: PASS=$PASS FAIL=$FAIL"
echo "   仕様書: $BASE/swagger/index.html"
rm -f "$JAR"
[ "$FAIL" = 0 ]
