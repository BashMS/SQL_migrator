package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/spf13/viper"
)

const (
	defaultConfigPath = "config/config.yml"

	// LogLevelDebug - уровень логгирования.
	LogLevelDebug = "debug"
	LogLevelInfo = "info"
	LogLevelWarn = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"

	// FormatSQL - sql формат.
	FormatSQL = "sql"
	// FormatGolang - go формат.
	FormatGolang = "golang"

	// ExtSQL - расширение для sql.
	ExtSQL = ".sql"
	// ExtGolang - расширение для go.
	ExtGolang = ".go"

	// PostfixUp up постфикс для наката.
	PostfixUp = ".up"
	// PostfixDown down постфикс для отката.
	PostfixDown = ".down"

	// Separator - разделитель.
	Separator = '_'
)

// ErrConfigurationFileNotFound - файл конфигурации не найден.
var ErrConfigurationFileNotFound = errors.New("configuration file not found")

// Config.
type Config struct {
	DSN            string
	Path           string
	Format         string
	LogPath        string
	LogLevel       string
	viperConfig    *viper.Viper
}

// ReadConfigFromFile - читает файл конфигурации.
func (c *Config) ReadConfigFromFile(path string) error {
	if !fileutil.Exist(path) {
		return ErrConfigurationFileNotFound
	}

	c.viper().SetConfigFile(path)
	c.viper().SetConfigType("yml")
	if err := c.viper().ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// ReadConfigWithEnv - читает файл конфигурации с использованием переменных среды.
func (c *Config) ReadConfigWithEnv() bool {
	configPath, ok := os.LookupEnv("MIGRATOR_CONFIG_PATH")
	if ok {
		if err := c.ReadConfigFromFile(configPath); err == nil && c.viper().IsSet("migrator") {
			return true
		}
	}

	return false
}

// ReadConfigFromDefaultPath - считывает файл конфигурации по пути по умолчанию `config/config.yml`
func (c *Config) ReadConfigFromDefaultPath() bool {
	dir, err := os.Getwd()
	if err == nil {
		if err := c.ReadConfigFromFile(filepath.Join(dir, defaultConfigPath)); err == nil && c.viper().IsSet("migrator") {
			return true
		}
	}

	return false
}

// Apply - применяет конфигурацию чтения к текущему объекту.
func (c *Config) Apply() {
	if c.DSN == "" {
		c.DSN = os.ExpandEnv(c.viper().GetString("migrator.dsn"))
	}
	if c.Path == "" {
		c.Path = os.ExpandEnv(c.viper().GetString("migrator.path"))
	}
	if c.Format == "" {
		c.Format = os.ExpandEnv(c.viper().GetString("migrator.format"))
	}
	if c.LogPath == "" {
		c.LogPath = os.ExpandEnv(c.viper().GetString("migrator.log.path"))
	}
	if c.LogLevel == "" {
		c.LogLevel = os.ExpandEnv(c.viper().GetString("migrator.log.level"))
	}
}

// PathConversion - заменяет относительные пути на абсолютные.
func (c *Config) PathConversion() error {
	var err error
	if c.Path != "" {
		c.Path, err = filepath.Abs(c.Path)
		if err != nil {
			return err
		}
	}
	if c.LogPath != "" {
		c.LogPath, err = filepath.Abs(c.LogPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) viper() *viper.Viper {
	if c.viperConfig == nil {
		c.viperConfig = viper.New()
		replacer := strings.NewReplacer(".", "_")
		c.viper().SetEnvKeyReplacer(replacer)
		c.viperConfig.AutomaticEnv()
		c.applyDefault()
	}

	return c.viperConfig
}

func (c *Config) applyDefault() {
	c.viper().SetDefault("migrator.format", FormatSQL)
}
