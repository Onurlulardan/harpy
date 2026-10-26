package postgres

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func validClaimCatalog(t *testing.T) jobsv1.Catalog {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(
		"forms.export", 1, "forms", jobsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestAssertJobCatalogRejectsUnsupportedLiveRows(t *testing.T) {
	for _, test := range []struct {
		name       string
		compatible bool
		want       error
	}{
		{name: "exact", compatible: true},
		{name: "unsupported", want: ErrJobCatalogIncompatible},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := newWorkerTestRepository()
			repository.database = &workerTestDatabase{queryRows: &workerTestRows{
				values: [][]any{{test.compatible}},
			}}
			err := repository.AssertJobCatalog(context.Background(), validClaimCatalog(t))
			if !errors.Is(err, test.want) {
				t.Fatalf("catalog assertion error = %v", err)
			}
		})
	}
}

func TestAssertJobCatalogFailsClosedOnMissingResult(t *testing.T) {
	repository := newWorkerTestRepository()
	repository.database = &workerTestDatabase{queryRows: &workerTestRows{}}
	err := repository.AssertJobCatalog(context.Background(), validClaimCatalog(t))
	if !errors.Is(err, ErrJobCatalogIncompatible) {
		t.Fatalf("missing catalog result error = %v", err)
	}
}

func TestClaimPoliciesPreserveAllExecutionFences(t *testing.T) {
	values, err := validatedClaimPolicies(validClaimCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(values.kinds) != 1 || values.kinds[0] != "forms.export" ||
		values.revisions[0] != 1 || values.domains[0] != "forms" ||
		values.effects[0] != string(jobsv1.EffectDatabaseTransactional) ||
		values.maximumAttempts[0] != 8 {
		t.Fatalf("claim policy arrays = %#v", values)
	}
}
