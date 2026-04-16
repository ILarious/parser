# pkg/postgres

Simple PostgreSQL helper package for this project.

## API

- `postgres.New(cfg)` opens DB and pings it
- `postgres.Close(db)` closes DB
- `postgres.Ping(ctx, db)` checks connection
- `postgres.IsHealthy(db)` returns bool health status
- `postgres.WithTx(ctx, db, fn)` runs callback in transaction

## Driver

Import postgres driver in your app binary:

```go
import _ "github.com/jackc/pgx/v5/stdlib"
```
