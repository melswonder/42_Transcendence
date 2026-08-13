// Atlas 設定。cmd/migrate が出力するDDLを「あるべきスキーマ」として読み、
// 現在の migrations/ との差分から migration ファイルを生成する。
//
//   atlas migrate diff <name> --env gorm   # 差分から migration を生成
//   atlas migrate hash --env gorm          # atlas.sum を再計算
//   atlas migrate apply --env gorm --url "$DATABASE_URL"
//
// dev は Atlas が差分計算に使うスクラッチDB。docker-compose の atlas-dev サービスで、
// 事前に `docker compose up -d atlas-dev` が必要。
//
// Atlas 組み込みの docker://postgres/16/dev は使わない。あちらは citext 拡張を
// 入れる手段が atlas login（有料）でしか無く、拡張の無いDBでは
//   pq: type "citext" does not exist
// で差分生成が落ちるため。atlas-dev は migrations の citext 用ファイルを
// 初期化スクリプトとしてマウントしてある（docker-compose.yml 参照）。

data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "./cmd/migrate",
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  // バージョンは db/postgres/Dockerfile と揃える
  dev = "postgres://postgres:postgres@localhost:5433/atlas_dev?search_path=public"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
