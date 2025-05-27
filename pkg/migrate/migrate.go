package migrate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4" //nolint:depguard

	"github.com/BashMS/SQL_migrator/internal/command" //nolint:depguard
	"github.com/BashMS/SQL_migrator/internal/core"    //nolint:depguard
	"github.com/BashMS/SQL_migrator/internal/storage" //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/config"       //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/domain"       //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/logger"       //nolint:depguard
	"go.uber.org/zap"                                 //nolint:depguard
)

const (
	// MigrationUp - накат миграции.
	MigrationUp = true
	// MigrationDown откат миграции.
	MigrationDown = false
)

// CustomMigrateFunc - пользовательская функция для миграций.
type CustomMigrateFunc func(ctx context.Context, tx pgx.Tx) error

// Migrate.
type Migrate interface {
	Create(name string) error
	Status(ctx context.Context) ([]domain.Migration, error)
	Up(ctx context.Context, requestToVersion uint64) (int, error)
	DownAll(ctx context.Context) (int, error)
	Down(ctx context.Context, requestToVersion uint64) (int, error)
	Redo(ctx context.Context) (*domain.Migration, error)
	RunMigrationWithCustomFunc(ctx context.Context,
		migrateFunc CustomMigrateFunc, name string, version uint64, direction bool) error
	MigrateVersion(ctx context.Context) (*domain.Migration, error)
}

type migrate struct {
	migrateCore *core.MigrateCore
	logger      *zap.Logger
	config      *config.Config
}

// NewMigrate конструктор.
func NewMigrate(zLogger *zap.Logger, config *config.Config) Migrate {
	migrateStorage := storage.NewStorage(zLogger, config)
	return &migrate{
		migrateCore: core.NewMigrateCore(migrateStorage, command.NewCommand(), zLogger, config),
		logger:      zLogger.Named(logger.ConsoleLogger),
		config:      config,
	}
}

// Create создать файл миграции.
// Создает файлы миграции с установленной версией (с меткой времени) и именем в каталоге.
func (m *migrate) Create(name string) error {
	version := uint64(time.Now().Unix()) //nolint:gosec
	return m.migrateCore.CreateMigrationFile(name, version)
}

// Up - применить все или N миграций вверх.
// Применяет все миграции с момента последней примененной миграции.
func (m *migrate) Up(ctx context.Context, requestToVersion uint64) (int, error) {
	closeFunc, err := m.migrateCore.ConnectDB(ctx)
	if err != nil {
		return 0, err
	}
	defer closeFunc()

	neededMigrations, err := m.migrateCore.LoadMigrations(ctx, requestToVersion, MigrationUp)
	if err != nil {
		return 0, err
	}

	if len(neededMigrations) == 0 {
		return 0, nil
	}

	return m.migrateCore.StartMigrate(ctx, neededMigrations, MigrationUp)
}

// Down - откатить все миграции.
// Откатить все миграции с момента последней примененной миграции.
func (m *migrate) DownAll(ctx context.Context) (int, error) {
	closeFunc, err := m.migrateCore.ConnectDB(ctx)
	if err != nil {
		return 0, err
	}
	defer closeFunc()

	neededMigrations, err := m.migrateCore.LoadMigrations(ctx, 0, MigrationDown)
	if err != nil {
		return 0, err
	}

	if len(neededMigrations) == 0 {
		return 0, nil
	}

	return m.migrateCore.StartMigrate(ctx, neededMigrations, MigrationDown)
}

// Down - откат одной или N миграций вниз.
// Откат одной миграции с момента последней примененной миграции.
func (m *migrate) Down(ctx context.Context, requestToVersion uint64) (int, error) {
	closeFunc, err := m.migrateCore.ConnectDB(ctx)
	if err != nil {
		return 0, err
	}
	defer closeFunc()

	if requestToVersion == 0 {
		migration, err := m.migrateCore.GetRecentMigration(ctx)
		if err != nil || migration == nil {
			return 0, err
		}
		requestToVersion = migration.Version
	}

	neededMigrations, err := m.migrateCore.LoadMigrations(ctx, requestToVersion, MigrationDown)
	if err != nil {
		return 0, err
	}

	if len(neededMigrations) == 0 {
		return 0, nil
	}

	return m.migrateCore.StartMigrate(ctx, neededMigrations, MigrationDown)
}

// Redo - откатывает последнюю примененную миграцию и накатывает ее снова.
func (m *migrate) Redo(ctx context.Context) (*domain.Migration, error) {
	closeFunc, err := m.migrateCore.ConnectDB(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	migration, err := m.migrateCore.GetRecentMigration(ctx)
	if err != nil || migration == nil {
		return nil, err
	}

	neededMigrations, err := m.migrateCore.LoadMigrations(ctx, migration.Version, MigrationDown)
	if err != nil {
		return migration, err
	}
	if len(neededMigrations) == 0 {
		return migration, nil
	}

	recentMigrations := neededMigrations[len(neededMigrations)-1:]

	affectedMigration, err := m.migrateCore.StartMigrate(ctx, recentMigrations, MigrationDown)
	if err != nil {
		return migration, err
	}
	if affectedMigration == 0 {
		return migration, fmt.Errorf("failed to roll back migration with version %d", migration.Version)
	}

	affectedMigration, err = m.migrateCore.StartMigrate(ctx, recentMigrations, MigrationUp)
	if err != nil {
		return migration, err
	}
	if affectedMigration == 0 {
		return migration, fmt.Errorf("failed to up migration with version %d", migration.Version)
	}

	return migration, nil
}

// MigrateVersion возвращает информацию о последней выведенной версии.
func (m *migrate) MigrateVersion(ctx context.Context) (*domain.Migration, error) {
	closeFunc, err := m.migrateCore.ConnectDB(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	migration, err := m.migrateCore.GetRecentMigration(ctx)
	if err != nil || migration == nil {
		return nil, err
	}

	return migration, err
}

// Status - возвращаемый статус всех миграций.
// Данные берутся из таблицы миграций.
func (m *migrate) Status(ctx context.Context) ([]domain.Migration, error) {
	closeFunc, err := m.migrateCore.ConnectDB(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	return m.migrateCore.GetMigrations(ctx)
}

// RunMigrationWithCustomFunc - запускает миграцию с помощью пользовательской функции.
func (m *migrate) RunMigrationWithCustomFunc(
	ctx context.Context,
	migrateFunc CustomMigrateFunc,
	name string,
	version uint64,
	direction bool,
) error {
	closeFunc, err := m.migrateCore.ConnectDB(ctx)
	if err != nil {
		return err
	}
	defer closeFunc()
	tx, err := m.migrateCore.CreateTransactionalMigration(ctx, domain.Migration{Version: version, Name: name}, direction)
	if err != nil {
		if errors.Is(err, storage.ErrQueryNoAffectRows) {
			return nil
		}

		return err
	}
	sDirection := "Down"
	if direction {
		sDirection = "Up"
	}

	m.logger.Info(fmt.Sprintf("running %s migration with version %d (%s) ...", name, version, sDirection))

	return migrateFunc(ctx, tx)
}
