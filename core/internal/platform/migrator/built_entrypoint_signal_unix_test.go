//go:build integration && !windows

package migrator

import (
	"os/exec"
	"syscall"
)

func configureBuiltEntrypointProcess(*exec.Cmd) {}

func signalBuiltEntrypoint(command *exec.Cmd) error {
	return command.Process.Signal(syscall.SIGTERM)
}
