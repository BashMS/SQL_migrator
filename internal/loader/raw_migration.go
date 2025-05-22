package loader

//RawMigration.
type RawMigration struct {
	Version   uint64
	Name      string
	PathUp    string
	PathDown  string
	Format    string
	QueryUp   string
	QueryDown string
}

//GetPath - возвращает путь в зависимости от направления миграции.
func (rm *RawMigration) GetPath(direction bool) string {
	if direction {
		return rm.PathUp
	}

	return rm.PathDown
}

//GetQuery - возвращает запрос в зависимости от направления.
func (rm *RawMigration) GetQuery(direction bool) string {
	if direction {
		return rm.QueryUp
	}

	return rm.QueryDown
}
