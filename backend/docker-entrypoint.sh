#!/bin/sh
# コンテナ起動時に migration を適用してからアプリを起動する。
#
# スキーマ管理は Atlas の責務で、アプリ側は AutoMigrate を呼ばない（infrastructure/db.go）。
# ここで適用しないと、空のDBに対してサーバだけが立ち上がり
#   ERROR: relation "oauth_accounts" does not exist (SQLSTATE 42P01)
# のように最初のクエリで落ちる。
#
# --env gorm は使わない。あちらは差分生成用で scratch DB（atlas-dev）と
# `go run ./cmd/migrate` を要求するため、適用だけなら dir と url の指定で足りる。
set -e

: "${DATABASE_URL:?DATABASE_URL is required}"

# depends_on の healthcheck で待ってはいるが、接続が安定するまで数回リトライする。
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
