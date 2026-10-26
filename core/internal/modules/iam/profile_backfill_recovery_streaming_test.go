package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestProfileRecoveryClassifierWipesDocumentsOnSuccess(t *testing.T) {
	classifier, source := profileRecoveryClassifierFixture(t)
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"Alice"}`)
	resources := []profilebackfillapp.SourceResource{profileRecoveryResource(
		"14000000-0000-4000-8000-000000000001", source.ID(), document,
	)}
	bytes := resources[0].Document
	err := classifier.AppendBatch(
		context.Background(), map[uuid.UUID]profileapp.Schema{source.ID(): source},
		resources,
	)
	if err != nil {
		t.Fatal(err)
	}
	assertProfileRecoveryDocumentWiped(t, resources[0].Document, bytes)
	inventory, err := classifier.Finish(context.Background())
	if err != nil || inventory.RemainingSourceCount() != 1 {
		t.Fatalf("inventory = %+v/%v", inventory, err)
	}
}

func TestProfileRecoveryClassifierWipesWholeBatchOnFailure(t *testing.T) {
	classifier, source := profileRecoveryClassifierFixture(t)
	first := []byte(`{"12000000-0000-4000-8000-000000000001":"Alice"}`)
	second := []byte("{}")
	resources := []profilebackfillapp.SourceResource{
		profileRecoveryResource(
			"14000000-0000-4000-8000-000000000001", source.ID(), first,
		),
		profileRecoveryResource(
			"14000000-0000-4000-8000-000000000002",
			uuid.MustParse("13000000-0000-4000-8000-000000000099"), second,
		),
	}
	err := classifier.AppendBatch(
		context.Background(), map[uuid.UUID]profileapp.Schema{source.ID(): source},
		resources,
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("classification error = %v", err)
	}
	assertProfileRecoveryDocumentWiped(t, resources[0].Document, first)
	assertProfileRecoveryDocumentWiped(t, resources[1].Document, second)
}

func TestProfileRecoveryClassifierUsesInjectedTimeBudget(t *testing.T) {
	startedAt := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	now := startedAt
	classifier, source := profileRecoveryClassifierAt(
		t, func() time.Time { return now },
	)
	now = startedAt.Add(profilebackfillapp.MaximumRecoveryInventoryDuration)
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"Alice"}`)
	resources := []profilebackfillapp.SourceResource{profileRecoveryResource(
		"14000000-0000-4000-8000-000000000001", source.ID(), document,
	)}
	err := classifier.AppendBatch(
		context.Background(), map[uuid.UUID]profileapp.Schema{source.ID(): source},
		resources,
	)
	if !errors.Is(err, profilebackfillapp.ErrBudgetExceeded) {
		t.Fatalf("time budget error = %v", err)
	}
	assertProfileRecoveryDocumentWiped(t, resources[0].Document, document)
}

func profileRecoveryClassifierFixture(
	t *testing.T,
) (*ProfileRecoveryInventoryClassifier, profileapp.Schema) {
	t.Helper()
	now := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	return profileRecoveryClassifierAt(t, func() time.Time { return now })
}

func profileRecoveryClassifierAt(
	t *testing.T,
	clock profilebackfillapp.RecoveryInventoryClock,
) (*ProfileRecoveryInventoryClassifier, profileapp.Schema) {
	t.Helper()
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 64)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 64)
	classifier, err := NewProfileRecoveryInventoryClassifier(
		ProfileRecoveryInventoryInput{
			Parent: profileRecoveryParent(t, source, target),
			Source: source, Target: target, Normalizer: profileHTTPNormalizer{},
			Clock: clock,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return classifier, source
}

func assertProfileRecoveryDocumentWiped(
	t *testing.T,
	document []byte,
	bytes []byte,
) {
	t.Helper()
	if document != nil {
		t.Fatalf("document reference retained: %q", document)
	}
	for index, value := range bytes {
		if value != 0 {
			t.Fatalf("document byte %d not wiped", index)
		}
	}
}
