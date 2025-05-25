package cmd

import (
	"context"
	"fmt"

	"github.com/BashMS/SQL_migrator/internal/converter" //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/domain"         //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/migrate"        //nolint:depguard
	"github.com/spf13/cobra"                            //nolint:depguard
	"go.uber.org/zap"                                   //nolint:depguard
)

const argDownAll = "all"

// downCmd команда отката.
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back of one or all or <version> down migrations",
	Long: `Roll back of one or all or <version> down migrations since the last applied migration.
You can specify which version to roll back migrations to (the version is a starting point and may not exist)
Command accepts all common flags. 
Depending on the format of the migrations, she can run the SQL file herself 
or build a program (golang) for executing and applying migrations

If parallel migration start is allowed in the settings, then parallel migrations are possible.
Attention, while the consistency of the database may suffer!`,
	SilenceUsage: true,
	Example:      "migrator down <version> [all] [flags] - where <version> is the version request",
	Run: func(_ *cobra.Command, args []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		runMigrate(ctx, cancelFunc, Down, args...)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

// Down - откатывает миграцию.
func Down(ctx context.Context, migrator migrate.Migrate, logger *zap.Logger, args ...string) error {
	var (
		requestToVersion uint64
		downAll          bool
		err              error
		count            int
	)
	argsCount := len(args)

	if argsCount > 0 && args[0] == argDownAll {
		downAll = true
	} else if argsCount > 0 && args[0] != argDownAll {
		requestToVersion, err = converter.VersionToUint(args[0])
		if err != nil {
			return domain.ErrMigrateVersionIncorrect
		}
	}
	if downAll {
		count, err = migrator.DownAll(ctx)
	} else {
		count, err = migrator.Down(ctx, requestToVersion)
	}
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("total %d migrations rolled back", count))

	return nil
}
