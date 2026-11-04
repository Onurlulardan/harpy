package recovery

import "testing"

func TestParseInvocationAcceptsRecoveryCommands(t *testing.T) {
	tests := []struct {
		arguments []string
		mode      invocationMode
	}{
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
		{"cutover_adoption"},
		{"cutover_finalization"},
		// PRESERVE_NON_MIGRATION: recovery must reject removed executor commands.
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
