package recovery

import (
	"fmt"
	"strings"
)

type invocationMode uint8

const (
	invocationDatabaseBackupExecutor invocationMode = iota + 1
	invocationDatabaseRestore
)

type invocation struct {
	mode invocationMode
}

func parseInvocation(arguments []string) (invocation, error) {
	if len(arguments) != 1 {
		return invocation{}, invalidInvocation()
	}
	switch arguments[0] {
	case "database-backup-executor":
		return invocation{mode: invocationDatabaseBackupExecutor}, nil
	case "database-restore":
		return invocation{mode: invocationDatabaseRestore}, nil
	default:
		return invocation{}, invalidInvocation()
	}
}

func commandArguments(arguments []string) []string {
	remaining := make([]string, 0, len(arguments))
	for index := 0; index < len(arguments); index++ {
		arg := arguments[index]
		switch {
		case arg == "--env-file":
			index++
		case strings.HasPrefix(arg, "--env-file="):
		default:
			remaining = append(remaining, arg)
		}
	}
	return remaining
}

func invalidInvocation() error {
	return fmt.Errorf(
		"usage: core-recovery database-backup-executor or database-restore",
	)
}
