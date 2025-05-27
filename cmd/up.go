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

// upCmd команда наката миграции.
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all or <version> up migrations",
	Long: `Applies all migrations since the last applied migration.
You can specify which version to start applying migrations from (the version is a starting point and may not exist)
Command accepts all common flags. 
Depending on the format of the migrations, she can run the SQL file herself 
or build a program (golang) for executing and applying migrations

If parallel migration start is allowed in the settings, then parallel migrations are possible.
Attention, while the consistency of the database may suffer!
`,
	SilenceUsage: true,
	Example:      "migrator up <version> [flags] - where <version> is the version request",
	Run: func(_ *cobra.Command, args []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		runMigrate(ctx, cancelFunc, Up, args...)
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}

// Up - накатывает миграцию.
func Up(ctx context.Context, migrator migrate.Migrate, logger *zap.Logger, args ...string) error {
	var (
		requestToVersion uint64
		err              error
		count            int
	)
	if len(args) > 0 {
		requestToVersion, err = converter.VersionToUint(args[0])
		if err != nil {
			return domain.ErrMigrateVersionIncorrect
		}
	}
	count, err = migrator.Up(ctx, requestToVersion)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("total applied %d migrations", count))

	return nil
}
