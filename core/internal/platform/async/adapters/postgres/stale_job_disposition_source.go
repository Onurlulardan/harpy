package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type staleJobDispositionSource struct {
	jobID, scopeID                    uuid.UUID
	tenantID                          pgtype.UUID
	scopeKind, groupID, domain        string
	groupGeneration, domainGeneration int64
	writerOrigin, kind                string
	payloadRevision                   int32
	payloadReference                  pgtype.Text
	payloadSHA256, dedupeKey          string
	effectClass                       string
	effectKey                         pgtype.Text
	priority, maximumAttempts         int32
	runAt                             time.Time
	state                             string
	parkedCode                        pgtype.Text
	parkedAt, effectStartedAt         pgtype.Timestamptz
	observedAt                        time.Time
}

type preparedStaleJobDisposition struct {
	rootID, actionID, resultJobID, resultRootID  uuid.UUID
	resultPayloadReference, resultDedupeKey      string
	canonical                                    []byte
	digest, sourceRunAt, resultRunAt, occurredAt string
}

func (source staleJobDispositionSource) prepare(
	authority JobAuthority,
	command StaleJobDispositionCommand,
) (preparedStaleJobDisposition, error) {
	if !source.valid(authority, command) {
		return preparedStaleJobDisposition{}, ErrStaleJobDispositionConflict
	}
	generation := fmt.Sprintf("%s:%d:%s:%d", authority.groupID,
		authority.groupGeneration, authority.domain, authority.generation)
	prepared := preparedStaleJobDisposition{
		rootID: deterministicEventGapID(
			"stale-job-disposition-root.v1", source.jobID.String(), generation,
		),
	}
	prepared.actionID = deterministicEventGapID(
		"stale-job-disposition-action.v1", prepared.rootID.String(),
		string(command.decision), command.reason, command.ticket,
		command.actorID.String(), command.causationID.String(),
	)
	prepared.sourceRunAt = source.runAt.UTC().Format(time.RFC3339Nano)
	if source.state == "cancelled" {
		if command.decision == jobsv1.StaleJobAdopt {
			prepared.resultJobID = deterministicEventGapID(
				"stale-job-adoption.v1", source.jobID.String(), generation,
			)
		}
		return prepared, nil
	}
	if command.decision == jobsv1.StaleJobAdopt {
		prepared.resultJobID = deterministicEventGapID(
			"stale-job-adoption.v1", source.jobID.String(), generation,
		)
		prepared.resultDedupeKey = fmt.Sprintf(
			"stale-adopt:%s:g%d:d%d", source.jobID,
			authority.groupGeneration, authority.generation,
		)
		prepared.resultPayloadReference = source.payloadReference.String
		resultRunAt := source.runAt
		if source.observedAt.After(resultRunAt) {
			resultRunAt = source.observedAt
		}
		prepared.resultRunAt = resultRunAt.UTC().Format(time.RFC3339Nano)
		if source.isEventGap() {
			prepared.resultRootID = deterministicEventGapID(
				"stale-event-gap-root.v2", source.jobID.String(), generation,
			)
			prepared.resultPayloadReference = "deadletter:event-gap:" +
				prepared.resultRootID.String()
		}
	}
	contract, err := source.contract(authority, command, prepared)
	if err != nil {
		return preparedStaleJobDisposition{}, err
	}
	document, err := json.Marshal(contract)
	if err != nil {
		return preparedStaleJobDisposition{}, fmt.Errorf("marshal stale job disposition: %w", err)
	}
	prepared.canonical, err = canonicaljsonv1.Canonicalize(document, 65_536)
	if err != nil {
		return preparedStaleJobDisposition{}, fmt.Errorf("canonicalize stale job disposition: %w", err)
	}
	prepared.digest = digestEventGap(prepared.canonical)
	prepared.occurredAt = contract.OccurredAt
	return prepared, nil
}

func (source staleJobDispositionSource) valid(
	authority JobAuthority,
	command StaleJobDispositionCommand,
) bool {
	effectValid := source.effectClass == string(jobsv1.EffectDatabaseTransactional) &&
		!source.effectKey.Valid ||
		source.effectClass == string(jobsv1.EffectExternal) && source.effectKey.Valid
	baseValid := command.valid() && command.jobID == source.jobID &&
		authority.valid(authority.owner) && source.scopeKind == authority.scopeKind &&
		source.scopeID == authority.scopeID && source.groupID == authority.groupID &&
		source.domain == authority.domain && source.groupGeneration <= authority.groupGeneration &&
		source.domainGeneration <= authority.generation &&
		(source.groupGeneration < authority.groupGeneration ||
			source.domainGeneration < authority.generation) &&
		(source.writerOrigin == "v1" || source.writerOrigin == "v2") &&
		claimKindPattern.MatchString(source.kind) && source.payloadRevision > 0 &&
		claimDigestPattern.MatchString(source.payloadSHA256) &&
		len(source.dedupeKey) > 0 && len(source.dedupeKey) <= jobsv1.MaximumKeyBytes &&
		effectValid && source.maximumAttempts >= 1 && source.maximumAttempts <= 25 &&
		(source.state == "queued" || source.state == "cancelled") &&
		source.parkedAt.Valid && source.parkedCode.Valid &&
		source.parkedCode.String == "JOB_AUTHORITY_STALE" &&
		!source.effectStartedAt.Valid && !source.runAt.IsZero() &&
		!source.observedAt.IsZero()
	if !baseValid {
		return false
	}
	if source.state == "cancelled" {
		return !source.payloadReference.Valid ||
			len(source.payloadReference.String) > 0 &&
				len(source.payloadReference.String) <= jobsv1.MaximumReferenceBytes
	}
	return source.payloadReference.Valid && len(source.payloadReference.String) > 0 &&
		len(source.payloadReference.String) <= jobsv1.MaximumReferenceBytes
}

func (source staleJobDispositionSource) isEventGap() bool {
	return source.kind == "platform_async.event_gap_reconcile."+source.domain &&
		len(source.payloadReference.String) > len("deadletter:event-gap:") &&
		source.payloadReference.String[:len("deadletter:event-gap:")] == "deadletter:event-gap:"
}

func optionalText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	copy := value.String
	return &copy
}
