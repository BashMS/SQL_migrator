package core_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BashMS/SQL_migrator/internal/command" //nolint:depguard
	"github.com/BashMS/SQL_migrator/internal/core"    //nolint:depguard
	"github.com/BashMS/SQL_migrator/internal/loader"  //nolint:depguard
	"github.com/BashMS/SQL_migrator/internal/storage" //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/config"       //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/domain"       //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/migrate"      //nolint:depguard
	"github.com/BashMS/SQL_migrator/test"             //nolint:depguard
	"github.com/jackc/pgconn"                         //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest" //nolint:depguard
)

const defaultMigratePath = "./../../test/data"

func TestMigrateCore_CreateMigrationFile(t *testing.T) {
	zLogger := zaptest.NewLogger(t)
	mockStorage := storage.MockMigrateStorage{}
	mockCommand := command.MockCommand{}

	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	tCases := []struct {
		name          string
		format        string
		giveName      string
		giveVersion   uint64
		expectedFiles []string
		expectedErr   error
	}{
		{
			name:        "bad sql name",
			format:      config.FormatSQL,
			giveName:    "bad name *?$%:+=1",
			giveVersion: 1,
			expectedFiles: []string{
				filepath.Join(tmpDir, "1_bad_name________1.up.sql"),
				filepath.Join(tmpDir, "1_bad_name________1.down.sql"),
			},
		},
		{
			name:        "bad go name",
			format:      config.FormatGolang,
			giveName:    "bad name *?$%:+=1",
			giveVersion: 1,
			expectedFiles: []string{
				filepath.Join(tmpDir, "1_bad_name________1.go"),
			},
		},
		{
			name:          "zero version",
			format:        config.FormatGolang,
			giveName:      "zero version",
			giveVersion:   0,
			expectedFiles: []string{},
			expectedErr:   domain.ErrMigrateVersionIncorrect,
		},
	}
	config := createConfig(t, tmpDir)
	migrateCore := core.NewMigrateCore(&mockStorage, &mockCommand, zLogger, config)

	for _, tCase := range tCases {
		t.Run(tCase.name, func(t *testing.T) {
			config.Format = tCase.format
			err := migrateCore.CreateMigrationFile(tCase.giveName, tCase.giveVersion)
			if tCase.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, err, tCase.expectedErr)
			}
			for _, expectedFile := range tCase.expectedFiles {
				assert.FileExists(t, expectedFile)
			}
		})
	}
}

func TestNewMigrateCore_CreateGoTemplate(t *testing.T) {
	zLogger := zaptest.NewLogger(t)
	mockStorage := storage.MockMigrateStorage{}
	mockCommand := command.MockCommand{}

	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	cfg := createConfig(t, tmpDir)
	cfg.Format = config.FormatGolang
	migrateCore := core.NewMigrateCore(&mockStorage, &mockCommand, zLogger, cfg)

	err := migrateCore.CreateMigrationFile("Test empty migration", 4)
	assert.NoError(t, err)
	newFile := filepath.Join(tmpDir, "4_test_empty_migration.go")
	if !assert.FileExists(t, newFile) {
		return
	}

	defaultPath, err := filepath.Abs(defaultMigratePath)
	assert.NoError(t, err)
	originalFile := filepath.Join(defaultPath, "4_test_empty_migration.go")
	if !assert.FileExists(t, originalFile) {
		return
	}

	assertCompareFiles(t, originalFile, newFile)
}

func TestMigrateCore_LoadMigrations(t *testing.T) {
	zLogger := zaptest.NewLogger(t)
	cfg := createConfig(t, defaultMigratePath)
	mockCommand := command.MockCommand{}

	dbMigrations := map[uint64]domain.Migration{
		1: test.GetMigrationByVersion(1, true),
		2: test.GetMigrationByVersion(2, true),
	}

	mockStorage := storage.MockMigrateStorage{}
	mockStorage.On("GetMigrationsByDirection", mock.Anything, migrate.MigrationUp).
		Return(dbMigrations, nil)
	mockStorage.On("GetMigrationsByDirection", mock.Anything, migrate.MigrationDown).
		Return(map[uint64]domain.Migration{}, nil)
	mockStorage.On("RecentMigration", mock.Anything).Return(test.GetMigrationByVersion(2, true), nil)

	tCases := []struct {
		name                  string
		format                string
		giveRequestToVersion  uint64
		giveDirection         bool
		expectedRawMigrations []loader.RawMigration
	}{
		// sql migration
		{
			name:                  "load all UP sql-file",
			format:                config.FormatSQL,
			giveRequestToVersion:  0,
			giveDirection:         migrate.MigrationUp,
			expectedRawMigrations: test.RawSQLMigrations(cfg, migrate.MigrationUp)[2:],
		},
		{
			name:                  "load all DOWN sql-file",
			format:                config.FormatSQL,
			giveRequestToVersion:  0,
			giveDirection:         migrate.MigrationDown,
			expectedRawMigrations: test.RawSQLMigrations(cfg, migrate.MigrationDown)[3:],
		},
		{
			name:                  "checking sql-file loading inclusively up to version",
			format:                config.FormatSQL,
			giveRequestToVersion:  3,
			giveDirection:         migrate.MigrationUp,
			expectedRawMigrations: test.RawSQLMigrations(cfg, migrate.MigrationUp)[2:3],
		},
		// go migration
		{
			name:                  "load all UP go-file",
			format:                config.FormatGolang,
			giveRequestToVersion:  0,
			giveDirection:         migrate.MigrationUp,
			expectedRawMigrations: test.RawGoMigrations(cfg, migrate.MigrationUp)[2:],
		},
		{
			name:                  "checking go-file loading inclusively up to version",
			format:                config.FormatGolang,
			giveRequestToVersion:  3,
			giveDirection:         migrate.MigrationUp,
			expectedRawMigrations: test.RawGoMigrations(cfg, migrate.MigrationUp)[2:3],
		},
	}

	migrateCore := core.NewMigrateCore(&mockStorage, &mockCommand, zLogger, cfg)

	for _, tCase := range tCases {
		t.Run(tCase.name, func(t *testing.T) {
			cfg.Format = tCase.format
			rawMigrations, err := migrateCore.LoadMigrations(
				context.Background(),
				tCase.giveRequestToVersion,
				tCase.giveDirection)

			assert.NoError(t, err)
			assert.EqualValuesf(t, tCase.expectedRawMigrations, rawMigrations, "loaded migrations do not match")
		})
	}
}

func TestMigrateCore_StartMigrate_FormatGolang(t *testing.T) { //nolint:gocognit
	zLogger := zaptest.NewLogger(t)
	cfg := createConfig(t, defaultMigratePath)
	cfg.Format = config.FormatGolang
	mockStorage := storage.MockMigrateStorage{}

	tCases := []struct {
		name                 string
		giveNeededMigrations []loader.RawMigration
		giveDirection        bool
		expectedFiles        []string
		expectedErr          error
		expectedCount        int
	}{
		{
			name:                 "migration up",
			giveNeededMigrations: test.RawGoMigrations(cfg, migrate.MigrationUp),
			giveDirection:        migrate.MigrationUp,

			expectedFiles: []string{
				"main.go",
				"1_test_create_first_table.go",
				"2_test_create_second_table.go",
				"3_test_create_third_table.go",
				"4_test_empty_migration.go",
				"5_test_error_migration.go",
			},
			expectedErr:   nil,
			expectedCount: 5,
		},
		{
			name:                 "migration down",
			giveNeededMigrations: test.RawGoMigrations(cfg, migrate.MigrationDown)[1:],
			giveDirection:        migrate.MigrationDown,

			expectedFiles: []string{
				"main.go",
				"1_test_create_first_table.go",
				"2_test_create_second_table.go",
				"3_test_create_third_table.go",
				"4_test_empty_migration.go",
			},
			expectedErr:   nil,
			expectedCount: 4,
		},
		{
			name:                 "migration error",
			giveNeededMigrations: test.RawGoMigrations(cfg, migrate.MigrationUp)[0:1],
			giveDirection:        migrate.MigrationUp,

			expectedFiles: []string{
				"main.go",
				"1_test_create_first_table.go",
			},
			expectedErr:   fmt.Errorf("%w: error", domain.ErrStartingProgramForMigrations),
			expectedCount: 0,
		},
	}

	for _, tCase := range tCases {
		t.Run(tCase.name, func(t *testing.T) {
			var returnErr error
			if tCase.expectedErr != nil {
				returnErr = fmt.Errorf("error")
			}

			mockCommand := command.MockCommand{}
			mockCommand.On(
				"Run",
				mock.Anything,
				"go",
				command.Args{"mod", "init", "go/migration"},
				mock.Anything,
				mock.Anything).
				Return(nil)
			mockCommand.On(
				"Run",
				mock.Anything,
				"go",
				command.Args{"mod", "tidy"},
				mock.Anything,
				mock.Anything).
				Return(nil)

			mockCommand.On(
				"RunWithGracefulShutdown",
				mock.Anything,
				"go",
				command.Args{"run", "./..."}, mock.Anything,
				mock.Anything).
				Run(func(args mock.Arguments) {
					if len(args) < 5 {
						t.Fatal("the number of arguments in the command.Run method is less than 5")
					}

					dir, ok := args[3].(string)
					if !ok || dir == "" {
						t.Fatal("in command.Run command, temporary directory is empty")
					}

					for _, expectedFile := range tCase.expectedFiles {
						newFile := filepath.Join(dir, expectedFile)
						if !assert.FileExists(t, newFile) {
							return
						}

						if expectedFile != "main.go" {
							if expectedFile == "3_test_create_third_table.go" {
								expectedFile = "third_table/3_test_create_third_table.go"
							}
							originalFile := filepath.Join(cfg.Path, expectedFile)
							if !assert.FileExists(t, originalFile) {
								t.Fatalf("original file not found %s", expectedFile)
							}

							assertCompareFiles(t, originalFile, newFile)
						}
					}
				}).Return(returnErr)

			migrateCore := core.NewMigrateCore(&mockStorage, &mockCommand, zLogger, cfg)
			count, err := migrateCore.StartMigrate(context.Background(), tCase.giveNeededMigrations, tCase.giveDirection)
			if tCase.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tCase.expectedErr, err)
			}

			assert.Equal(t, tCase.expectedCount, count)
		})
	}
}

func TestMigrateCore_StartMigrate_FormatSQL(t *testing.T) {
	zLogger := zaptest.NewLogger(t)
	cfg := createConfig(t, defaultMigratePath)
	cfg.Format = config.FormatSQL
	mockCommand := command.MockCommand{}

	tCases := []struct {
		name                 string
		giveNeededMigrations []loader.RawMigration
		giveDirection        bool
		expectedCount        int
	}{
		{
			name:                 "",
			giveNeededMigrations: test.RawSQLMigrations(cfg, migrate.MigrationUp),
			giveDirection:        migrate.MigrationUp,
			expectedCount:        4,
		},
		{
			name:                 "",
			giveNeededMigrations: test.RawSQLMigrations(cfg, migrate.MigrationDown),
			giveDirection:        migrate.MigrationDown,
			expectedCount:        5,
		},
	}

	for _, tCase := range tCases {
		t.Run(tCase.name, func(t *testing.T) {
			var currentMigration domain.Migration
			mockTx := test.MockTx{}
			mockTx.On("Commit", mock.Anything).Return(nil)
			mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					if len(args) < 2 {
						t.Fatal("the number of arguments in the tx.Exec method is less than 3")
					}
					if currentMigration.Version == 0 {
						t.Fatal("current migration not set")
					}

					actualQuery, ok := args[1].(string)
					if !ok {
						t.Fatal("failed to get the current version of the request")
					}

					var expectedQuery string
					if tCase.giveDirection {
						expectedQuery = test.GetRawMigrationByVersion(currentMigration.Version).QueryUp
					} else {
						expectedQuery = test.GetRawMigrationByVersion(currentMigration.Version).QueryDown
					}
					assert.EqualValues(t, expectedQuery, actualQuery)
				}).Return(pgconn.CommandTag{}, nil)

			mockStorage := storage.MockMigrateStorage{}
			mockStorage.On("BeginTxMigration", mock.Anything, mock.Anything, tCase.giveDirection).
				Run(func(args mock.Arguments) {
					if len(args) < 3 {
						t.Fatal("the number of arguments in the storage.BeginTxMigration method is less than 3")
					}
					var ok bool
					currentMigration, ok = args[1].(domain.Migration)
					if !ok {
						t.Fatal("failed to get current migration")
					}
				}).Return(&mockTx, nil)

			migrateCore := core.NewMigrateCore(&mockStorage, &mockCommand, zLogger, cfg)
			count, err := migrateCore.StartMigrate(context.Background(), tCase.giveNeededMigrations, tCase.giveDirection)
			assert.NoError(t, err)
			assert.Equal(t, tCase.expectedCount, count)
		})
	}
}

func assertCompareFiles(t *testing.T, originalFile, newFile string) {
	t.Helper()
	assert.EqualValues(t, fileGetContents(t, originalFile), fileGetContents(t, newFile))
}

func fileGetContents(t *testing.T, pathFile string) []byte {
	t.Helper()
	data, err := os.ReadFile(pathFile)
	if !assert.NoError(t, err) {
		return []byte{}
	}

	return data
}

func createTempDir(t *testing.T) string {
	t.Helper()
	tmpPath, err := os.MkdirTemp(os.TempDir(), "test_migrator_*")
	assert.NoError(t, err)

	return tmpPath
}

func createConfig(t *testing.T, migratePath string) *config.Config {
	t.Helper()
	config := config.Config{
		DSN:      "dsn",
		Path:     migratePath,
		Format:   "sql",
		LogLevel: config.LogLevelError,
	}

	err := config.PathConversion()
	assert.NoError(t, err)

	return &config
}
