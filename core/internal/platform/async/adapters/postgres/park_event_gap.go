package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (repository *consumerRepository) parkEventGap(
	ctx context.Context,
	envelopeID string,
	record eventGapRecord,
) (ParkedSequenceGap, error) {
	envelope := record.evidence.Envelope
	tenantID := *envelope.TenantID
	reasonClass, reasonCode := eventGapReason()
	var rootID string
	var jobID string
	var created bool
	err := repository.database.QueryRow(
		ctx, parkEventGapSQL,
		record.rootID, record.jobID, tenantID, "tenant", repository.authority.scopeID,
		repository.authority.groupID, repository.authority.groupGeneration,
		repository.authority.domain, repository.authority.generation,
		envelope.EventID, repository.consumer, repository.consumerVersion,
		envelope.ContractVersion, envelope.EventVersion, record.sourceSHA,
		string(reasonClass), reasonCode, record.errorSHA, record.canonical,
		record.sourceSHA, record.redacted, repository.gapPolicy.Kind,
		repository.gapPolicy.PayloadRevision, record.payloadRef, record.dedupeKey,
		repository.gapPolicy.MaximumAttempts,
	).Scan(&rootID, &jobID, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		return ParkedSequenceGap{}, ErrEventGapConflict
	}
	if err != nil {
		return ParkedSequenceGap{}, classifyConsumerDatabaseError("park event sequence gap", err)
	}
	if rootID != record.rootID.String() || jobID != record.jobID.String() {
		return ParkedSequenceGap{}, fmt.Errorf("park event sequence gap identity: %w", ErrEventGapConflict)
	}
	return ParkedSequenceGap{
		eventID: envelopeID, expected: record.evidence.ExpectedSequence,
		received: record.evidence.ReceivedSequence,
		rootID:   rootID, jobID: jobID, created: created,
	}, nil
}
