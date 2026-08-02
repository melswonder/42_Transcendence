// Atlas 設定。cmd/migrate が出力するDDLを「あるべきスキーマ」として読み、
// 現在の migrations/ との差分から migration ファイルを生成する。
//
//   atlas migrate diff <name> --env gorm   # 差分から migration を生成
//   atlas migrate hash --env gorm          # atlas.sum を再計算
//   atlas migrate apply --env gorm --url "$DATABASE_URL"
//
// dev URL は Atlas が使い捨ての Postgres コンテナを起動して使う。差分計算のためだけの
// 空DBで、実行のたびに作られて捨てられる。Docker が動いていればよく、事前準備は不要。

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
  dev = "docker://postgres/16/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
