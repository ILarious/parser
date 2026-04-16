package postgres

import (
	"context"
	"database/sql"
	"time"
)

func Ping(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrNilDB
	}
	return db.PingContext(ctx)
}

func IsHealthy(db *sql.DB) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return Ping(ctx, db) == nil
}
