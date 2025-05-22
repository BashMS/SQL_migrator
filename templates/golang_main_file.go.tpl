package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BashMS/SQL_migrator/pkg/config"
	"github.com/BashMS/SQL_migrator/pkg/logger"
	"github.com/BashMS/SQL_migrator/pkg/migrate"
	"go.uber.org/zap"
)

type customFunc struct {
	migrateFunc migrate.CustomMigrateFunc
	name        string
	version     uint64
	direction   bool
}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	config := config.Config{
		DSN:              "{{.Config.DSN}}",
		LogPath:          "{{.Config.LogPath}}",
		LogLevel:         "{{.Config.LogLevel}}",
	}

	zLogger, err := logger.New(&config)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Flush(zLogger)

	migrator := migrate.NewMigrate(zLogger, &config)
	var migrationFuncs []customFunc
{{$direction := .Direction}}
{{$prefix := "Down"}}
{{if .Direction}} {{$prefix = "Up"}} {{end}}
{{range $k, $migration := .Migrations}}
{{$FN := printf "%s_%d_%s" $prefix $migration.Version $migration.Name}}
    migrationFuncs = append(migrationFuncs, customFunc{
        migrateFunc: {{$FN}},
        name:        "{{$migration.Name}}",
        version:     {{$migration.Version}},
        direction:   {{$direction}},
    })
{{end}}
	go func() {
    		for _, f := range migrationFuncs {
    			if err := migrator.RunMigrationWithCustomFunc(ctx, f.migrateFunc, f.name, f.version, f.direction);
    				err != nil {
    				zLogger.Error("failed migration {{$prefix}}", zap.Error(err))
    				os.Exit(3)
    			} else {
    				zLogger.Info("migration completed successfully",
    					zap.Uint64("version", f.version),
    					zap.String("name", f.name),
    				)
    			}
    		}
    		cancelFunc()
    }()

    select {
    case <-interrupt:
        cancelFunc()
        timer := time.NewTimer(3 * time.Second)
        zLogger.Error("subprogram was interrupted by the user")
        select {
        case <-timer.C:
        }
        os.Exit(1)
    case <-ctx.Done():
    }
}
