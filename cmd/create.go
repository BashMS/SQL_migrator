package cmd

import (
	"context"

	"github.com/BashMS/SQL_migrator/pkg/domain"  //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/migrate" //nolint:depguard
	"github.com/spf13/cobra"                     //nolint:depguard
	"go.uber.org/zap"                            //nolint:depguard
)

// createCmd команда создания.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a migration file",
	Long: `Creates migration files with the installed version (timestamped) and name in directory [--path/-p]
For the format [--format / -f] 'sql', two files with up/down postfixes are created, 
and for the 'go' format a go-file with 'Up*/Down*'' methods is generated`,
	Example: "migrator create <name> [flags]",
	Run: func(_ *cobra.Command, args []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		runMigrate(ctx, cancelFunc, Create, args...)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

// Create создает файл миграции.
func Create(_ context.Context, migrator migrate.Migrate, _ *zap.Logger, args ...string) error {
	if len(args) == 0 {
		return domain.ErrMigrationNameRequired
	}
	if err := migrator.Create(args[0]); err != nil {
		return err
	}

	return nil
}
