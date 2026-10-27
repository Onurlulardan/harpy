package postgres

import (
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func validLifecycleClaim(
	t *testing.T,
	effectClass jobsv1.EffectClass,
) (*workerRepository, JobClaim) {
	t.Helper()
	repository := newWorkerTestRepository()
	authority := validWorkerAuthority(t, repository)
	workerID := uuid.New()
	row := validJobClaimRow(authority, workerID, time.Now().UTC())
	row.effectClass = string(effectClass)
	if effectClass == jobsv1.EffectExternal {
		row.effectKey = pgtype.Text{String: "effect:forms:export:42", Valid: true}
	}
	claim, err := newJobClaim(row, authority, workerID)
	if err != nil {
		t.Fatal(err)
	}
	return repository, claim
}

func validExternalSpec(t *testing.T) ExternalEffectRequestSpec {
	t.Helper()
	spec, err := NewExternalEffectRequestSpec(
		uuid.New(), "forms-exporter", "export",
		"request:forms.export.42", uuid.New(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func validGlobalClaim(t *testing.T, repository *workerRepository) JobClaim {
	t.Helper()
	authority, err := newJobAuthority(
		repository.identity, "global", uuid.New(), "operations", 2, "workflow", 4,
	)
	if err != nil {
		t.Fatal(err)
	}
	workerID := uuid.New()
	row := validJobClaimRow(authority, workerID, time.Now().UTC())
	row.tenantID = pgtype.UUID{}
	claim, err := newJobClaim(row, authority, workerID)
	if err != nil {
		t.Fatal(err)
	}
	return claim
}
