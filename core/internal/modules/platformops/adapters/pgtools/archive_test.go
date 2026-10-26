package pgtools

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"
)

type captureRunner struct{ command Command }

func (runner *captureRunner) Run(_ context.Context, command Command) error {
	runner.command = command
	if command.Stdout != nil {
		_, _ = command.Stdout.Write([]byte("archive"))
	}
	return nil
}

func TestDumpAndRestoreNeverPlacePasswordInArguments(t *testing.T) {
	runner := &captureRunner{}
	tools := New(runner, "/usr/bin/pg_dump", "/usr/bin/pg_restore")
	dsn := "postgresql://harpy_backup:very-secret@db:5432/harpy?sslmode=verify-full"
	var archive bytes.Buffer
	if err := tools.Dump(context.Background(), dsn, "00000003-0000001B-1", &archive); err != nil {
		t.Fatal(err)
	}
	assertSecretBoundary(t, runner.command)
	if err := tools.Restore(context.Background(), dsn, &archive); err != nil {
		t.Fatal(err)
	}
	assertSecretBoundary(t, runner.command)
	if !slices.Contains(runner.command.Arguments, "--role=harpy_owner") {
		t.Fatalf("restore owner role is absent: %#v", runner.command.Arguments)
	}
	if slices.Contains(runner.command.Arguments, "--no-privileges") {
		t.Fatalf("restore discarded database ACLs: %#v", runner.command.Arguments)
	}
}

func TestDSNRejectsArbitrarySessionOptions(t *testing.T) {
	dsn := "postgresql://user:secret@db/harpy?options=-c%20search_path%3Dpublic"
	if _, _, err := sanitizedDSN(dsn); err == nil {
		t.Fatal("arbitrary DSN options accepted")
	}
}

func assertSecretBoundary(t *testing.T, command Command) {
	t.Helper()
	if strings.Contains(strings.Join(command.Arguments, " "), "very-secret") ||
		len(command.Environment) != 1 || command.Environment[0] != "PGPASSWORD=very-secret" {
		t.Fatalf("secret boundary violated: %#v %#v", command.Arguments, command.Environment)
	}
}
