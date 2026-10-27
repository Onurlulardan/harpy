package customerlifecycle

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

type MutationCommand struct {
	Credential       CredentialEnvelope
	TenantID         uuid.UUID
	ExpectedRevision int64
	IdempotencyKey   string
	Name             string
}

func (service *Service) Update(ctx context.Context, command MutationCommand) (Tenant, error) {
	if strings.TrimSpace(command.Name) != command.Name || len(command.Name) == 0 ||
		len(command.Name) > 255 {
		return Tenant{}, ErrInvalid
	}
	return service.mutate(ctx, command, MutationUpdate, CapabilityUpdate)
}

func (service *Service) Suspend(ctx context.Context, command MutationCommand) (Tenant, error) {
	command.Name = ""
	return service.mutate(ctx, command, MutationSuspend, CapabilitySuspend)
}

func (service *Service) Reactivate(ctx context.Context, command MutationCommand) (Tenant, error) {
	command.Name = ""
	return service.mutate(ctx, command, MutationReactivate, CapabilityReactivate)
}

func (service *Service) mutate(ctx context.Context, command MutationCommand,
	kind, capability string,
) (Tenant, error) {
	if ctx == nil || ctx.Err() != nil || command.TenantID == uuid.Nil ||
		command.ExpectedRevision < 1 || !validIdempotencyKey(command.IdempotencyKey) {
		return Tenant{}, ErrInvalid
	}
	edge, err := service.boundary()
	if err != nil {
		return Tenant{}, err
	}
	if command.TenantID == edge.controlTenantID {
		return Tenant{}, ErrNotFound
	}
	decision, err := service.authorize(ctx, edge, command.Credential, capability,
		command.TenantID, false, [32]byte{})
	if err != nil {
		return Tenant{}, accessError(err)
	}
	now := service.dependencies.Clock.Now().UTC().Truncate(microsecond)
	digest := mutationDigest(kind, command)
	result, err := service.dependencies.Store.Mutate(ctx, MutationStoreRequest{
		EnvironmentID: edge.environmentID, ControlTenantID: edge.controlTenantID,
		ActorID:        decision.SubjectID,
		TargetTenantID: command.TenantID, Kind: kind, Name: command.Name,
		ExpectedRevision: command.ExpectedRevision, IdempotencyKey: command.IdempotencyKey,
		RequestSHA256: digest, Now: now,
	})
	if err != nil {
		return Tenant{}, ErrUnavailable
	}
	return mutationOutcome(result)
}

func mutationDigest(kind string, command MutationCommand) [sha256.Size]byte {
	document, _ := json.Marshal(struct {
		Kind, TenantID, Name string
		ExpectedRevision     int64
	}{kind, command.TenantID.String(), command.Name, command.ExpectedRevision})
	return sha256.Sum256(document)
}

func mutationOutcome(result MutationStoreResult) (Tenant, error) {
	switch result.Outcome {
	case "applied", "replay":
		return result.Tenant, nil
	case "not_found":
		return Tenant{}, ErrNotFound
	case "revision_conflict":
		return Tenant{}, ErrRevisionConflict
	case "idempotency_conflict":
		return Tenant{}, ErrIdempotencyConflict
	case "state_conflict":
		return Tenant{}, ErrStateConflict
	default:
		return Tenant{}, ErrUnavailable
	}
}
