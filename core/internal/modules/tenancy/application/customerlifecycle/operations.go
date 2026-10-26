package customerlifecycle

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const microsecond = time.Microsecond

var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)

type AsyncCommand struct {
	Credential       CredentialEnvelope
	TenantID         uuid.UUID
	ExpectedRevision int64
	IdempotencyKey   string
}

func (service *Service) RequestOffboard(ctx context.Context, command AsyncCommand) (Operation, error) {
	return service.requestAsync(ctx, command, OperationOffboard, CapabilityOffboard)
}

func (service *Service) RequestPurge(ctx context.Context, command AsyncCommand) (Operation, error) {
	return service.requestAsync(ctx, command, OperationPurge, CapabilityPurge)
}

func (service *Service) requestAsync(ctx context.Context, command AsyncCommand,
	kind, capability string,
) (Operation, error) {
	if ctx == nil || ctx.Err() != nil || command.TenantID == uuid.Nil ||
		command.ExpectedRevision < 1 || !validIdempotencyKey(command.IdempotencyKey) {
		return Operation{}, ErrInvalid
	}
	edge, err := service.boundary()
	if err != nil {
		return Operation{}, err
	}
	if command.TenantID == edge.controlTenantID {
		return Operation{}, ErrNotFound
	}
	digest := asyncDigest(kind, command)
	purpose := lifecycleStepUpPurpose(kind, edge.controlTenantID, command, digest)
	decision, err := service.authorize(ctx, edge, command.Credential, capability,
		command.TenantID, true, purpose)
	if err != nil {
		return Operation{}, accessError(err)
	}
	now := service.dependencies.Clock.Now().UTC().Truncate(microsecond)
	proof := PurgeProof{}
	if kind == OperationPurge {
		proof, err = service.dependencies.PurgeReadiness.Verify(ctx, command.TenantID)
		if err != nil || !proof.Valid(now) {
			return Operation{}, ErrPurgeProof
		}
	}
	result, err := service.dependencies.Store.Accept(ctx, AsyncStoreRequest{
		EnvironmentID: edge.environmentID, ControlTenantID: edge.controlTenantID,
		ActorID: decision.SubjectID, TargetTenantID: command.TenantID,
		Kind: kind, ExpectedRevision: command.ExpectedRevision,
		IdempotencyKey: command.IdempotencyKey, RequestSHA256: digest,
		Proof: proof, Now: now,
	})
	if err != nil {
		return Operation{}, ErrUnavailable
	}
	return asyncOutcome(result)
}

func validIdempotencyKey(value string) bool { return idempotencyKeyPattern.MatchString(value) }

func asyncDigest(kind string, command AsyncCommand) [sha256.Size]byte {
	document, _ := json.Marshal(struct {
		Kind, TenantID   string
		ExpectedRevision int64
	}{kind, command.TenantID.String(), command.ExpectedRevision})
	return sha256.Sum256(document)
}

func lifecycleStepUpPurpose(kind string, control uuid.UUID, command AsyncCommand,
	digest [sha256.Size]byte,
) [sha256.Size]byte {
	document, _ := json.Marshal(struct {
		Version, Kind, ControlTenantID, TargetTenantID, IdempotencyKey string
		RequestSHA256                                                  [32]byte
	}{"tenancy.customer-lifecycle.step-up.v1", kind, control.String(),
		command.TenantID.String(), command.IdempotencyKey, digest})
	return sha256.Sum256(document)
}

func asyncOutcome(result AsyncStoreResult) (Operation, error) {
	switch result.Outcome {
	case "accepted", "replay":
		return result.Operation, nil
	case "not_found":
		return Operation{}, ErrNotFound
	case "revision_conflict":
		return Operation{}, ErrRevisionConflict
	case "idempotency_conflict":
		return Operation{}, ErrIdempotencyConflict
	case "state_conflict", "proof_required":
		return Operation{}, ErrStateConflict
	default:
		return Operation{}, ErrUnavailable
	}
}
