package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type scriptRunner struct {
	stdout []byte
	show   error
	run    error
	calls  [][]string
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
