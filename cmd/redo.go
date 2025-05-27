package cmd

import (
	"context"
	"fmt"

	"github.com/BashMS/SQL_migrator/pkg/migrate" //nolint:depguard
	"github.com/spf13/cobra"                     //nolint:depguard
	"go.uber.org/zap"                            //nolint:depguard
)

// redoCmd команда перенаката.
var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Roll back the most recently applied migration, then run it again",
	Long:  `The command rolls back the last applied migration and applies it again`,
	Run: func(_ *cobra.Command, _ []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		runMigrate(ctx, cancelFunc, Redo)
	},
}

func init() {
	rootCmd.AddCommand(redoCmd)
}

// Redo - откатывает и накатывает последнюю миграцию.
func Redo(ctx context.Context, migrator migrate.Migrate, logger *zap.Logger, _ ...string) error {
	migration, err := migrator.Redo(ctx)
	if err != nil {
		return err
	}
	if migration == nil {
		logger.Warn("not found in the database of applied migrations")
		return nil
	}

	logger.Info(fmt.Sprintf("version %d successfully rolled back and applied again", migration.Version))
	return nil
}
