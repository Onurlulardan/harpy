package postgres

import (
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestExternalEffectPermitSealsRequestAndWorkerAuthority(t *testing.T) {
	t.Parallel()
	repository, claim := validLifecycleClaim(t, jobsv1.EffectExternal)
	spec := validExternalSpec(t)
	if spec.Provider() != "forms-exporter" || spec.Operation() != "export" {
		t.Fatalf("spec getters = %q %q", spec.Provider(), spec.Operation())
	}
	startedAt := time.Now().UTC()
	request, err := buildExternalEffectRequest(claim, spec, startedAt)
	if err != nil {
		t.Fatal(err)
	}
	permit, err := newExternalEffectPermit(
		repository.identity, claim, request, startedAt,
	)
	if err != nil || !permit.valid(repository.identity) {
		t.Fatalf("valid permit = %#v, %v", permit, err)
	}
	copy := permit.Request()
	if copy.TenantID == nil {
		t.Fatal("tenant effect request lost tenant identity")
	}
	*copy.TenantID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	second := permit.Request()
	if second.TenantID == nil || *second.TenantID != claim.tenantID.String() {
		t.Fatal("permit request exposed mutable tenant identity")
	}
	if permit.valid(&workerIdentity{marker: 1}) {
		t.Fatal("permit accepted a different worker identity")
	}
}

func TestExternalEffectRequestRejectsWrongEffectBoundary(t *testing.T) {
	t.Parallel()
	_, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	if _, err := buildExternalEffectRequest(
		claim, validExternalSpec(t), time.Now().UTC(),
	); err == nil {
		t.Fatal("database-transactional claim produced an external request")
	}
}
