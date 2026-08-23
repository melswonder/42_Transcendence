#!/bin/sh
# 1 コンテナで backend + frontend + ルータ(Caddy) を起動する（無料ホスティング用）。
# 同一オリジンに同居させることで、Cookie 認証と WebSocket をそのまま使える。
set -e

: "${DATABASE_URL:?DATABASE_URL is required}"

i=1
# マネージド DB (Supabase 等) は auth などの独自スキーマを持つため、
# 管理対象を public に限定し、revision テーブルの置き場所を明示する。
# DATABASE_URL には &search_path=public を付けておくこと。
until atlas migrate apply --dir "file:///app/migrations" \
	--revisions-schema atlas_schema_revisions --url "$DATABASE_URL"; do
	if [ "$i" -ge 10 ]; then
		echo "migrate: failed after $i attempts" >&2
		exit 1
	fi
	echo "migrate: retrying ($i/10)..." >&2
	i=$((i + 1))
	sleep 2
done

mkdir -p "${MEDIA_DIR:-/app/uploads}"

# PORT はホスティングが公開ポートとして渡してくるので、内部の 2 つには別番号を明示する。
PORT=4000 /usr/local/bin/serv &
(cd /app/web && PORT=3000 HOSTNAME=127.0.0.1 node server.js) &

exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
