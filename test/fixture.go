package test

import (
	"path/filepath"
	"time"

	"github.com/BashMS/SQL_migrator/internal/loader"
	"github.com/BashMS/SQL_migrator/pkg/config"
	"github.com/BashMS/SQL_migrator/pkg/domain"
)

func GetMigrationByVersion(version uint64, IsApplied bool) domain.Migration {
	if version > 5 {
		return domain.Migration{}
	}

	return domain.Migration{
		Version:   version,
		Name:      GetRawMigrationByVersion(version).Name,
		IsApplied: IsApplied,
		UpdateAt:  time.Time{},
	}
}

func GetRawMigrationByVersion(version uint64) loader.RawMigration {
	return RawSQLMigrations(&config.Config{}, true)[version-1 : version][0]
}

func RawSQLMigrations(cfg *config.Config, direction bool) []loader.RawMigration {
	if direction {
		return []loader.RawMigration{
			{
				Version:   1,
				Name:      "testCreateFirstTable",
				PathUp:    filepath.Join(cfg.Path, "1_test_create_first_table.up.sql"),
				PathDown:  filepath.Join(cfg.Path, "1_test_create_first_table.down.sql"),
				Format:    config.FormatSQL,
				QueryUp:   `CREATE TABLE IF NOT EXISTS "test_first_table"();`,
				QueryDown: `DROP TABLE IF EXISTS "test_first_table";`,
			},
			{
				Version:   2,
				Name:      "testCreateSecondTable",
				PathUp:    filepath.Join(cfg.Path, "2_test_create_second_table.up.sql"),
				PathDown:  filepath.Join(cfg.Path, "2_test_create_second_table.down.sql"),
				Format:    config.FormatSQL,
				QueryUp:   `CREATE TABLE IF NOT EXISTS "test_second_table"();`,
				QueryDown: `DROP TABLE IF EXISTS "test_second_table";`,
			},
			{
				Version:   3,
				Name:      "testCreateThirdTable",
				PathUp:    filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.up.sql"),
				PathDown:  filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.down.sql"),
				Format:    config.FormatSQL,
				QueryUp:   `CREATE TABLE IF NOT EXISTS "test_third_table"();`,
				QueryDown: `DROP TABLE IF EXISTS "test_third_table";`,
			},
			{
				Version:   4,
				Name:      "testEmptyMigration",
				PathUp:    filepath.Join(cfg.Path, "4_test_empty_migration.up.sql"),
				PathDown:  filepath.Join(cfg.Path, "4_test_empty_migration.down.sql"),
				Format:    config.FormatSQL,
				QueryUp:   "",
				QueryDown: "",
			},
			{
				Version:   5,
				Name:      "testErrorMigration",
				PathUp:    filepath.Join(cfg.Path, "5_test_error_migration.up.sql"),
				PathDown:  filepath.Join(cfg.Path, "5_test_error_migration.down.sql"),
				Format:    config.FormatSQL,
				QueryUp:   `SELECT Bad_Migratiom FORM MORF;`,
				QueryDown: `SELECT Bad_Migratiom FORM MORF;`,
			},
		}
	}

	return []loader.RawMigration{
		{
			Version:   5,
			Name:      "testErrorMigration",
			PathUp:    filepath.Join(cfg.Path, "5_test_error_migration.up.sql"),
			PathDown:  filepath.Join(cfg.Path, "5_test_error_migration.down.sql"),
			Format:    config.FormatSQL,
			QueryUp:   `SELECT Bad_Migratiom FORM MORF;`,
			QueryDown: `SELECT Bad_Migratiom FORM MORF;`,
		},
		{
			Version:   4,
			Name:      "testEmptyMigration",
			PathUp:    filepath.Join(cfg.Path, "4_test_empty_migration.up.sql"),
			PathDown:  filepath.Join(cfg.Path, "4_test_empty_migration.down.sql"),
			Format:    config.FormatSQL,
			QueryUp:   "",
			QueryDown: "",
		},
		{
			Version:   3,
			Name:      "testCreateThirdTable",
			PathUp:    filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.up.sql"),
			PathDown:  filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.down.sql"),
			Format:    config.FormatSQL,
			QueryUp:   `CREATE TABLE IF NOT EXISTS "test_third_table"();`,
			QueryDown: `DROP TABLE IF EXISTS "test_third_table";`,
		},
		{
			Version:   2,
			Name:      "testCreateSecondTable",
			PathUp:    filepath.Join(cfg.Path, "2_test_create_second_table.up.sql"),
			PathDown:  filepath.Join(cfg.Path, "2_test_create_second_table.down.sql"),
			Format:    config.FormatSQL,
			QueryUp:   `CREATE TABLE IF NOT EXISTS "test_second_table"();`,
			QueryDown: `DROP TABLE IF EXISTS "test_second_table";`,
		},
		{
			Version:   1,
			Name:      "testCreateFirstTable",
			PathUp:    filepath.Join(cfg.Path, "1_test_create_first_table.up.sql"),
			PathDown:  filepath.Join(cfg.Path, "1_test_create_first_table.down.sql"),
			Format:    config.FormatSQL,
			QueryUp:   `CREATE TABLE IF NOT EXISTS "test_first_table"();`,
			QueryDown: `DROP TABLE IF EXISTS "test_first_table";`,
		},
	}
}

func RawGoMigrations(cfg *config.Config, direction bool) []loader.RawMigration {
	if direction {
		return []loader.RawMigration{
			{
				Version:  1,
				Name:     "testCreateFirstTable",
				PathUp:   filepath.Join(cfg.Path, "1_test_create_first_table.go"),
				PathDown: filepath.Join(cfg.Path, "1_test_create_first_table.go"),
				Format:   config.FormatGolang,
			},
			{
				Version:  2,
				Name:     "testCreateSecondTable",
				PathUp:   filepath.Join(cfg.Path, "2_test_create_second_table.go"),
				PathDown: filepath.Join(cfg.Path, "2_test_create_second_table.go"),
				Format:   config.FormatGolang,
			},
			{
				Version:  3,
				Name:     "testCreateThirdTable",
				PathUp:   filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.go"),
				PathDown: filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.go"),
				Format:   config.FormatGolang,
			},
			{
				Version:  4,
				Name:     "testEmptyMigration",
				PathUp:   filepath.Join(cfg.Path, "4_test_empty_migration.go"),
				PathDown: filepath.Join(cfg.Path, "4_test_empty_migration.go"),
				Format:   config.FormatGolang,
			},
			{
				Version:  5,
				Name:     "testErrorMigration",
				PathUp:   filepath.Join(cfg.Path, "5_test_error_migration.go"),
				PathDown: filepath.Join(cfg.Path, "5_test_error_migration.go"),
				Format:   config.FormatGolang,
			},
		}
	}

	return []loader.RawMigration{
		{
			Version:  5,
			Name:     "testErrorMigration",
			PathUp:   filepath.Join(cfg.Path, "5_test_error_migration.go"),
			PathDown: filepath.Join(cfg.Path, "5_test_error_migration.go"),
			Format:   config.FormatGolang,
		},
		{
			Version:  4,
			Name:     "testEmptyMigration",
			PathUp:   filepath.Join(cfg.Path, "4_test_empty_migration.go"),
			PathDown: filepath.Join(cfg.Path, "4_test_empty_migration.go"),
			Format:   config.FormatGolang,
		},
		{
			Version:  3,
			Name:     "testCreateThirdTable",
			PathUp:   filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.go"),
			PathDown: filepath.Join(cfg.Path, "subfolder/3_test_create_third_table.go"),
			Format:   config.FormatGolang,
		},
		{
			Version:  2,
			Name:     "testCreateSecondTable",
			PathUp:   filepath.Join(cfg.Path, "2_test_create_second_table.go"),
			PathDown: filepath.Join(cfg.Path, "2_test_create_second_table.go"),
			Format:   config.FormatGolang,
		},
		{
			Version:  1,
			Name:     "testCreateFirstTable",
			PathUp:   filepath.Join(cfg.Path, "1_test_create_first_table.go"),
			PathDown: filepath.Join(cfg.Path, "1_test_create_first_table.go"),
			Format:   config.FormatGolang,
		},
	}
}
