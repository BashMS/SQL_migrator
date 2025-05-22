package util

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strings"
)

var (
	// ErrCopyFile - не удалось скопировать файл.
	ErrCopyFile = errors.New("could not copy file")
	// ErrCreateFile - не удалось создать файл.
	ErrCreateFile = errors.New("could not create file")
)

// CopyFile - копирует файл.
func CopyFile(dest, src string) error {
	sourceFile, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFile.Mode().IsRegular() {
		return fmt.Errorf("%w: not a regular file (%s)", ErrCopyFile, src)
	}
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("%w: %s (%s)", ErrCopyFile, err, src)
	}
	defer source.Close()
	destination, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("%w: %s (%s)", ErrCopyFile, err, src)
	}
	defer destination.Close()
	ioBytes, err := io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("%w: %s (%s)", ErrCopyFile, err, src)
	}
	if ioBytes == 0 {
		return fmt.Errorf("%w: copied 0 bytes (%s)", ErrCopyFile, src)
	}

	return nil
}

// CreateFile - создает файл.
func CreateFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("%w: %s (%s)", ErrCreateFile, err, path)
	}
	defer f.Close()

	return nil
}

// GenerateUID - генерирует уникального ключа по имени.
func GenerateUID(name string, keys ...string) uint32 {
	if len(keys) > 0 {
		name = strings.Join(append(keys, name), ":")
	}

	return crc32.ChecksumIEEE([]byte(name))
}
