package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func insertDeadLetterEventRetry(
	ctx context.Context,
	transaction pgx.Tx,
	root deadLetterRootRow,
	jobID uuid.UUID,
) error {
	effectKey, err := eventRetryEffectKey(root.redactedEnvelope, root.sourceID)
	if err != nil {
		return err
	}
	kind, err := deadletter.EventRetryKind(root.domain)
	if err != nil {
		return ErrDeadLetterRetryInvalid
	}
	payloadReference := "deadletter:event:" + root.id.String()
	dedupeKey := "deadletter-event-retry:" + jobID.String()
	var inserted uuid.UUID
	err = transaction.QueryRow(
		ctx, insertDeadLetterEventRetrySQL,
		root.id, jobID, kind, payloadReference, dedupeKey,
		effectKey, jobsv1.DefaultMaximumAttempts,
	).Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDeadLetterRetryInvalid
	}
	if err != nil {
		return fmt.Errorf("enqueue dead-letter event retry: %w", err)
	}
	if inserted != jobID {
		return ErrDeadLetterRetryInvalid
	}
	return nil
}

func eventRetryEffectKey(document []byte, eventID uuid.UUID) (string, error) {
	var metadata struct {
		EventID     string `json:"event_id"`
		BusinessKey string `json:"business_effect_idempotency_key"`
	}
	if err := json.Unmarshal(document, &metadata); err != nil ||
		metadata.EventID != eventID.String() || metadata.BusinessKey == "" ||
		!utf8.ValidString(metadata.BusinessKey) || len(metadata.BusinessKey) > 512 ||
		strings.TrimSpace(metadata.BusinessKey) != metadata.BusinessKey {
		return "", ErrDeadLetterRetryInvalid
	}
	return metadata.BusinessKey, nil
}
