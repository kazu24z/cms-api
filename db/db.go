package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/sqlite3/*.sql
var sqliteMigrationsFS embed.FS

// TODO: PostgreSQL 対応時に追加
// //go:embed migrations/postgres/*.sql
// var postgresMigrationsFS embed.FS

var DB *sql.DB

// DBDriver は現在使用中のドライバ名
var DBDriver string

func Init() error {
	DBDriver = os.Getenv("DB_DRIVER")
	if DBDriver == "" {
		DBDriver = "sqlite3" // デフォルト
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		if DBDriver == "sqlite3" {
			dbURL = "cms.db"
		} else {
			return fmt.Errorf("DATABASE_URL is required for %s", DBDriver)
		}
	}

	var err error
	DB, err = sql.Open(DBDriver, dbURL)
	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {
		return err
	}

	log.Printf("Database connected: %s (%s)", DBDriver, dbURL)
	return nil
}

func Migrate() error {
	var m *migrate.Migrate

	switch DBDriver {
	case "sqlite3":
		sourceDriver, err := iofs.New(sqliteMigrationsFS, "migrations/sqlite3")
		if err != nil {
			return err
		}
		dbDriver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
		if err != nil {
			return err
		}
		m, err = migrate.NewWithInstance("iofs", sourceDriver, DBDriver, dbDriver)
		if err != nil {
			return err
		}

	case "postgres":
		// TODO: PostgreSQL 対応時に実装
		// 1. migrations/postgres/ にマイグレーションファイルを作成
		// 2. //go:embed migrations/postgres/*.sql を有効化
		// 3. ここのコメントを解除
		return fmt.Errorf("postgres driver is not implemented yet. See README.md for setup instructions")

	default:
		return fmt.Errorf("unsupported database driver: %s", DBDriver)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrated")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
