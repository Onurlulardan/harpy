package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const terminalGapJobKindPrefix = "platform_async.event_gap_reconcile."

type gapRetryTerminalReference struct {
	ContractVersion  string `json:"contract_version"`
	JobID            string `json:"job_id"`
	Kind             string `json:"kind"`
	PayloadReference string `json:"payload_reference"`
	PayloadSHA256    string `json:"payload_sha256"`
}

func lockEventGapJobRetryEvidence(
	ctx context.Context,
	transaction pgx.Tx,
	root deadLetterRootRow,
) error {
	if !strings.HasPrefix(root.sourceContract, terminalGapJobKindPrefix) {
		return nil
	}
	var reference gapRetryTerminalReference
	if err := json.Unmarshal(root.payload, &reference); err != nil ||
		reference.ContractVersion != "harpy.job-terminal-reference.v1" ||
		reference.JobID != root.sourceID.String() ||
		reference.Kind != root.sourceContract ||
		reference.PayloadSHA256 != root.sourceSHA256 ||
		!strings.HasPrefix(reference.PayloadReference, "deadletter:event-gap:") {
		return ErrDeadLetterRetryInvalid
	}
	gapIDText := strings.TrimPrefix(
		reference.PayloadReference, "deadletter:event-gap:",
	)
	gapID, err := uuid.Parse(gapIDText)
	if err != nil || gapID.String() != gapIDText {
		return ErrDeadLetterRetryInvalid
	}
	evidence, err := lockDeadLetterRoot(ctx, transaction, gapID, root.tenantID)
	if errors.Is(err, ErrDeadLetterPayloadExpired) {
		return err
	}
	if err != nil || evidence.sourceKind != deadletter.SourceEvent ||
		evidence.payloadSHA256 != root.sourceSHA256 {
		return ErrDeadLetterRetryInvalid
	}
	return nil
}

func insertDeadLetterJobRetry(
	ctx context.Context,
	transaction pgx.Tx,
	root deadLetterRootRow,
	jobID uuid.UUID,
) error {
	dedupeKey := "deadletter-job-retry:" + jobID.String()
	var inserted uuid.UUID
	err := transaction.QueryRow(
		ctx, insertDeadLetterJobRetrySQL, root.id, jobID, dedupeKey,
	).Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDeadLetterRetryInvalid
	}
	if isDeadLetterAuthorityError(err) {
		return ErrDeadLetterAuthority
	}
	if err != nil {
		return fmt.Errorf("enqueue dead-letter job retry: %w", err)
	}
	if inserted != jobID {
		return ErrDeadLetterRetryInvalid
	}
	return nil
}
