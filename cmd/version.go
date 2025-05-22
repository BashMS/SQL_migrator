package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/BashMS/SQL_migrator/internal/report"
	"github.com/BashMS/SQL_migrator/pkg/migrate"
	"go.uber.org/zap"
)

// versionCmd команда версии миграции.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print current migration version",
	Long:  `The command displays information about the latest version applied.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		runMigrate(ctx, cancelFunc, Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

//Version - возвращает текущую миграцию.
func Version(ctx context.Context, migrator migrate.Migrate, logger *zap.Logger, args ...string) error {
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
