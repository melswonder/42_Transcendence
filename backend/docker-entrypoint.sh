#!/bin/sh
# migration を適用してからアプリを起動する。適用しないと空のDBにサーバだけが立ち、
# relation "oauth_accounts" does not exist で落ちる。
#
# --env gorm は使わない。差分生成用で scratch DB（atlas-dev）を要求するため。
set -e

: "${DATABASE_URL:?DATABASE_URL is required}"

# healthcheck で待ってはいるが、接続が安定するまで数回リトライする。
i=1
until atlas migrate apply --dir "file://migrations" --url "$DATABASE_URL"; do
	if [ "$i" -ge 10 ]; then
		echo "migrate: failed after $i attempts" >&2
		exit 1
	fi
	echo "migrate: retrying ($i/10)..." >&2
	i=$((i + 1))
	sleep 2
done

exec "$@"
