package db

import (
	"database/sql"
	"fmt"
)

func Tx(db *sql.DB, cb func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := cb(tx); err != nil {
		if txe := tx.Rollback(); txe != nil {
			err = fmt.Errorf(
				"in addition to %w: the rollback failed with: %s",
				err,
				txe,
			)
		}

		return err
	}

	return tx.Commit()
}

type chain struct {
	tx  *sql.Tx
	err error
	ids []int
}

func (c *chain) Err() error { return c.err }
func (c *chain) Exec(query string, args ...any) *chain {
	if c.err != nil {
		return c
	}

	_, c.err = c.tx.Exec(query, args...)

	return c
}

func Chain(tx *sql.Tx) *chain { return &chain{tx: tx, ids: make([]int, 0)} }
