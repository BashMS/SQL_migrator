package domain

import "errors"

var (
	// ErrConnection - невозможно подключиться к базе данных.
	ErrConnection = errors.New("unable to connect to database")
	// ErrInvalidFormat - неверный формат миграции.
	ErrInvalidFormat = errors.New("invalid migration format specified")

	// ErrMigrationFileExists - файл миграции уже существует.
	ErrMigrationFileExists = errors.New("migration file already exists")
	// ErrCreateMigrationFile - не удалось создать файл миграции.
	ErrCreateMigrationFile = errors.New("failed to create migration file")
	// ErrMigrationNameRequired - имя миграции не указанно.
	ErrMigrationNameRequired = errors.New("migration name is required")

	// ErrMigrateVersionIncorrect - версия миграции должна быть больше нуля.
	ErrMigrateVersionIncorrect = errors.New("migration version must be greater than zero")
	// ErrTransactionCancel - ошибка отмены транзакции.
	ErrTransactionCancel = errors.New("transaction cancellation error")
	// ErrApplyingMigration - ошибка наката миграции.
	ErrApplyingMigration = errors.New("error applying migration")
	// ErrParallelApp - приложение миграций выполняется параллельно.
	ErrParallelApp = errors.New(`the response time from the database has expired, 
it is possible that the application of migrations is running in parallel`)
	// ErrGetRecentMigration - не удалось получить последнюю версию миграции.
	ErrGetRecentMigration = errors.New("failed to get the recent migration version")
	// ErrLoadMigrations - не удалось загрузить миграции.
	ErrLoadMigrations = errors.New("failed to load migrations")
	// ErrBuildProgramForMigrations - ошибка при сборке программы для миграций.
	ErrBuildProgramForMigrations = errors.New("error while building the program for migrations")
	// ErrStartingProgramForMigrations - ошибка при запуске программы для миграций.
	ErrStartingProgramForMigrations = errors.New("an error occurred while starting the program for migrations")
)
