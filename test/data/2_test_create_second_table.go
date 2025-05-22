//Package main - migration 2 named testCreateSecondTable.
package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func Up_2_testCreateSecondTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS "test_second_table"();`)
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("%w: %s", err, errRollback)
		}

		return err
	}

	return tx.Commit(ctx)
}

func Down_2_testCreateSecondTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `DROP TABLE IF EXISTS "test_second_table";`)
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("%w: %s", err, errRollback)
		}

		return err
	}

	return tx.Commit(ctx)
}
