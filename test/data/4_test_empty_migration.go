// Package main - migration 4 named testEmptyMigration.
package main

import (
	"context"

	"github.com/jackc/pgx/v4" //nolint:depguard
)

// Up4testEmptyMigration - apply migration.
func Up4testEmptyMigration(ctx context.Context, tx pgx.Tx) error {
	// rows, err := tx.Exec(ctx, "-- SQL SCRIPT")
	// if err != nil {
	//	if errRollback := tx.Rollback(ctx); errRollback != nil {
	//		return fmt.Errorf("%w: %s", err, errRollback.Error())
	//	}
	//
	//	return err
	// }
	//
	// fmt.Printf("Rows affected: %d\n", rows.RowsAffected())
	return tx.Commit(ctx)
}

// Down4testEmptyMigration - rollback migration.
func Down4testEmptyMigration(ctx context.Context, tx pgx.Tx) error {
	// rows, err := tx.Exec(ctx, "-- SQL SCRIPT")
	// if err != nil {
	//	if errRollback := tx.Rollback(ctx); errRollback != nil {
	//		return fmt.Errorf("%w: %s", err, errRollback.Error())
	//	}
	//
	//	return err
	// }
	//
	// fmt.Printf("Rows affected: %d\n", rows.RowsAffected())
	return tx.Commit(ctx)
}
