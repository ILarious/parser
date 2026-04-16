//go:build ignore

package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"todo_crud/pkg/postgres"
)

func main() {
	db, err := postgres.New(postgres.Config{
		Host:         "127.0.0.1",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		Database:     "todo",
		SSLMode:      "disable",
		MaxOpenConns: 20,
		MaxIdleConns: 20,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = postgres.Close(db) }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := postgres.WithTx(ctx, db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "select 1")
		return err
	}); err != nil {
		log.Fatal(err)
	}

	log.Printf("db healthy: %v", postgres.IsHealthy(db))
}
