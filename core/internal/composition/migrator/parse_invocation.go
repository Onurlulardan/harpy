package migrator

import "fmt"

type invocationMode uint8

const (
	invocationCutoverAdoption invocationMode = iota + 1
	invocationCutoverFinalization
	invocationDatabaseBackupExecutor
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
	case "cutover_adoption":
		return invocation{mode: invocationCutoverAdoption}, nil
	case "cutover_finalization":
		return invocation{mode: invocationCutoverFinalization}, nil
	case "database-backup-executor":
		return invocation{mode: invocationDatabaseBackupExecutor}, nil
	case "database-restore":
		return invocation{mode: invocationDatabaseRestore}, nil
	default:
		return invocation{}, invalidInvocation()
	}
}

func invalidInvocation() error {
	return fmt.Errorf(
		"usage: core-migrator cutover_adoption, cutover_finalization, " +
			"database-backup-executor, or database-restore",
	)
}
