package migrator

import "testing"

func TestParseInvocationAcceptsOnlyContractedCommands(t *testing.T) {
	tests := []struct {
		arguments []string
		mode      invocationMode
		action    deployAction
	}{
		{arguments: nil, mode: invocationDeploy, action: deployUp},
		{arguments: []string{}, mode: invocationDeploy, action: deployUp},
		{arguments: []string{"deploy", "plan"}, mode: invocationDeploy, action: deployPlan},
		{arguments: []string{"deploy", "verify"}, mode: invocationDeploy, action: deployVerify},
		{arguments: []string{"deploy", "up"}, mode: invocationDeploy, action: deployUp},
		{arguments: []string{"schema-executor"}, mode: invocationSchemaExecutor},
		{arguments: []string{"tenant-migration-executor"}, mode: invocationTenantMigrationExecutor},
		{arguments: []string{"cutover_adoption"}, mode: invocationCutoverAdoption},
		{arguments: []string{"cutover_finalization"}, mode: invocationCutoverFinalization},
		{arguments: []string{"database-backup-executor"}, mode: invocationDatabaseBackupExecutor},
		{arguments: []string{"database-restore"}, mode: invocationDatabaseRestore},
	}
	for _, test := range tests {
		actual, err := parseInvocation(test.arguments)
		if err != nil {
			t.Fatalf("parse %v: %v", test.arguments, err)
		}
		if actual.mode != test.mode || actual.action != test.action {
			t.Fatalf("parse %v = %#v", test.arguments, actual)
		}
	}
}

func TestParseInvocationRejectsEveryUncontractedShape(t *testing.T) {
	tests := [][]string{
		{"deploy"},
		{"deploy", "down"},
		{"deploy", "up", "--out-of-order"},
		{"schema-executor", "--once"},
		{"cutover_adoption", "--batch"},
		{"cutover_finalization", "--force"},
		{"database-backup-executor", "--once"},
		{"database-restore", "--force"},
		{"extension-admin"},
		{"help"},
	}
	for _, arguments := range tests {
		if _, err := parseInvocation(arguments); err == nil {
			t.Fatalf("accepted uncontracted invocation %v", arguments)
		}
	}
}
