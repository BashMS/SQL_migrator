package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/BashMS/SQL_migrator/internal/report"
	"github.com/BashMS/SQL_migrator/pkg/migrate"
	"go.uber.org/zap"
)

//statusCmd команда статуса.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Displays the status of migrations in a table",
	Long: `Output status of all migrations. 
Data is taken from the migration table and contains the following fields:
Version - migration version (may contain only numbers)
Name - human-readable name of migration
Is applied? - migration status (applied or not applied)
Data update - Last update date at which any actions on migration were performed (for example, up, down, redo)
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		runMigrate(ctx, cancelFunc, Status)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

//Status - возвращает статус миграции.
func Status(ctx context.Context, migrator migrate.Migrate, logger *zap.Logger, args ...string) error {
	migrations, err := migrator.Status(ctx)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		logger.Warn("no migration found")
		return nil
	}

	report.PrintMigrations(migrations)
	return nil
}
