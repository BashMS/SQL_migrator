package storage

//go:generate mockery --case=underscore --output=. --inpackage --all

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"                //nolint:depguard
	"github.com/jackc/pgx/v4"                //nolint:depguard
	"github.com/jackc/pgx/v4/log/zapadapter" //nolint:depguard

	"go.uber.org/zap" //nolint:depguard

	"github.com/BashMS/SQL_migrator/pkg/config" //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/domain" //nolint:depguard
)

const (
	// MigrationsScheme - схема, где находится таблица миграция.
	MigrationsScheme = "public"
	// MigrationsTable - таблица миграции.
	MigrationsTable = "tmigration"

	connTimeout  = 2 * time.Second
	closeTimeout = 2 * time.Second
	checkTimeout = 200 * time.Millisecond

	fallbackLogLevel = pgx.LogLevelInfo
)

var (
	// ErrQueryNoAffectRows - запрос ничего не изменил.
	ErrQueryNoAffectRows = errors.New("query did not affect the rows")
	// ErrQueryDeadlineExceeded - время запроса истекло.
	ErrQueryDeadlineExceeded = errors.New("query deadline exceeded")
	// ErrLock - не удалось установить блокировку.
	ErrLock = errors.New("failed to apply lock")

	errVersionOrNameEmpty    = errors.New("version or migration name cannot be empty")
	errCreateStorage         = errors.New("failed to create table for migrations")
	errCheckStorage          = errors.New("failed to check table existence for migrations")
	errStartTransaction      = errors.New("failed to start transaction")
	errBeginMigration        = errors.New("failed begin migration")
	errCreateMigrationRecord = errors.New("failed to create migration record")
	errDNSEmpty              = errors.New("no DNS connection string")
)

type MigrateStorage interface {
	Connect(ctx context.Context) error
	Close()
	GetConnection(ctx context.Context) (*pgx.Conn, error)
	Stats(ctx context.Context) ([]domain.Migration, error)
	GetMigrationsByDirection(ctx context.Context, isApplied bool) (map[uint64]domain.Migration, error)
	BeginTxMigration(ctx context.Context, migration domain.Migration, direction bool) (pgx.Tx, error)
	RecentMigration(ctx context.Context) (domain.Migration, error)
	Lock(ctx context.Context, uid uint32) error
	UnLock(ctx context.Context) error
}

// postgresStorage слой для работы с БД.
type postgresStorage struct {
	storage MigrateStorage
	config  *config.Config
	conn    *pgx.Conn
	logger  *zap.Logger
}

// NewStorage.
func NewStorage(logger *zap.Logger, config *config.Config) MigrateStorage {
	return &postgresStorage{
		config: config,
		logger: logger,
	}
}

// Connect устанавливает соединение с БД.
func (ps *postgresStorage) Connect(ctx context.Context) error {
	var (
		err        error
		level      pgx.LogLevel
		connConfig *pgx.ConnConfig
	)
	if ps.config.DSN == "" {
		return errDNSEmpty
	}

	connConfig, err = pgx.ParseConfig(ps.config.DSN)
	if err != nil {
		return err
	}

	level, err = pgx.LogLevelFromString(ps.config.LogLevel)
	if err != nil {
		level = fallbackLogLevel
	}

	connConfig.Logger = zapadapter.NewLogger(ps.logger)
	connConfig.LogLevel = level
	connConfig.PreferSimpleProtocol = true
	connConfig.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
	}

	connCtx, cancelFunc := context.WithTimeout(ctx, connTimeout)
	defer cancelFunc()
	ps.conn, err = pgx.ConnectConfig(connCtx, connConfig)
	if err != nil {
		return err
	}

	if err = ps.provideStorage(ctx); err != nil {
		return err
	}

	return nil
}

func (ps *postgresStorage) GetConnection(ctx context.Context) (*pgx.Conn, error) {
	if ps.isClosed() {
		if err := ps.Connect(ctx); err != nil {
			return nil, err
		}
	}

	return ps.conn, nil
}

// Close закрывает соединение с БД.
func (ps *postgresStorage) Close() {
	if ps.conn == nil {
		return
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), closeTimeout)
	defer cancelFunc()

	err := ps.conn.Close(ctx)
	if err != nil {
		ps.logger.Error("failed to close postgresStorage connection correctly", zap.Error(err))
	}
	ps.conn = nil
	ps.storage = nil
}

func (ps *postgresStorage) BeginTxMigration(
	ctx context.Context,
	migration domain.Migration,
	direction bool,
) (pgx.Tx, error) {
	if ps.isClosed() {
		if err := ps.Connect(ctx); err != nil {
			return nil, err
		}
	}
	if err := ps.provideMigration(ctx, migration); err != nil {
		return nil, err
	}

	tx, err := ps.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w, %s", errStartTransaction, err.Error())
	}

	query := `
	WITH desiredMigration AS (
		SELECT version
		FROM "public"."tmigration"
		WHERE version = $1
		  AND is_applied = NOT $2
			FOR UPDATE
	)
	UPDATE "public"."tmigration" m
	SET is_applied = $2,
		update_at  = localtimestamp
	FROM desiredMigration
	WHERE m.version = desiredMigration.version
	RETURNING m.version;
`
	ctx, cancelFunc := context.WithTimeout(ctx, checkTimeout)
	defer cancelFunc()

	tag, err := tx.Exec(ctx, query, migration.Version, direction)
	if err != nil {
		if pgconn.Timeout(err) {
			return nil, ErrQueryDeadlineExceeded
		}

		return nil, fmt.Errorf("%w: %s", errBeginMigration, err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrQueryNoAffectRows
	}

	return tx, nil
}

func (ps *postgresStorage) RecentMigration(ctx context.Context) (domain.Migration, error) {
	if ps.isClosed() {
		if err := ps.Connect(ctx); err != nil {
			return domain.Migration{}, err
		}
	}
	var migration domain.Migration
	query := `
	SELECT version, name, is_applied, update_at  
	FROM "public"."tmigration" 
	WHERE is_applied = TRUE
	ORDER BY version DESC 
	LIMIT 1; 
`
	if err := ps.conn.QueryRow(ctx, query).Scan(
		&migration.Version,
		&migration.Name,
		&migration.IsApplied,
		&migration.UpdateAt); err != nil {
		return migration, err
	}

	return migration, nil
}

func (ps *postgresStorage) GetMigrationsByDirection(
	ctx context.Context,
	isApplied bool,
) (map[uint64]domain.Migration, error) {
	if ps.isClosed() {
		if err := ps.Connect(ctx); err != nil {
			return nil, err
		}
	}
	query := `
	SELECT version, name, is_applied, update_at 
	FROM "public"."tmigration" 
	WHERE is_applied = $1
	ORDER BY version DESC;
`
	rows, err := ps.conn.Query(ctx, query, isApplied)
	if err != nil {
		return nil, err
	}

	migrations := make(map[uint64]domain.Migration)
	for rows.Next() {
		var migration domain.Migration
		if err := rows.Scan(
			&migration.Version,
			&migration.Name,
			&migration.IsApplied,
			&migration.UpdateAt); err != nil {
			return nil, err
		}
		migrations[migration.Version] = migration
	}

	return migrations, nil
}

func (ps *postgresStorage) Stats(ctx context.Context) ([]domain.Migration, error) {
	if ps.isClosed() {
		if err := ps.Connect(ctx); err != nil {
			return nil, err
		}
	}
	query := `
	SELECT version, name, is_applied, update_at 
	FROM "public"."tmigration"
	ORDER BY version;
`
	rows, err := ps.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var stats []domain.Migration
	for rows.Next() {
		var migration domain.Migration
		if err := rows.Scan(
			&migration.Version,
			&migration.Name,
			&migration.IsApplied,
			&migration.UpdateAt); err != nil {
			return nil, err
		}

		stats = append(stats, migration)
	}

	return stats, nil
}

func (ps *postgresStorage) Lock(ctx context.Context, uid uint32) error {
	if ps.isClosed() {
		if err := ps.Connect(ctx); err != nil {
			return err
		}
	}
	var isLocked bool
	if err := ps.conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", uid).Scan(&isLocked); err != nil {
		return err
	}

	if !isLocked {
		return ErrLock
	}

	return nil
}

func (ps *postgresStorage) UnLock(ctx context.Context) error {
	if ps.isClosed() {
		if err := ps.Connect(ctx); err != nil {
			return err
		}
	}
	if _, err := ps.conn.Exec(ctx, "SELECT pg_advisory_unlock_all()"); err != nil {
		return err
	}

	return nil
}

// provideStorage - создает таблицу для контроля миграций.
func (ps *postgresStorage) provideStorage(ctx context.Context) error {
	ok, err := ps.checkStorage(ctx)
	if err != nil {
		return fmt.Errorf("%w: %s", errCheckStorage, err.Error())
	}
	if !ok {
		query := `
	CREATE TABLE IF NOT EXISTS "public"."tmigration" (
	    version BIGINT NOT NULL,
		name VARCHAR(255) NOT NULL,
		is_applied BOOLEAN NOT NULL,
		update_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now()
	);
	CREATE UNIQUE INDEX IF NOT EXISTS uidx_version  ON "public"."tmigration" USING  btree(version);
	CREATE INDEX IF NOT EXISTS idx_applied_version ON "public"."tmigration" USING btree(is_applied, version);
`
		if _, err := ps.conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("%w: %s", errCreateStorage, err.Error())
		}
	}

	return nil
}

func (ps *postgresStorage) provideMigration(ctx context.Context, migration domain.Migration) error {
	if migration.Name == "" || migration.Version == 0 {
		return fmt.Errorf("%w: version = '%d', name = '%s'",
			errVersionOrNameEmpty, migration.Version, migration.Name)
	}

	query := `
	DO
	$func$
		DECLARE
			_version BIGINT;
		BEGIN
			SELECT version INTO _version FROM tmigration WHERE version = $1;
			IF NOT FOUND THEN
				INSERT INTO "public"."tmigration" (version, name, is_applied) VALUES ($1, $2, FALSE);
				RAISE NOTICE 'New migration record added';
			END IF;
		END;
	$func$;
`
	_, err := ps.conn.Exec(ctx, query, migration.Version, migration.Name)
	if err != nil {
		return fmt.Errorf("%w: %s", errCreateMigrationRecord, err.Error())
	}

	return nil
}

func (ps *postgresStorage) isClosed() bool {
	return ps.conn == nil || ps.conn.IsClosed()
}

func (ps *postgresStorage) checkStorage(ctx context.Context) (bool, error) {
	query := `
	SELECT EXISTS (
   		SELECT FROM information_schema.tables 
   		WHERE table_schema = $1 AND table_name = $2
   );
`
	var ok bool
	if err := ps.conn.QueryRow(ctx, query, MigrationsScheme, MigrationsTable).Scan(&ok); err != nil {
		return false, err
	}

	return ok, nil
}
