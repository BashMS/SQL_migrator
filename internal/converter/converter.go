package converter

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/BashMS/SQL_migrator/pkg/config" //nolint:depguard
)

// SanitizeMigrationName - очищает имя миграции, удаляя или заменяя лишние символы.
func SanitizeMigrationName(name string) string {
	name = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}

		return config.Separator
	}, name)
	if strings.HasSuffix(name, "test") {
		name = fmt.Sprintf("%s%s", name, string(config.Separator))
	}
	return name
}

// VersionToUint - преобразует версию в число.
func VersionToUint(version string) (uint64, error) {
	return strconv.ParseUint(version, 10, 64)
}
