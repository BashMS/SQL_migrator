//+build integration

package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/BashMS/SQL_migrator/internal/command"
	"github.com/BashMS/SQL_migrator/internal/storage"
	"github.com/BashMS/SQL_migrator/pkg/config"
)

const binMigrator = "/src/.bin/migrator"
const timeout = 5 * time.Minute

func TestMigrator_FormatSQL(t *testing.T) {
	fmt.Println("Run test for SQL format ...")
	migratorTest(t, config.FormatSQL)
}

func TestMigrator_FormatGolang(t *testing.T) {
	fmt.Println("Run test for Golang format ...")
	migratorTest(t, config.FormatGolang)
}

func migratorTest(t *testing.T, format string) {
	conn, err := ConnectDB()
	if err != nil {
		t.Fatalf("failed to connect to database: %s", err)
	}
	defer CloseConnectDB(conn)
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	fmt.Println("** Clear database **")
	if err := clearDatabase(ctx, conn); err != nil {
		t.Fatalf("failed to clear database: %s", err)
	}

	env := os.Environ()
	cmd := command.NewCommand()
	fmt.Println("** Rolling migration with version 1 **")
	if err := cmd.Run(ctx, binMigrator, append(command.Args{"up", "1"}, getDefaultArgs(format)...), "/src", env); err != nil {
		t.Fatalf("failed to start command %s : %s", "migrator up 1", err)
	}
	assertTableExists(ctx, t, conn, "test_first_table")
	assertTableExists(ctx, t, conn, storage.MigrationsTable)

	fmt.Println("** Rolling migration with version 2 **")
	if err := cmd.Run(ctx, binMigrator, append(command.Args{"up", "2"}, getDefaultArgs(format)...), "/src", env); err != nil {
		t.Fatalf("failed to start command %s : %s", "migrator up 2", err)
	}
	assertTableExists(ctx, t, conn, "test_second_table")

	fmt.Println("** Running the redo command **")
	if err := cmd.Run(ctx, binMigrator, append(command.Args{"redo"}, getDefaultArgs(format)...), "/src", env); err != nil {
		t.Fatalf("failed to start command %s : %s", "migrator redo", err)
	}
	assertTableExists(ctx, t, conn, "test_second_table")

	fmt.Println("** Roll up to version 3 **")
	if err := cmd.Run(ctx, binMigrator, append(command.Args{"up", "3"}, getDefaultArgs(format)...), "/src", env); err != nil {
		t.Fatalf("failed to start command %s : %s", "migrator up 3", err)
	}
	assertTableExists(ctx, t, conn, "test_third_table")

	fmt.Println("** Rollback to one version (up to 2) **")
	if err := cmd.Run(ctx, binMigrator, append(command.Args{"down"}, getDefaultArgs(format)...), "/src", env); err != nil {
		t.Fatalf("failed to start command %s : %s", "migrator down", err)
	}
	assertTableExists(ctx, t, conn, "test_second_table")
	assertTableNoExists(ctx, t, conn, "test_third_table")

	fmt.Println("** Rollback to the second inclusive version (up to 1) **")
	if err := cmd.Run(ctx, binMigrator, append(command.Args{"down", "2"}, getDefaultArgs(format)...), "/src", env); err != nil {
		t.Fatalf("failed to start command %s : %s", "migrator down 2", err)
	}
	assertTableExists(ctx, t, conn, "test_first_table")
	assertTableNoExists(ctx, t, conn, "test_second_table")
	assertTableNoExists(ctx, t, conn, "test_third_table")

	fmt.Println("** Roll up to version 5 and wait for an error **")
	err = cmd.Run(ctx, binMigrator, append(command.Args{"up", "5"}, getDefaultArgs(format)...), "/src", env)
	if err == nil {
		t.Fatalf("there is no error in case of bad migration : %s", "migrator up 5")
	}

	fmt.Println("** Rollback all versions **")
	if err := cmd.Run(ctx, binMigrator, append(command.Args{"down", "all"}, getDefaultArgs(format)...), "/src", env); err != nil {
		t.Fatalf("failed to start command %s : %s", "down all", err)
	}
	assertTableNoExists(ctx, t, conn, "test_first_table")
	assertTableNoExists(ctx, t, conn, "test_second_table")
	assertTableNoExists(ctx, t, conn, "test_third_table")
}

func getDefaultArgs(format string) command.Args {
	return command.Args{"-c", "/src/.bin/config.yml", "-f", format, "-p", "/src/test/data"}
}

func assertTableNoExists(ctx context.Context, t *testing.T, conn *pgx.Conn, tableCheck string) {
	var (
		ok  bool
		err error
	)
	ok, err = checkTable(ctx, conn, storage.MigrationsScheme, tableCheck)
	if err != nil {
		t.Fatalf("error occurred while checking the %s table: %s", tableCheck, err)
	}
	if ok {
		t.Fatalf("%s table still exists", tableCheck)
	}
}

func assertTableExists(ctx context.Context, t *testing.T, conn *pgx.Conn, tableCheck string) {
	var (
		ok  bool
		err error
	)
	ok, err = checkTable(ctx, conn, storage.MigrationsScheme, tableCheck)
	if err != nil {
		t.Fatalf("error occurred while checking the %s table: %s", tableCheck, err)
	}
	if !ok {
		t.Fatalf("%s not found", tableCheck)
	}
}

func checkTable(ctx context.Context, conn *pgx.Conn, scheme, table string) (bool, error) {
	query := `
	SELECT EXISTS (
   		SELECT FROM information_schema.tables 
   		WHERE table_schema = $1 AND table_name = $2
   );
`
	var ok bool
	if err := conn.QueryRow(ctx, query, scheme, table).Scan(&ok); err != nil {
		return false, err
	}

	return ok, nil
}

func clearDatabase(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
DROP TABLE IF EXISTS "public"."tmigration"; 
DROP TABLE IF EXISTS "public"."test_first_table";
DROP TABLE IF EXISTS "public"."test_second_table";
DROP TABLE IF EXISTS "public"."test_third_table";
`)
	if err != nil {
		return err
	}

	return nil
}
