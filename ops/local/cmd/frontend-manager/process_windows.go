//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

func configureManagedProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.CREATE_NO_WINDOW,
		HideWindow:    true,
	}
}

func processIdentity(pid int) (string, bool) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return "", false
	}
	defer windows.CloseHandle(handle)
	var creation, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
		return "", false
	}
	return strconv.FormatInt(creation.Nanoseconds(), 10), true
}

func terminateProcessTree(pid int) error {
	if _, alive := processIdentity(pid); !alive {
		return nil
	}
	command := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.CombinedOutput()
	if err != nil {
		if _, alive := processIdentity(pid); !alive {
			return nil
		}
		return fmt.Errorf("stop frontend process tree: %w: %s", err, output)
	}
	return nil
}
