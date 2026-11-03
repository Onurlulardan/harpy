package migrator

import "testing"

func TestParseInvocationAcceptsRecoveryAndCutoverCommands(t *testing.T) {
	tests := []struct {
		arguments []string
		mode      invocationMode
	}{
		{arguments: []string{"cutover_adoption"}, mode: invocationCutoverAdoption},
		{arguments: []string{"cutover_finalization"}, mode: invocationCutoverFinalization},
		{arguments: []string{"database-backup-executor"}, mode: invocationDatabaseBackupExecutor},
		{arguments: []string{"database-restore"}, mode: invocationDatabaseRestore},
	}
	for _, test := range tests {
		actual, err := parseInvocation(test.arguments)
		if err != nil || actual.mode != test.mode {
			t.Fatalf("parse %v = %#v %v", test.arguments, actual, err)
		}
	}
}

func TestParseInvocationRejectsMigrationCommands(t *testing.T) {
	tests := [][]string{
		nil,
		{},
		{"deploy"},
		{"deploy", "up"},
		{"schema-executor"},
		{"tenant-migration-executor"},
		{"database-backup-executor", "--once"},
		{"database-restore", "--force"},
		{"help"},
	}
	for _, arguments := range tests {
		if _, err := parseInvocation(arguments); err == nil {
			t.Fatalf("accepted uncontracted invocation %v", arguments)
		}
	}
}
