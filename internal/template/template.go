package template

//go:generate go run ./../../generate/sample.go ./../../templates/ ./

import (
	"os"
	"text/template"

	"github.com/BashMS/SQL_migrator/internal/loader" //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/config"      //nolint:depguard
	"github.com/BashMS/SQL_migrator/pkg/domain"      //nolint:depguard
)

type (
	// Sample struct.
	Sample struct {
		Name string
		Text string
		Data interface{}
	}

	dataMain struct {
		Config     *config.Config
		Migrations []loader.RawMigration
		Direction  bool
	}

	dataGolangMigrationMethod struct {
		Version uint64
		Name    string
	}
)

func Create(path string, sample Sample) error {
	var err error
	tpl := template.New(sample.Name)
	tpl, err = tpl.Parse(sample.Text)
	if err != nil {
		return err
	}

	return writeSample(path, tpl, sample.Data)
}

func CreateMainSample(path string, config *config.Config, migrations []loader.RawMigration, direction bool) error {
	sample := SampleGolangMainFile
	sample.Data = dataMain{
		Config:     config,
		Migrations: migrations,
		Direction:  direction,
	}

	return Create(path, sample)
}

func CreateGolangMigrationMethod(path string, migration domain.Migration) error {
	sample := SampleGolangMigrationMethod
	sample.Data = dataGolangMigrationMethod{
		Version: migration.Version,
		Name:    migration.Name,
	}

	return Create(path, sample)
}

func writeSample(path string, tpl *template.Template, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tpl.Execute(file, data); err != nil {
		return err
	}

	return nil
}
