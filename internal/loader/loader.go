package loader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/iancoleman/strcase"
	"github.com/BashMS/SQL_migrator/internal/converter"
	"github.com/BashMS/SQL_migrator/pkg/config"
	"github.com/BashMS/SQL_migrator/pkg/domain"
	"go.uber.org/zap"
)

var (
	// ErrMigrationPath - неверный путь миграции.
	ErrMigrationPath = errors.New("migration path is not specified or it is incorrect")
	// ErrPostfix - постфикс для миграции не найден (.down или .up).
	ErrPostfix = errors.New("postfix not found for migration (.down or .up)")
	// ErrSeparatorNotFound - не найден разделитель.
	ErrSeparatorNotFound = errors.New("no separator found")
	// ErrMigrationsSameName - миграции sql (down и up) должны иметь одинаковое имя.
	ErrMigrationsSameName = errors.New("sql migrations (down and up) must have the same name")
	// ErrMigrationVersionUnique - версия миграции должна быть уникальной.
	ErrMigrationVersionUnique = errors.New("migration version must be unique")
	// ErrReadFile - ошибка чтения файла.
	ErrReadFile = errors.New("error reading file")
	// ErrSkipFile - пропустить этот файл.
	ErrSkipFile = errors.New("skip this file")
	// ErrMigrateVersionFile - версия должна быть больше 0 в файле миграции.
	ErrMigrateVersionFile = errors.New("version must be greater than 0 in the migration file")
)

// Loader.
type Loader struct {
	logger         *zap.Logger
	allowExt       string
	format         string
	listMigrations []RawMigration
	hash           map[uint64]int
}

// NewLoader конструктор.
func NewLoader(logger *zap.Logger) Loader {
	return Loader{logger: logger, allowExt: config.ExtSQL, format: config.FormatSQL}
}

// SetFormat - устанавливает формат миграции.
func (l *Loader) SetFormat(format string) {
	l.format = format
	switch format {
	case config.FormatSQL:
		l.allowExt = config.ExtSQL
	case config.FormatGolang:
		l.allowExt = config.ExtGolang
	}
}

// LoadMigrations - загружает все миграции (с фильтром).
func (l *Loader) LoadMigrations(ctx context.Context, filter Filter, path string, direction bool) ([]RawMigration, error) {
	l.resetMigrations()

	if !fileutil.Exist(path) {
		return nil, ErrMigrationPath
	}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return context.DeadlineExceeded
		default:
		}

		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}

		migration, err := l.parseFile(filePath)
		if errors.Is(err, ErrSkipFile) {
			l.logger.Debug(fmt.Sprintf("skipped %s file", filePath))
			return nil
		} else if err != nil {
			return err
		}

		if filter.IsExcluded(migration) ||
			(direction && !filter.AllowUp(migration)) ||
			(!direction && !filter.AllowDown(migration)) {
			l.logger.Debug(fmt.Sprintf("%s file not loaded", filePath))
			return nil
		}

		return l.addMigration(migration)
	})
	if err != nil {
		return nil, err
	}

	if len(l.listMigrations) == 0 {
		return l.listMigrations, nil
	}

	if direction {
		sort.Sort(l)
	} else {
		sort.Sort(sort.Reverse(l))
	}

	return l.listMigrations, nil
}

func (l *Loader) addMigration(migration RawMigration) error {
	idx, ok := l.hash[migration.Version]
	if ok {
		if l.listMigrations[idx].Format == config.FormatGolang {
			return ErrMigrationVersionUnique
		}
		if err := l.mergeMigration(idx, migration); err != nil {
			return err
		}

		return nil
	}

	l.listMigrations = append(l.listMigrations, migration)
	l.hash[migration.Version] = len(l.listMigrations) - 1

	return nil
}

func (l *Loader) resetMigrations() {
	l.listMigrations = []RawMigration{}
	l.hash = make(map[uint64]int)
}

func (l *Loader) mergeMigration(idx int, migration RawMigration) error {
	if l.listMigrations[idx].Format == config.FormatSQL && l.listMigrations[idx].Name != migration.Name {
		return fmt.Errorf(
			"%w: %s and %s are different names for version %d",
			ErrMigrationsSameName,
			l.listMigrations[idx].Name,
			migration.Name, migration.Version,
		)
	}

	if l.listMigrations[idx].PathDown == "" {
		l.listMigrations[idx].PathDown = migration.PathDown
	}

	if l.listMigrations[idx].PathUp == "" {
		l.listMigrations[idx].PathUp = migration.PathUp
	}

	if l.listMigrations[idx].QueryUp == "" {
		l.listMigrations[idx].QueryUp = migration.QueryUp
	}

	if l.listMigrations[idx].QueryDown == "" {
		l.listMigrations[idx].QueryDown = migration.QueryDown
	}

	return nil
}

func (l *Loader) parseFile(path string) (RawMigration, error) {
	var (
		migration    RawMigration
		idxDirection int
		err          error
	)
	name := filepath.Base(path)

	ext := filepath.Ext(name)
	if ext != l.allowExt {
		return migration, ErrSkipFile
	}
	migration.Format = l.format

	switch migration.Format {
	case config.FormatGolang:
		idxDirection = strings.LastIndex(name, ext)
		migration.PathUp = path
		migration.PathDown = path
	case config.FormatSQL:
		query, err := os.ReadFile(path)
		if err != nil {
			return migration, fmt.Errorf("%w %s", ErrReadFile, path)
		}

		if idxDirection = strings.LastIndex(name, config.PostfixUp); idxDirection > 0 {
			migration.PathUp = path
			migration.QueryUp = string(query)
		} else if idxDirection = strings.LastIndex(name, config.PostfixDown); idxDirection > 0 {
			migration.PathDown = path
			migration.QueryDown = string(query)
		} else {
			return migration, ErrPostfix
		}
	default:
		return migration, fmt.Errorf("%w %s", domain.ErrInvalidFormat, path)
	}

	idx := strings.Index(name, string(config.Separator))
	if idx < 0 {
		return migration, ErrSeparatorNotFound
	}

	migration.Name = converter.SanitizeMigrationName(strcase.ToLowerCamel(name[idx+1 : idxDirection]))
	migration.Version, err = converter.VersionToUint(name[:idx])
	if err != nil || migration.Version == 0 {
		return migration, fmt.Errorf("%w (%s)", ErrMigrateVersionFile, path)
	}

	return migration, nil
}

func (l Loader) Len() int {
	return len(l.listMigrations)
}

// Swap меняет местами элементы с индексами i и j.
func (l Loader) Swap(i, j int) {
	l.listMigrations[i], l.listMigrations[j] = l.listMigrations[j], l.listMigrations[i]
	l.hash[l.listMigrations[i].Version], l.hash[l.listMigrations[j].Version] =
		l.hash[l.listMigrations[j].Version], l.hash[l.listMigrations[i].Version]
}

// Less сообщает, следует ли сортировать элемент с индексом i перед элементом с индексом j.
func (l Loader) Less(i, j int) bool {
	return l.listMigrations[i].Version < l.listMigrations[j].Version
}

