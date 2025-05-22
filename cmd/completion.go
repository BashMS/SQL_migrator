package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/BashMS/SQL_migrator/pkg/logger"
)

const (
	shellBash = "bash"
	shellZsh  = "zsh"
)

//completionCmd команда завершения.
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates completion scripts",
	Long: `Для загрузки завершения запустите 'migrator completion', команда может автоматически определить текущую оболочку
Команда создаст файл '.bash_completion_migrator' или '.zsh_completion_migrator' в домашнем каталоге пользователя
Чтобы настроить оболочку bash/zsh для загрузки завершений для каждого сеанса, добавьте в свой bashrc/zshrc: 

# ~/.bashrc or ~/.profile
if [-f ~/.bash_completion_migrator ]; then
	. ~/.bash_completion_migrator
fi

OR

# ~/.zshrc or ~/.profile
if [-f ~/.zsh_completion_migrator ]; then
	. ~/.zsh_completion_migrator
fi
`,
	SilenceUsage: true,
	Example:      "migrator completion [bash|zsh]",
	Run: func(cmd *cobra.Command, args []string) {
		zLogger, err := logger.New(&cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer logger.Flush(zLogger)

		home, err := homedir.Dir()
		if err != nil {
			zLogger.Error(fmt.Sprintf("could not determine home directory: %s", err))
			os.Exit(1)
		}

		var shell string
		if len(args) > 0 {
			shell = args[0]
		}

		if shell == "" {
			envShell := os.Getenv("SHELL")
			if strings.HasSuffix(envShell, shellBash) {
				shell = shellBash
			} else if strings.HasSuffix(envShell, shellZsh) {
				shell = shellZsh
			}

			if shell != "" {
				zLogger.Info("command shell was detected automatically")
			}
		}

		var filePath string
		switch shell {
		case shellBash:
			filePath = path.Join(home, ".bash_completion_migrator")
			if err := rootCmd.GenBashCompletionFile(filePath); err != nil {
				zLogger.Error(fmt.Sprintf("could not create file on path %s : %s", filePath, err))
				os.Exit(1)
			}
		case shellZsh:
			filePath = path.Join(home, ".zsh_completion_migrator")
			if err := rootCmd.GenZshCompletionFile(filePath); err != nil {
				zLogger.Error(fmt.Sprintf("could not create file on path %s : %s", filePath, err))
				os.Exit(1)
			}
		default:
			zLogger.Error("could not determine shell, use bash or zsh arguments")
		}

		if filePath != "" {
			zLogger.Info(fmt.Sprintf("%s file to completion was created successfully", filePath))
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
