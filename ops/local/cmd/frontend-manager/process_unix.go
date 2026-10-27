//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func configureManagedProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func processIdentity(pid int) (string, bool) {
	if err := syscall.Kill(pid, 0); err != nil {
		return "", false
	}
	if runtime.GOOS == "linux" {
		content, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			return "", false
		}
		closing := strings.LastIndexByte(string(content), ')')
		if closing < 0 {
			return "", false
		}
		fields := strings.Fields(string(content)[closing+1:])
		if len(fields) < 20 {
			return "", false
		}
		return fields[19], true
	}
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=").Output()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return "", false
	}
	return strings.TrimSpace(string(output)), true
}

func terminateProcessTree(pid int) error {
	if _, alive := processIdentity(pid); !alive {
		return nil
	}
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal frontend process group: %w", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, alive := processIdentity(pid); !alive {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("kill frontend process group: %w", err)
	}
	return nil
}
