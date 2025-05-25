package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4" //nolint:depguard
)

func Up3testCreateThirdTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS "test_third_table"();`)
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("%w: %s", err, errRollback.Error())
		}

		return err
	}

	return tx.Commit(ctx)
}

func Down3testCreateThirdTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `DROP TABLE IF EXISTS "test_third_table";`)
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("%w: %s", err, errRollback.Error())
		}

		return err
	}

	return tx.Commit(ctx)
}
