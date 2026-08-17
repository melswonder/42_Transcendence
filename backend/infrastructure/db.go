package infrastructure

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// maxOpenConns同時に開ける接続の上限
// maxIdleConns遊んでいる接続を何本まで維持するか
// connMaxLifetime接続の寿命
// 実行するときのキャパ的な
const (
	maxOpenConns    = 25
	maxIdleConns    = 25
	connMaxLifetime = 5 * time.Minute
)

// NewDB は DSN から接続を作り、疎通確認まで済ませて返す。
//
// postgres ドライバは gorm.Open の時点で実際に接続するため、
// DBが落ちていればここで失敗する。後段の Ping は、プール設定を
// 適用した状態でもう一度確かめるための保険。
//
// スキーマの管理は Atlas の責務なので AutoMigrate は呼ばない。
func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
