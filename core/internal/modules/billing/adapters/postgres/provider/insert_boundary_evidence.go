package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (storage *processorStore) insertBoundaryEvidence(
	ctx context.Context,
	row webhookRow,
	eventID uuid.UUID,
	fact subscriptionFact,
	recordedAt time.Time,
) error {
	evidenceID, err := storage.nextID()
	if err != nil || evidenceID == uuid.Nil ||
		evidenceID.Version() != uuid.Version(7) || fact.boundary == nil {
		return fmt.Errorf("create billing boundary evidence identity")
	}
	var nextStart, nextEnd, graceEnd *time.Time
	switch fact.boundary.outcome {
	case "paid":
		start, end := fact.period.Start(), fact.period.End()
		nextStart, nextEnd = &start, &end
	case "grace":
		graceEnd = fact.graceEndsAt
	}
	var inserted uuid.UUID
	err = storage.database.queryRow(
		ctx, insertBoundaryEvidenceSQL,
		evidenceID, row.authorityID, storage.tenantID, fact.subscriptionID,
		fact.boundary.jobID, eventID, fact.boundary.kind, fact.boundary.dueAt,
		fact.boundary.expectedRevision, fact.boundary.outcome,
		nextStart, nextEnd, graceEnd, recordedAt,
	).Scan(&inserted)
	if err != nil || inserted != evidenceID {
		return fmt.Errorf("insert billing boundary evidence: %w", err)
	}
	return nil
}
