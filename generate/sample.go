package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/BashMS/SQL_migrator/internal/template"
)

const (
	fileSample  = "sample.go.tpl"
	extTemplate = ".tpl"
)

var (
	errNotFoundArguments = errors.New("not all arguments passed")
)

type _data struct {
	SampleName string
	Content    string
	FileName   string
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal(errNotFoundArguments)
	}
	templatePath := os.Args[1]
	outputDir := os.Args[2]

	if err := handle(templatePath, outputDir); err != nil {
		log.Fatal(err)
	}
}

func handle(templatePath, outputDir string) error {
	samplePath := filepath.Join(templatePath, fileSample)
	sampleContent, err := ioutil.ReadFile(samplePath)
	if err != nil {
		return err
	}
	return filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if path == samplePath {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file contents (%s): %w", path, err)
		}

		base := strings.TrimSuffix(filepath.Base(path), extTemplate)
		ext := filepath.Ext(base)

		sample := template.Sample{}
		sample.Name = base
		sample.Text = string(sampleContent)
		sample.Data = _data{
			SampleName: strcase.ToCamel(strings.TrimSuffix(base, ext)),
			Content:    string(content),
			FileName:   base,
		}

		outputPath := filepath.Join(outputDir, base)
		outputPath, err = filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("invalid path for output file: %w", err)
		}
		err = template.Create(outputPath, sample)
		if err != nil {
			return fmt.Errorf("failed to apply template engine: %w", err)
		}

		fmt.Printf("Created or replace file %s\n", outputPath)
		cmd := exec.Command("go", "fmt", outputPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("`go fmt -w` failed: %w", err)
		}

		return nil
	})
}
