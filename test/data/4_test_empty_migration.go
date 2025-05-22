//Package main - migration 4 named testEmptyMigration.
package main

import (
	"context"

	"github.com/jackc/pgx/v4"
)

//Up_4_testEmptyMigration - apply migration.
func Up_4_testEmptyMigration(ctx context.Context, tx pgx.Tx) error {
	//rows, err := tx.Exec(ctx, "-- SQL SCRIPT")
	//if err != nil {
	//	if errRollback := tx.Rollback(ctx); errRollback != nil {
	//		return fmt.Errorf("%w: %s", err, errRollback)
	//	}
	//
	//	return err
	//}
	//
	//fmt.Printf("Rows affected: %d\n", rows.RowsAffected())
	return tx.Commit(ctx)
}

//Down_4_testEmptyMigration - rollback migration.
func Down_4_testEmptyMigration(ctx context.Context, tx pgx.Tx) error {
	//rows, err := tx.Exec(ctx, "-- SQL SCRIPT")
	//if err != nil {
	//	if errRollback := tx.Rollback(ctx); errRollback != nil {
	//		return fmt.Errorf("%w: %s", err, errRollback)
	//	}
	//
	//	return err
	//}
	//
	//fmt.Printf("Rows affected: %d\n", rows.RowsAffected())
	return tx.Commit(ctx)
}
