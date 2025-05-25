package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BashMS/SQL_migrator/pkg/config"  //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/logger"  //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/migrate" //nolint:depguard
	"github.com/spf13/cobra"                     //nolint:depguard
	"go.uber.org/zap"                            //nolint:depguard
)

const timeoutShutdown = 3 * time.Second

// AppVersion - версия.
const AppVersion = "0.0.1"

var (
	configFile string
	cfg        config.Config
)

// rootCmd базовая команда при вызове без каких-либо подкоманд.
var rootCmd = &cobra.Command{
	Use:   "migrator",
	Short: "Migration Tool",
	Long: `
Tool for working with migrations written in Go or represented as SQL files
Capabilities:
	* create - generate a migration template;
	* up - apply migration;
	* down - roll back migrations
	* redo - repetition of the last applied migration (down and up again)
	* status - displays the status of migrations in a table
	* version - output current version of migration
`,
	Version: AppVersion,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if configFile != "" {
			if err := cfg.ReadConfigFromFile(configFile); err != nil {
				return err
			}
		} else if cfg.ReadConfigFromDefaultPath() {
			fmt.Println("default configuration file loaded successfully")
		}
		cfg.Apply()
		return cfg.PathConversion()
	},
	Run: func(_ *cobra.Command, _ []string) {},
}

func init() {
	var err error
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "path to configuration file")
	rootCmd.PersistentFlags().StringVar(&cfg.DSN, "dsn", "", "database connection string (Data Source Name or DSN)")
	rootCmd.PersistentFlags().StringVarP(&cfg.Path, "path", "p", "", "absolute path to the migration folder")

	flagFormat := "format"
	rootCmd.PersistentFlags().StringVarP(
		&cfg.Format,
		flagFormat,
		"f",
		"",
		"format of migrations (\"sql\", \"golang\")")
	err = rootCmd.RegisterFlagCompletionFunc(
		flagFormat,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{config.FormatSQL, config.FormatGolang}, cobra.ShellCompDirectiveDefault
		})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVar(&cfg.LogPath, "log-path", "", "absolute path to the log")

	flagLogLevel := "log-level"
	rootCmd.PersistentFlags().StringVar(
		&cfg.LogLevel,
		flagLogLevel,
		"",
		"logging level (\"debug\", \"info\", \"warn\", \"error\" and \"fatal\")")
	err = rootCmd.RegisterFlagCompletionFunc(
		flagLogLevel,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{
				config.LogLevelDebug,
				config.LogLevelInfo,
				config.LogLevelWarn,
				config.LogLevelError,
				config.LogLevelFatal,
			}, cobra.ShellCompDirectiveDefault
		})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// Execute добавляет все дочерние команды к корневой команде и устанавливает соответствующие флаги.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// migrateFunc реализация функции работы с мигратором.
type migrateFunc func(ctx context.Context, migrator migrate.Migrate, logger *zap.Logger, args ...string) error

func runMigrate(ctx context.Context, cancelFunc context.CancelFunc, migrateFunc migrateFunc, args ...string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	zLogger, err := logger.New(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Flush(zLogger)

	consoleLogger := zLogger.Named(logger.ConsoleLogger)
	migrator := migrate.NewMigrate(zLogger, &cfg)
	chErr := make(chan error, 1)
	go func(consoleLogger *zap.Logger, chErr chan<- error) {
		if err := migrateFunc(ctx, migrator, consoleLogger, args...); err != nil {
			chErr <- err
		}

		cancelFunc()
	}(consoleLogger, chErr)

	select {
	case <-interrupt:
		cancelFunc()
		consoleLogger.Error("program was interrupted by the user")
		timer := time.NewTimer(timeoutShutdown)
		<-timer.C
		logger.Flush(zLogger)
		os.Exit(15) //nolint:gocritic
	case err := <-chErr:
		consoleLogger.Error(fmt.Sprintf("program terminated with an error: %s", err))
		logger.Flush(zLogger)
		os.Exit(1)
	case <-ctx.Done():
	}
}
