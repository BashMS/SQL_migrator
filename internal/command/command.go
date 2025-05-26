package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type (
	Command interface {
		Run(ctx context.Context, name string, args Args, dir string, env Env) error
		RunWithGracefulShutdown(ctx context.Context, name string, args Args, dir string, env Env) error
	}
	command struct{}

	// Args - аргументы командной строки.
	Args []string
	// Env - переменные окружения.
	Env []string
)

// NewCommand консруктор.
func NewCommand() Command {
	return &command{}
}

// Run запуск с переменными.
func (c *command) Run(ctx context.Context, name string, args Args, dir string, env Env) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if err := c.apply(cmd, dir, env).Run(); err != nil {
		return fmt.Errorf("`%s %v` failed: %w", name, args, err)
	}

	return nil
}

// RunWithGracefulShutdown - выполнить в определенном каталоге и среде (с корректным завершением работы).
func (c *command) RunWithGracefulShutdown(ctx context.Context, name string, args Args, dir string, env Env) error {
	cmd := exec.Command(name, args...)
	if err := c.apply(cmd, dir, env).Start(); err != nil {
		return fmt.Errorf("`%s %v` failed: %w", name, args, err)
	}

	chWait := make(chan struct{})
	chErr := make(chan error, 1)

	go func() {
		defer close(chErr)
		defer close(chWait)
		if err := cmd.Wait(); err != nil {
			chErr <- fmt.Errorf("failed to complete the `%s %v : %w`", name, args, err)
			return
		}
	}()

	select {
	case <-ctx.Done():
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			if err := cmd.Process.Kill(); err != nil {
				return err
			}
		}
		_, err := cmd.Process.Wait()
		if err != nil {
			return fmt.Errorf("process `%s %v` is not completed correctly: %w", name, args, err)
		}
	case err := <-chErr:
		return err
	case <-chWait:
	}

	return nil
}

func (c *command) apply(cmd *exec.Cmd, dir string, env []string) *exec.Cmd {
	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
