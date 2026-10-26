//go:build integration

package integrationfixture

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

type CatalogDemoEvidence struct {
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
	CatalogReady   bool
	FormsReady     bool
	RequestEqual   bool
	ReadbackEqual  bool
	ReadbackSHA256 string
	RecordA        string
	RecordB        string
	DescriptionA   string
	DescriptionB   string
}

func (environment *CatalogDemoEnvironment) AssertCatalogDemoEvidence(
	t *testing.T,
	ctx context.Context,
	legacyA string,
	legacyB string,
	external string,
) CatalogDemoEvidence {
	t.Helper()
	var evidence CatalogDemoEvidence
	err := environment.owner.QueryRowContext(
		ctx, string(catalogDemoFixture("evidence_counts.fixture")),
	).Scan(
		&evidence.ProvisionCount, &evidence.HandoffCount,
		&evidence.RunCount, &evidence.RowCount,
		&evidence.ReceiptCount, &evidence.AuditCount, &evidence.OutboxCount,
		&evidence.CreateCount, &evidence.UpdateCount,
		&evidence.VersionTwoRows, &evidence.VersionOneRows,
		&evidence.CatalogReady, &evidence.FormsReady,
		&evidence.RequestEqual, &evidence.ReadbackEqual,
		&evidence.ReadbackSHA256,
	)
	if err != nil {
		t.Fatalf("load Catalog demo evidence counts: %v", err)
	}
	if evidence.ProvisionCount != 1 || evidence.HandoffCount != 1 ||
		evidence.RunCount != 1 || evidence.RowCount != 2 ||
		evidence.ReceiptCount != 3 || evidence.AuditCount != 3 ||
		evidence.OutboxCount != 3 || evidence.CreateCount != 2 ||
		evidence.UpdateCount != 1 || evidence.VersionTwoRows != 1 ||
		evidence.VersionOneRows != 1 || !evidence.CatalogReady ||
		!evidence.FormsReady || !evidence.RequestEqual ||
		!evidence.ReadbackEqual || len(evidence.ReadbackSHA256) != 64 {
		t.Fatalf("Catalog demo evidence counts differ: %+v", evidence)
	}
	descriptionID := environment.catalogDemoDescriptionField(t, ctx)
	records := environment.catalogDemoEvidenceRows(t, ctx, descriptionID)
	recordA, foundA := records[legacyA]
	recordB, foundB := records[legacyB]
	if !foundA || !foundB {
		t.Fatalf("Catalog demo rows missing A=%v B=%v rows=%+v", foundA, foundB, records)
	}
	if recordA.version != 2 || recordB.version != 1 ||
		recordA.description != recordB.id || recordB.description != external {
		t.Fatalf("Catalog demo forward reference differs: A=%+v B=%+v", recordA, recordB)
	}
	for label, id := range map[string]string{"A": recordA.id, "B": recordB.id} {
		parsed, parseErr := uuid.Parse(id)
		if parseErr != nil || parsed.Version() != 7 {
			t.Fatalf("Catalog demo real record %s is not UUIDv7: %s %v", label, id, parseErr)
		}
	}
	evidence.RecordA, evidence.RecordB = recordA.id, recordB.id
	evidence.DescriptionA, evidence.DescriptionB =
		recordA.description, recordB.description
	return evidence
}

func (environment *CatalogDemoEnvironment) catalogDemoDescriptionField(
	t *testing.T, ctx context.Context,
) string {
	t.Helper()
	var document []byte
	if err := environment.owner.QueryRowContext(
		ctx, string(catalogDemoFixture("evidence_fields.fixture")),
	).Scan(&document); err != nil {
		t.Fatalf("load Catalog demo field catalog: %v", err)
	}
	var fields []struct {
		FieldID    string `json:"field_id"`
		PublicName string `json:"public_name"`
	}
	if err := json.Unmarshal(document, &fields); err != nil {
		t.Fatalf("decode Catalog demo field catalog: %v", err)
	}
	for _, field := range fields {
		if field.PublicName == "description" {
			if parsed, err := uuid.Parse(field.FieldID); err == nil &&
				parsed.Version() == 7 {
				return field.FieldID
			}
		}
	}
	t.Fatalf("Catalog demo description field is missing: %s", document)
	return ""
}

type catalogDemoEvidenceRow struct {
	id          string
	version     int
	description string
}

func (environment *CatalogDemoEnvironment) catalogDemoEvidenceRows(
	t *testing.T, ctx context.Context, descriptionID string,
) map[string]catalogDemoEvidenceRow {
	t.Helper()
	rows, err := environment.owner.QueryContext(
		ctx, string(catalogDemoFixture("evidence_rows.fixture")),
	)
	if err != nil {
		t.Fatalf("load Catalog demo evidence rows: %v", err)
	}
	defer rows.Close()
	result := make(map[string]catalogDemoEvidenceRow, 2)
	for rows.Next() {
		var legacyID, realID string
		var version int
		var document []byte
		if err = rows.Scan(&legacyID, &realID, &version, &document); err != nil {
			t.Fatalf("scan Catalog demo evidence row: %v", err)
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(document, &fields); err != nil {
			t.Fatalf("decode Catalog demo record document: %v", err)
		}
		var description string
		if err = json.Unmarshal(fields[descriptionID], &description); err != nil {
			t.Fatalf("decode Catalog demo description: %v document=%s", err, document)
		}
		result[legacyID] = catalogDemoEvidenceRow{
			id: realID, version: version, description: description,
		}
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("iterate Catalog demo evidence rows: %v", err)
	}
	return result
}
