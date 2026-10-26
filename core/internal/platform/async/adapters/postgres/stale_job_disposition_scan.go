package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func inspectStaleJobDisposition(
	ctx context.Context,
	transaction pgx.Tx,
	authority JobAuthority,
	jobID uuid.UUID,
) (staleJobDispositionSource, error) {
	var source staleJobDispositionSource
	err := transaction.QueryRow(
		ctx, inspectStaleJobDispositionSQL, jobID, authority.scopeKind,
		authority.scopeID, authority.groupID, authority.domain,
	).Scan(
		&source.jobID, &source.tenantID, &source.scopeKind, &source.scopeID,
		&source.groupID, &source.groupGeneration, &source.domain,
		&source.domainGeneration, &source.writerOrigin, &source.kind,
		&source.payloadRevision, &source.payloadReference, &source.payloadSHA256,
		&source.dedupeKey, &source.effectClass, &source.effectKey,
		&source.priority, &source.runAt, &source.maximumAttempts, &source.state,
		&source.parkedAt, &source.parkedCode, &source.effectStartedAt,
		&source.observedAt,
	)
	return source, err
}
