//go:build integration

package integrationfixture

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

type CatalogDemoCrossPackageEvidence struct {
	ProvisionCount int
	HandoffCount   int
	RunCount       int
	RowCount       int
	ReceiptCount   int
	AuditCount     int
	OutboxCount    int
	CreateCount    int
	UpdateCount    int
	VersionTwoRows int
	VersionOneRows int
	LegacyID       string
	RecordID       string
	RowVersion     int
	Description    string
}

func (environment *CatalogDemoEnvironment) AssertCatalogDemoCrossPackageEvidence(
	t *testing.T,
	ctx context.Context,
	packageID string,
	legacyID string,
	expectedReference string,
) CatalogDemoCrossPackageEvidence {
	t.Helper()
	var evidence CatalogDemoCrossPackageEvidence
	err := environment.owner.QueryRowContext(
		ctx, string(catalogDemoFixture("cross_package_counts.fixture")),
	).Scan(
		&evidence.ProvisionCount, &evidence.HandoffCount,
		&evidence.RunCount, &evidence.RowCount,
		&evidence.ReceiptCount, &evidence.AuditCount, &evidence.OutboxCount,
		&evidence.CreateCount, &evidence.UpdateCount,
		&evidence.VersionTwoRows, &evidence.VersionOneRows,
	)
	if err != nil {
		t.Fatalf("load cross-package Catalog demo evidence counts: %v", err)
	}
	if evidence.ProvisionCount != 2 || evidence.HandoffCount != 2 ||
		evidence.RunCount != 2 || evidence.RowCount != 3 ||
		evidence.ReceiptCount != 5 || evidence.AuditCount != 5 ||
		evidence.OutboxCount != 5 || evidence.CreateCount != 3 ||
		evidence.UpdateCount != 2 || evidence.VersionTwoRows != 2 ||
		evidence.VersionOneRows != 1 {
		t.Fatalf("cross-package Catalog demo counts differ: %+v", evidence)
	}
	descriptionID := environment.catalogDemoCrossPackageDescriptionField(
		t, ctx, packageID,
	)
	var document []byte
	err = environment.owner.QueryRowContext(
		ctx, string(catalogDemoFixture("cross_package_row.fixture")), packageID,
	).Scan(
		&evidence.LegacyID, &evidence.RecordID, &evidence.RowVersion, &document,
	)
	if err != nil {
		t.Fatalf("load cross-package Catalog demo row: %v", err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(document, &fields); err != nil {
		t.Fatalf("decode cross-package Catalog demo row: %v", err)
	}
	if err = json.Unmarshal(fields[descriptionID], &evidence.Description); err != nil {
		t.Fatalf("decode cross-package Catalog demo reference: %v", err)
	}
	recordID, parseErr := uuid.Parse(evidence.RecordID)
	if evidence.LegacyID != legacyID || evidence.RowVersion != 2 ||
		evidence.Description != expectedReference || parseErr != nil ||
		recordID.Version() != 7 {
		t.Fatalf("cross-package Catalog demo rewrite differs: %+v err=%v",
			evidence, parseErr)
	}
	return evidence
}

func (environment *CatalogDemoEnvironment) catalogDemoCrossPackageDescriptionField(
	t *testing.T, ctx context.Context, packageID string,
) string {
	t.Helper()
	var document []byte
	err := environment.owner.QueryRowContext(
		ctx, string(catalogDemoFixture("cross_package_fields.fixture")), packageID,
	).Scan(&document)
	if err != nil {
		t.Fatalf("load cross-package field catalog: %v", err)
	}
	var fields []struct {
		FieldID    string `json:"field_id"`
		PublicName string `json:"public_name"`
	}
	if err = json.Unmarshal(document, &fields); err != nil {
		t.Fatalf("decode cross-package field catalog: %v", err)
	}
	for _, field := range fields {
		if field.PublicName == "description" {
			return field.FieldID
		}
	}
	t.Fatalf("cross-package description field is missing: %s", document)
	return ""
}
