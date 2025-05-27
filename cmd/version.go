package cmd

import (
	"context"
	"fmt"

	"github.com/BashMS/SQL_migrator/internal/report" //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/migrate"     //nolint:depguard
	"github.com/spf13/cobra"                         //nolint:depguard
	"go.uber.org/zap"                                //nolint:depguard
)

// versionCmd команда версии миграции.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print current migration version",
	Long:  `The command displays information about the latest version applied.`,
	Run: func(_ *cobra.Command, _ []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		runMigrate(ctx, cancelFunc, Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Version - возвращает текущую миграцию.
func Version(ctx context.Context, migrator migrate.Migrate, logger *zap.Logger, _ ...string) error {
	migration, err := migrator.MigrateVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest migration version: %w", err)
	}

	if migration == nil {
		logger.Warn("no migration applied")
		return nil
	}
	report.PrintMigration(*migration)
	return nil
}
