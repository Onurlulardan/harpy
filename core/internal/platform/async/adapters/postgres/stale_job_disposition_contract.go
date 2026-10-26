package postgres

import (
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func (source staleJobDispositionSource) contract(
	authority JobAuthority,
	command StaleJobDispositionCommand,
	prepared preparedStaleJobDisposition,
) (jobsv1.StaleJobDisposition, error) {
	sourcePayload := jobsv1.StaleJobPayloadIdentity{
		Revision: int(source.payloadRevision), SHA256: source.payloadSHA256,
		Reference: source.payloadReference.String, DedupeKey: source.dedupeKey,
		EffectKey: optionalText(source.effectKey),
	}
	value := jobsv1.StaleJobDisposition{
		EvidenceRootID:  prepared.rootID.String(),
		DispositionID:   prepared.actionID.String(),
		SourceJobID:     source.jobID.String(),
		Decision:        command.decision,
		Kind:            source.kind,
		EffectClass:     jobsv1.EffectClass(source.effectClass),
		MaximumAttempts: int(source.maximumAttempts),
		Priority:        source.priority,
		OldAuthority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeKind(source.scopeKind), ScopeID: source.scopeID.String(),
			Domain: source.domain, CutoverGeneration: source.domainGeneration,
			WriterOrigin: jobsv1.WriterOrigin(source.writerOrigin), GroupID: source.groupID,
			GroupGeneration: source.groupGeneration,
		},
		CurrentAuthority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeKind(authority.scopeKind), ScopeID: authority.scopeID.String(),
			Domain: authority.domain, CutoverGeneration: authority.generation,
			WriterOrigin: jobsv1.WriterV2, GroupID: authority.groupID,
			GroupGeneration: authority.groupGeneration,
		},
		SourcePayload: sourcePayload, SourceRunAt: prepared.sourceRunAt,
		Reason:          command.reason,
		TicketReference: command.ticket, ActorID: command.actorID.String(),
		CausationID: command.causationID.String(),
		OccurredAt:  source.observedAt.UTC().Format(time.RFC3339Nano),
	}
	if command.decision == jobsv1.StaleJobAdopt {
		resultPayload := sourcePayload
		resultPayload.Reference = prepared.resultPayloadReference
		resultPayload.DedupeKey = prepared.resultDedupeKey
		resultID := prepared.resultJobID.String()
		value.ResultJobID = &resultID
		value.ResultPayload = &resultPayload
		value.ResultRunAt = &prepared.resultRunAt
	}
	return jobsv1.NewStaleJobDisposition(value)
}
