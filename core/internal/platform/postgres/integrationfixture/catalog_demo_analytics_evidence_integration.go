//go:build integration

package integrationfixture

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type CatalogDemoAnalyticsReferenceEvidence struct {
	Count              int
	ActiveCount        int
	ReferenceID        string
	SourceID           string
	SourceVersionID    string
	TargetDefinitionID string
	TargetVersionID    string
	TargetEpoch        int64
	RegisterDigestSize int
	ExactTimestamps    bool
	Unreleased         bool
}

func (environment *CatalogDemoEnvironment) AssertCatalogDemoAnalyticsReferences(
	t *testing.T,
	ctx context.Context,
	componentID string,
	targetDefinitionID string,
	want int,
) CatalogDemoAnalyticsReferenceEvidence {
	t.Helper()
	var evidence CatalogDemoAnalyticsReferenceEvidence
	err := environment.owner.QueryRowContext(
		ctx,
		string(catalogDemoFixture("analytics_reference_counts.fixture")),
		componentID,
	).Scan(&evidence.Count, &evidence.ActiveCount)
	if err != nil {
		t.Fatalf("load Analytics Forms reference counts: %v", err)
	}
	if evidence.Count != want || evidence.ActiveCount != want {
		t.Fatalf(
			"Analytics Forms reference counts=%d/%d want=%d",
			evidence.Count, evidence.ActiveCount, want,
		)
	}
	if want == 0 {
		return evidence
	}
	err = environment.owner.QueryRowContext(
		ctx,
		string(catalogDemoFixture("analytics_reference_row.fixture")),
		componentID,
	).Scan(
		&evidence.ReferenceID, &evidence.SourceID,
		&evidence.SourceVersionID, &evidence.TargetDefinitionID,
		&evidence.TargetVersionID, &evidence.TargetEpoch,
		&evidence.RegisterDigestSize, &evidence.ExactTimestamps,
		&evidence.Unreleased,
	)
	if err != nil {
		t.Fatalf("load Analytics Forms reference evidence: %v", err)
	}
	if evidence.TargetDefinitionID != targetDefinitionID ||
		evidence.TargetEpoch < 1 || evidence.RegisterDigestSize != 32 ||
		!evidence.ExactTimestamps || !evidence.Unreleased ||
		!catalogDemoReferenceUUIDv7(evidence.ReferenceID) ||
		!catalogDemoReferenceUUIDv7(evidence.SourceID) ||
		!catalogDemoReferenceUUIDv7(evidence.SourceVersionID) ||
		!catalogDemoReferenceUUIDv7(evidence.TargetVersionID) {
		t.Fatalf(
			"Analytics Forms reference evidence differs: %+v",
			evidence,
		)
	}
	return evidence
}

func catalogDemoReferenceUUIDv7(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.Version() == 7
}
