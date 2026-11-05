package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type scriptRunner struct {
	stdout   []byte
	show     error
	run      error
	combined []byte
	code     int
	calls    [][]string
}

func (runner *scriptRunner) Output(_ context.Context, argv []string) ([]byte, error) {
	runner.calls = append(runner.calls, append([]string(nil), argv...))
	if runner.show != nil {
		return nil, runner.show
	}
	return runner.stdout, nil
}

func (runner *scriptRunner) Run(_ context.Context, argv []string) error {
	runner.calls = append(runner.calls, append([]string(nil), argv...))
	return runner.run
}

func (runner *scriptRunner) Combined(_ context.Context, argv []string) ([]byte, int, error) {
	runner.calls = append(runner.calls, append([]string(nil), argv...))
	if runner.code == 0 {
		return runner.combined, 0, nil
	}
	return runner.combined, runner.code, errors.New("exit")
}

func TestProvisionTenantUsesTheUUIDFromShow(t *testing.T) {
	const tenantID = "00112233-4455-6677-8899-aabbccddeeff"
	runner := &scriptRunner{stdout: []byte(tenantID + "\n")}
	if err := provisionTenant(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("calls=%v", runner.calls)
	}
	show := strings.Join(runner.calls[0], " ")
	provision := strings.Join(runner.calls[1], " ")
	if !strings.Contains(show, "run -T --rm --no-deps core-operator show-provision-request") ||
		!strings.Contains(show, "-f ops/local/compose.runtime.yml") {
		t.Fatal(show)
	}
	if !strings.HasSuffix(provision, "tenant-provisioning") ||
		!strings.Contains(provision, "--rm --no-deps -e HARPY_PROVISION_TENANT_ID="+tenantID) ||
		strings.Contains(provision, "docker.sock") {
		t.Fatal(provision)
	}
}

func TestProvisionTenantDoesNotStartComposeWhenTheIDIsInvalid(t *testing.T) {
	runner := &scriptRunner{stdout: []byte("not-a-uuid\n")}
	err := provisionTenant(context.Background(), runner)
	if !errors.Is(err, errInvalidTenant) || len(runner.calls) != 1 {
		t.Fatalf("err=%v calls=%d", err, len(runner.calls))
	}
}

func TestProvisionTenantDoesNotStartComposeWhenShowFails(t *testing.T) {
	runner := &scriptRunner{show: errors.New("exit status 2")}
	err := provisionTenant(context.Background(), runner)
	if err == nil || errors.Is(err, errInvalidTenant) || len(runner.calls) != 1 {
		t.Fatalf("err=%v calls=%d", err, len(runner.calls))
	}
}

func TestExecuteRejectsArguments(t *testing.T) {
	if err := execute(context.Background(), []string{"tenant"}); err == nil {
		t.Fatal("expected usage error")
	}
}

func TestWaitOperatorProvisionsOnlyForHostWait(t *testing.T) {
	const tenantID = "00112233-4455-6677-8899-aabbccddeeff"
	runner := &scriptRunner{
		combined: []byte("tenant provisioning is waiting for host execution\n"),
		code:     5,
		stdout:   []byte(tenantID + "\n"),
	}
	if err := waitOperator(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("calls=%d", len(runner.calls))
	}
	if !strings.Contains(strings.Join(runner.calls[2], " "), "HARPY_PROVISION_TENANT_ID="+tenantID) {
		t.Fatal(runner.calls[2])
	}
}

func TestWaitOperatorStopsOnAnotherExitFive(t *testing.T) {
	runner := &scriptRunner{
		combined: []byte("tenant schema name is invalid: harpy\n"),
		code:     5,
	}
	err := waitOperator(context.Background(), runner)
	if err == nil || len(runner.calls) != 1 {
		t.Fatalf("err=%v calls=%d", err, len(runner.calls))
	}
}
