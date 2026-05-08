package migration

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
)

//go:embed sql/*.sql
var migrations embed.FS

func Up(ctx context.Context, db *sql.DB) error {
	entries, err := migrations.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		query, err := migrations.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := db.ExecContext(ctx, string(query)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}

	return nil
}
