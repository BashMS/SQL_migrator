package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/logrusorgru/aurora"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/BashMS/SQL_migrator/pkg/config"
)

//ConsoleLogger - имя логгера в консоли.
const ConsoleLogger = "migrate"

//ErrMakeDir - не удалось создать каталог для логгов.
var ErrMakeDir = errors.New("failed to create directory for logs")

//New конструктор логгера.
func New(config *config.Config) (*zap.Logger, error) {
	var (
		encodeConfig zapcore.EncoderConfig
		err          error
	)

	var outputPaths []string
	var errOutputPaths []string
	if len(config.LogPath) > 0 {
		if !fileutil.Exist(config.LogPath) {
			if err := makeLogPath(config.LogPath); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrMakeDir, err)
			}
		}
		outputPaths = append(outputPaths, config.LogPath)
		errOutputPaths = append(errOutputPaths, config.LogPath)
	}

	level := zap.NewAtomicLevel()
	if len(config.LogLevel) > 0 {
		if err = level.UnmarshalText([]byte(config.LogLevel)); err != nil {
			return nil, err
		}
	}

	encodeConfig = zap.NewProductionEncoderConfig()

	logConfig := zap.Config{
		Level:             level,
		Development:       true,
		Encoding:          "json",
		EncoderConfig:     encodeConfig,
		OutputPaths:       outputPaths,
		ErrorOutputPaths:  errOutputPaths,
		DisableCaller:     false,
		DisableStacktrace: false,
	}

	return logConfig.Build(zap.Hooks(consoleHook))
}

func consoleHook(entry zapcore.Entry) error {
	if entry.LoggerName == ConsoleLogger {
		switch entry.Level {
		case zapcore.DebugLevel:
			fmt.Println(entry.Message)
		case zapcore.InfoLevel:
			fmt.Println(aurora.Cyan(entry.Message))
		case zapcore.WarnLevel:
			fmt.Println(aurora.Yellow(entry.Message))
		case zapcore.ErrorLevel:
			fallthrough
		case zapcore.DPanicLevel:
			fallthrough
		case zapcore.PanicLevel:
			fallthrough
		case zapcore.FatalLevel:
			fmt.Println(aurora.Red(entry.Message))
		default:
			fmt.Println(entry.Message)
		}
	}

	return nil
}

//Flush - записывает логи на диск.
func Flush(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		pathError, ok := err.(*os.PathError)
		if ok && (pathError.Path == "/dev/stdout" || pathError.Path == "/dev/stderr") {
			return
		}
		log.Fatalf("error flush logs to disk: %s", err)
	}
}

func makeLogPath(logPath string) error {
	directory := filepath.Dir(logPath)
	if !fileutil.Exist(directory) {
		err := fileutil.TouchDirAll(directory)
		if err != nil {
			return err
		}
	}
	return nil
}
