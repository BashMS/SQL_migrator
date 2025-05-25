package loader

import "github.com/BashMS/SQL_migrator/pkg/domain" //nolint:depguard

// Filter.
type Filter struct {
	Exclude          map[uint64]domain.Migration
	Recent           domain.Migration
	RequestToVersion uint64
}

// IsExcluded - проверяет, была ли миграция добавлена ​​в исключенные.
func (f *Filter) IsExcluded(migration RawMigration) bool {
	exclMigration, ok := f.Exclude[migration.Version]
	if ok && exclMigration.Name == migration.Name {
		return true
	}

	return false
}

// AllowUp - проверяет, разрешено ли накатывать версию миграции.
func (f *Filter) AllowUp(migration RawMigration) bool {
	if f.RequestToVersion != 0 && migration.Version > f.RequestToVersion {
		return false
	} else if f.Recent.Version != 0 && migration.Version <= f.Recent.Version {
		return false
	}

	return true
}

// AllowDown - проверяет, разрешен ли откат версии миграции.
func (f *Filter) AllowDown(migration RawMigration) bool {
	if f.RequestToVersion != 0 && migration.Version < f.RequestToVersion {
		return false
	} else if f.Recent.Version != 0 && migration.Version > f.Recent.Version {
		return false
	}

	return true
}
