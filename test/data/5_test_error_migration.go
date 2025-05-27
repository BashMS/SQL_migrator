// Package main - migration 5 named testErrorMigration.
package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4" //nolint:depguard
)

func Up5testErrorMigration(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, "SELECT Bad_Migratiom FORM MORF;")
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("%w: %s", err, errRollback.Error())
		}

		return err
	}

	return tx.Commit(ctx)
}

func Down5testErrorMigration(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, "SELECT Bad_Migratiom FORM MORF;")
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("%w: %s", err, errRollback.Error())
		}

		return err
	}

	return tx.Commit(ctx)
}
