package pgtools

import (
	"context"
	"errors"
	"os/exec"
)

type OSRunner struct{}

func (OSRunner) Run(ctx context.Context, command Command) error {
	if ctx == nil || ctx.Err() != nil || !validExecutable(command.Executable) {
		return errors.New("PostgreSQL client command is invalid")
	}
	process := exec.CommandContext(ctx, command.Executable, command.Arguments...)
	process.Env = append([]string(nil), command.Environment...)
	process.Stdin, process.Stdout, process.Stderr = command.Stdin, command.Stdout, command.Stderr
	return process.Run()
}
