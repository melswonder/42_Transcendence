// Command migrate は GORM の構造体定義から PostgreSQL の DDL を生成し、標準出力へ書き出す。
//
// 単体で実行して生成結果を確認できるほか、atlas.hcl の data "external_schema" から
// 呼び出されて migration の差分生成にも使われる。
//
//	go run ./cmd/migrate            # DDLを標準出力へ
//	atlas migrate diff --env gorm   # migrationファイルを生成
//
// partial unique index、CREATE EXTENSION citext、循環FK は GORM のタグでは
// 表現できないため、手書きSQL migration で補う。
package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"

	"transcendence-backend/infrastructure"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&infrastructure.User{},
		&infrastructure.MediaAsset{},
		&infrastructure.OAuthAccount{},
		&infrastructure.Session{},
		&infrastructure.Friendship{},
		&infrastructure.Block{},
		&infrastructure.Match{},
		&infrastructure.MatchParticipant{},
		&infrastructure.MatchAction{},
		&infrastructure.APIKey{},
		&infrastructure.UserAchievement{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}

	if _, err := io.WriteString(os.Stdout, stmts); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write schema: %v\n", err)
		os.Exit(1)
	}
}
