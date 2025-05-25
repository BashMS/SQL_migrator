package report

import (
	"fmt"

	"github.com/BashMS/SQL_migrator/pkg/domain" //nolint:depguard
	"github.com/alexeyco/simpletable"           //nolint:depguard
	"github.com/logrusorgru/aurora"             //nolint:depguard
)

// PrintMigrations - выводит таблицу всех перенесенных миграций.
func PrintMigrations(migrations []domain.Migration) {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Span: 0, Text: "#"},
			{Align: simpletable.AlignCenter, Span: 0, Text: "Version"},
			{Align: simpletable.AlignCenter, Span: 0, Text: "Name"},
			{Align: simpletable.AlignCenter, Span: 0, Text: "Is applied?"},
			{Align: simpletable.AlignCenter, Span: 0, Text: "Date update"},
		},
	}

	for index, migration := range migrations {
		isApplied := aurora.Blue("No").String()
		if migration.IsApplied {
			isApplied = aurora.Cyan("Yes").String()
		}

		row := []*simpletable.Cell{
			{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", index+1)},
			{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%d", migration.Version)},
			{Align: simpletable.AlignCenter, Text: migration.Name},
			{Align: simpletable.AlignCenter, Text: isApplied},
			{Align: simpletable.AlignCenter, Text: migration.UpdateAt.String()},
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}

	table.SetStyle(simpletable.StyleDefault)
	table.Println()
}

// PrintMigration - выводит информацию о миграции.
func PrintMigration(migration domain.Migration) {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Span: 0, Text: "Version"},
			{Align: simpletable.AlignCenter, Span: 0, Text: "Name"},
			{Align: simpletable.AlignCenter, Span: 0, Text: "Date update"},
		},
	}

	var (
		version aurora.Value
		name    aurora.Value
		date    aurora.Value
	)

	if migration.IsApplied {
		version = aurora.Cyan(migration.Version)
		name = aurora.Cyan(migration.Name)
		date = aurora.Cyan(migration.UpdateAt.String())
	} else {
		version = aurora.Blue(migration.Version)
		name = aurora.Blue(migration.Name)
		date = aurora.Blue(migration.UpdateAt.String())
	}

	row := []*simpletable.Cell{
		{Align: simpletable.AlignCenter, Text: version.String()},
		{Align: simpletable.AlignCenter, Text: name.String()},
		{Align: simpletable.AlignCenter, Text: date.String()},
	}
	table.Body.Cells = append(table.Body.Cells, row)

	table.SetStyle(simpletable.StyleCompactLite)
	table.Println()
}
