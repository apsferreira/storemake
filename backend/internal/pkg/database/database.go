package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect(dsn string) error {
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL não configurada")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("erro ao abrir conexão: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("erro ao conectar ao banco: %w", err)
	}

	DB = db
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
