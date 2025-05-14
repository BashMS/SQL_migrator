package core

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	//"strings"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/BashMS/SQL_migrator/internal/command"
	"github.com/BashMS/SQL_migrator/internal/converter"
	"github.com/BashMS/SQL_migrator/internal/loader"
	"github.com/BashMS/SQL_migrator/internal/storage"
	"github.com/BashMS/SQL_migrator/internal/template"
	"github.com/BashMS/SQL_migrator/internal/util"
	"github.com/BashMS/SQL_migrator/pkg/config"
	"github.com/BashMS/SQL_migrator/pkg/domain"
	"github.com/BashMS/SQL_migrator/pkg/logger"
)

const uidName = "migrator"

type (
	DeferFunc func()

	// MigrateCore.
	MigrateCore struct {
		storage storage.MigrateStorage
		command command.Command
		logger  *zap.Logger
		config  *config.Config
		loader  loader.Loader
	}
)

// NewMigrateCore  конструктор.
func NewMigrateCore(
	storage storage.MigrateStorage,
	cmd command.Command,
	zLogger *zap.Logger,
	config *config.Config,
) *MigrateCore {
	consoleLogger := zLogger.Named(logger.ConsoleLogger)
	return &MigrateCore{
		storage: storage,
		command: cmd,
		logger:  consoleLogger,
		config:  config,
		loader:  loader.NewLoader(consoleLogger),
	}
}

// ConnectDB - соединение с БД.
func (mc *MigrateCore) ConnectDB(ctx context.Context) (DeferFunc, error) {
	if err := mc.storage.Connect(ctx); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrConnection, err)
	}

	return mc.storage.Close, nil
}

// LoadMigrations - загружает все файлы миграции.
func (mc *MigrateCore) LoadMigrations(
	ctx context.Context,
	requestToVersion uint64,
	direction bool,
) ([]loader.RawMigration, error) {
	if err := mc.validateFormat(mc.config.Format); err != nil {
		return nil, err
	}
	mc.loader.SetFormat(mc.config.Format)
	excludeMigrations, err := mc.storage.GetMigrationsByDirection(ctx, direction)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrLoadMigrations, err)
	}

	filter := loader.Filter{Exclude: excludeMigrations}
	if requestToVersion != 0 {
		filter.RequestToVersion = requestToVersion
	} else {
		filter.Recent, err = mc.storage.RecentMigration(ctx)
		if err != nil && err != pgx.ErrNoRows {
			return nil, fmt.Errorf("%w: %s", domain.ErrGetRecentMigration, err)
		}
		// Если миграций нет и текущая — MigrationDown,
        // то считаем, что все миграции откатываются.
		if filter.Recent.Version == 0 && !direction {
			return nil, nil
		}
	}
	neededMigrations, err := mc.loader.LoadMigrations(ctx, filter, mc.config.Path, direction)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrLoadMigrations, err)
	}

	return neededMigrations, nil
}

// StartMigrate - запускает процесc миграции.
func (mc *MigrateCore) StartMigrate(
	ctx context.Context,
	neededMigrations []loader.RawMigration,
	direction bool,
) (int, error) {
	if err := mc.validateFormat(mc.config.Format); err != nil {
		return 0, err
	}
	switch mc.config.Format {
	case config.FormatSQL:
		return mc.runSQLMigration(ctx, neededMigrations, direction)
	case config.FormatGolang:
		return mc.runGoMigration(ctx, neededMigrations, direction)
	}

	return 0, nil
}

// GetRecentMigration - возвращает последнюю примененную миграцию.
func (mc *MigrateCore) GetRecentMigration(ctx context.Context) (*domain.Migration, error) {
	migration, err := mc.storage.RecentMigration(ctx)
	if err != nil && err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &migration, nil
}

// GetMigrations - возвращает все миграции из БД.
func (mc *MigrateCore) GetMigrations(ctx context.Context) ([]domain.Migration, error) {
	return mc.storage.Stats(ctx)
}

// CreateTransactionalMigration - создает транзакционную миграцию в направлении вверх или вниз.
func (mc *MigrateCore) CreateTransactionalMigration(
	ctx context.Context,
	migration domain.Migration,
	direction bool,
) (pgx.Tx, error) {
	tx, err := mc.storage.BeginTxMigration(ctx, migration, direction)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// CreateMigrationFile - создать файл миграции в зависимости от формата.
func (mc *MigrateCore) CreateMigrationFile(name string, version uint64) error {
	if version == 0 {
		return domain.ErrMigrateVersionIncorrect
	}
	name = converter.SanitizeMigrationName(name)
	paths, err := mc.getFilePath(name, version)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		return domain.ErrCreateMigrationFile
	}

	switch mc.config.Format {
	case config.FormatGolang:
		migration := domain.Migration{
			Version: version,
			Name:    strcase.ToLowerCamel(name),
		}
		if fileutil.Exist(paths[0]) {
			return fmt.Errorf("%w: %s", domain.ErrMigrationFileExists, paths[0])
		}
		if err := template.CreateGolangMigrationMethod(paths[0], migration); err != nil {
			return fmt.Errorf("%w: %s", domain.ErrCreateMigrationFile, paths[0])
		}
		mc.logger.Info(fmt.Sprintf("%s created successfully", paths[0]))
	case config.FormatSQL:
		for _, filePath := range paths {
			if fileutil.Exist(filePath) {
				return fmt.Errorf("%w: %s", domain.ErrMigrationFileExists, filePath)
			}
			if err := util.CreateFile(filePath); err != nil {
				return fmt.Errorf("%w: %s", domain.ErrCreateMigrationFile, filePath)
			}
			mc.logger.Info(fmt.Sprintf("%s created successfully", filePath))
		}
	}

	return nil
}

func (mc *MigrateCore) getFilePath(name string, version uint64) ([]string, error) {
	if name == "" {
		return nil, domain.ErrMigrationNameRequired
	}
	if err := mc.validateFormat(mc.config.Format); err != nil {
		return nil, err
	}
	name = converter.SanitizeMigrationName(name)
	fileName := fmt.Sprintf("%d_%s", version, strcase.ToDelimited(name, config.Separator))
	var filePaths []string
	switch mc.config.Format {
	case config.FormatGolang:
		fullName := fmt.Sprintf("%s%s", fileName, config.ExtGolang)
		filePaths = append(filePaths, filepath.Join(mc.config.Path, fullName))
	case config.FormatSQL:
		for _, postfix := range []string{config.PostfixUp, config.PostfixDown} {
			fullName := fmt.Sprintf("%s%s%s", fileName, postfix, config.ExtSQL)
			filePaths = append(filePaths, filepath.Join(mc.config.Path, fullName))
		}
	}

	return filePaths, nil
}

func (mc *MigrateCore) runSQLMigration(
	ctx context.Context,
	rawMigrations []loader.RawMigration,
	direction bool,
) (int, error) {
	var (
		count int
		err   error
	)
	for _, rawMigration := range rawMigrations {
		query := rawMigration.GetQuery(direction)

		//skip empty up-migration
		if query == "" && direction {
			mc.logger.Warn(fmt.Sprintf("%s empty migration file detected, it will be skipped",
				rawMigration.GetPath(direction)))
			continue
		}

		var tx pgx.Tx
		tx, err = mc.CreateTransactionalMigration(ctx, domain.Migration{
			Version: rawMigration.Version,
			Name:    rawMigration.Name,
		}, direction)
		if err != nil {
			if errors.Is(err, storage.ErrQueryNoAffectRows) {
				continue
			}
			return count, err
		}

		sDirection := "Down"
		if direction {
			sDirection = "Up"
		}

		mc.logger.Info(fmt.Sprintf("running %s migration with version %d (%s)...",
			rawMigration.Name, rawMigration.Version, sDirection))

		var rowAffected int64
		rowAffected, err = mc.exec(ctx, tx, query)

		if err != nil {
			return count, err
		}
		mc.logger.Debug(fmt.Sprintf("%d row affected", rowAffected))
		count++
	}

	return count, nil
}

func (mc *MigrateCore) runGoMigration(
	ctx context.Context,
	rawMigrations []loader.RawMigration,
	direction bool,
) (int, error) {
	mc.logger.Info("build a program for migrations...")
	tmpPath, err := ioutil.TempDir(os.TempDir(), "migrator_*")
	if err != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrBuildProgramForMigrations, err)
	}
	defer os.RemoveAll(tmpPath)

	for _, rawMigration := range rawMigrations {
		base := path.Base(rawMigration.GetPath(direction))
		if err := util.CopyFile(filepath.Join(tmpPath, base), rawMigration.GetPath(direction)); err != nil {
			return 0, fmt.Errorf("%w: %s", domain.ErrBuildProgramForMigrations, err)
		}
	}

	if err := template.CreateMainSample(filepath.Join(tmpPath, "main.go"), mc.config, rawMigrations, direction); err != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrBuildProgramForMigrations, err)
	}

	var env command.Env
	if err := mc.command.Run(ctx, "go", command.Args{"mod", "init", "go/migration"}, tmpPath, env); err != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrBuildProgramForMigrations, err)
	}

	if err := mc.command.Run(ctx, "go", command.Args{"mod", "tidy"}, tmpPath, env); err != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrBuildProgramForMigrations, err)
	}
	env = append(env, os.Environ()...)
	env = append(env, "GO111MODULE=on")
	mc.logger.Info("starting a program for migrations...")
	if err := mc.command.RunWithGracefulShutdown(ctx, "go", command.Args{"run", "./..."}, tmpPath, env); err != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrStartingProgramForMigrations, err)
	}

	return len(rawMigrations), nil
}

func (mc *MigrateCore) exec(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) (int64, error) {
	var (
		err    error
		result pgconn.CommandTag
	)

	result, err = tx.Exec(ctx, query, args...)

	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return 0, fmt.Errorf("%w: %s: %s", domain.ErrTransactionCancel, domain.ErrApplyingMigration, errRollback)
		}

		return 0, fmt.Errorf("%w: %s", domain.ErrApplyingMigration, err)
	}

	if errCommit := tx.Commit(ctx); errCommit != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrApplyingMigration, errCommit)
	}

	return result.RowsAffected(), nil
}

func (mc *MigrateCore) execWithoutTransaction(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var (
		err    error
		result pgconn.CommandTag
	)

	var conn *pgx.Conn
	mc.storage.Close()
	if conn, err = mc.storage.GetConnection(ctx); err != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrApplyingMigration, err)
	}

	if result, err = conn.Exec(ctx, query, args...); err != nil {
		return 0, fmt.Errorf("%w: %s", domain.ErrApplyingMigration, err)
	}

	return result.RowsAffected(), nil
}

func (mc *MigrateCore) validateFormat(format string) error {
	if format == config.FormatSQL || format == config.FormatGolang {
		return nil
	}

	return fmt.Errorf("%w (allow %s or %s)",
		domain.ErrInvalidFormat, config.FormatSQL, config.FormatGolang)
}
