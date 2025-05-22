package template

var SampleGolangMigrationMethod = Sample{
	Name: "SampleGolangMigrationMethod",
	Text: `//Package main - migration {{.Version}} named {{.Name}}.
package main

import (
	"context"

	"github.com/jackc/pgx/v4"
)

//Up_{{.Version}}_{{.Name}} - apply migration.
func Up_{{.Version}}_{{.Name}}(ctx context.Context, tx pgx.Tx) error {
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

//Down_{{.Version}}_{{.Name}} - rollback migration.
func Down_{{.Version}}_{{.Name}}(ctx context.Context, tx pgx.Tx) error {
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
`,
}
