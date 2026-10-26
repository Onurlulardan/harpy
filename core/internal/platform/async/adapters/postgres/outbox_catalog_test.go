package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestAssertOutboxCatalogRejectsUnsupportedLiveRows(t *testing.T) {
	for _, test := range []struct {
		name       string
		compatible bool
		want       error
	}{
		{name: "exact", compatible: true},
		{name: "unsupported", want: ErrOutboxCatalogIncompatible},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := newWorkerTestRepository()
			repository.database = &workerTestDatabase{queryRows: &workerTestRows{
				values: [][]any{{test.compatible}},
			}}
			err := repository.AssertOutboxCatalog(context.Background(), validPublisherCatalog(t))
			if !errors.Is(err, test.want) {
				t.Fatalf("catalog assertion error = %v", err)
			}
		})
	}
}

func TestAssertOutboxCatalogFailsClosedOnMissingResult(t *testing.T) {
	repository := newWorkerTestRepository()
	repository.database = &workerTestDatabase{queryRows: &workerTestRows{}}
	err := repository.AssertOutboxCatalog(context.Background(), validPublisherCatalog(t))
	if !errors.Is(err, ErrOutboxCatalogIncompatible) {
		t.Fatalf("missing catalog result error = %v", err)
	}
}

func TestOutboxPoliciesPreserveEveryPublishFence(t *testing.T) {
	values, err := validatedOutboxPolicies(validPublisherCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(values.types) != 1 || values.types[0] != "workflow.changed" ||
		values.versions[0] != 1 || values.domains[0] != "workflow" ||
		values.schemaSHA256s[0] != validPublisherCatalog(t).Policies()[0].PayloadSchemaSHA256 {
		t.Fatalf("outbox policy arrays = %#v", values)
	}
}
