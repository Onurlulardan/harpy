//go:build integration && windows

package migrator

import (
	"fmt"
	"os/exec"
	"syscall"
)

const createNewProcessGroup = 0x00000200

var generateConsoleControlEvent = syscall.NewLazyDLL("kernel32.dll").
	NewProc("GenerateConsoleCtrlEvent")

func configureBuiltEntrypointProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
}

func signalBuiltEntrypoint(command *exec.Cmd) error {
	result, _, callErr := generateConsoleControlEvent.Call(
		uintptr(1), uintptr(command.Process.Pid),
	)
	if result == 0 {
		return fmt.Errorf("GenerateConsoleCtrlEvent: %w", callErr)
	}
	return nil
}
