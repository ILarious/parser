package postgres

import "errors"

var (
	ErrNilDB   = errors.New("postgres: nil db")
	ErrNilTxFn = errors.New("postgres: nil tx callback")
)
