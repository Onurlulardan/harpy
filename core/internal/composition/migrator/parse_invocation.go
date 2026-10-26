package migrator

import "fmt"

type invocationMode uint8

const (
	invocationDeploy invocationMode = iota + 1
	invocationSchemaExecutor
	invocationTenantMigrationExecutor
	invocationCutoverAdoption
	invocationCutoverFinalization
	invocationDatabaseBackupExecutor
	invocationDatabaseRestore
)

type deployAction string

const (
	deployPlan   deployAction = "plan"
	deployVerify deployAction = "verify"
	deployUp     deployAction = "up"
)

type invocation struct {
	mode   invocationMode
	action deployAction
}

func parseInvocation(arguments []string) (invocation, error) {
	if len(arguments) == 1 {
		switch arguments[0] {
		case "schema-executor":
			return invocation{mode: invocationSchemaExecutor}, nil
		case "tenant-migration-executor":
			return invocation{mode: invocationTenantMigrationExecutor}, nil
		case "cutover_adoption":
			return invocation{mode: invocationCutoverAdoption}, nil
		case "cutover_finalization":
			return invocation{mode: invocationCutoverFinalization}, nil
		case "database-backup-executor":
			return invocation{mode: invocationDatabaseBackupExecutor}, nil
		case "database-restore":
			return invocation{mode: invocationDatabaseRestore}, nil
		}
	}
	if len(arguments) != 2 || arguments[0] != "deploy" {
		return invocation{}, invalidInvocation()
	}
	action := deployAction(arguments[1])
	switch action {
	case deployPlan, deployVerify, deployUp:
		return invocation{mode: invocationDeploy, action: action}, nil
	default:
		return invocation{}, invalidInvocation()
	}
}

func invalidInvocation() error {
	return fmt.Errorf(
		"usage: core-migrator deploy plan|verify|up, schema-executor, " +
			"tenant-migration-executor, cutover_adoption, cutover_finalization, " +
			"database-backup-executor, or database-restore",
	)
}
