package main

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// Up{{.Version}}{{.Name}} - apply migration.
func Up{{.Version}}{{.Name}}(ctx context.Context, tx pgx.Tx) error {
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

//Down{{.Version}}{{.Name}} - rollback migration.
func Down{{.Version}}{{.Name}}(ctx context.Context, tx pgx.Tx) error {
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
