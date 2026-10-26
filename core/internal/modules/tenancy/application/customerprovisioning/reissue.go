package customerprovisioning

import (
	"context"

	"github.com/google/uuid"
)

type ReissueCommand struct {
	Caller         CallerKind
	TenantID       uuid.UUID
	IdempotencyKey string
	Credential     CredentialEnvelope
}

func (coordinator *Coordinator) ReissueInitialAdminSetup(
	ctx context.Context,
	command ReissueCommand,
) (Operation, error) {
	if coordinator == nil || ctx == nil || ctx.Err() != nil ||
		!command.Caller.valid() || command.TenantID == uuid.Nil ||
		!validIdempotencyKey(command.IdempotencyKey) {
		return Operation{}, ErrInvalid
	}
	policy, primary, err := coordinator.saaSBoundary()
	if err != nil {
		return Operation{}, err
	}
	controlID, err := uuid.Parse(primary.TenantID())
	if err != nil || controlID == command.TenantID {
		return Operation{}, ErrForbidden
	}
	accessRequest := AccessRequest{
		Caller: command.Caller, Capability: CapabilityAdminSetupReissue,
		ControlTenantID: controlID, TargetTenantID: command.TenantID,
		IdempotencyKey: command.IdempotencyKey, Credential: command.Credential,
	}
	decision, err := coordinator.dependencies.Access.Verify(ctx, accessRequest)
	if err != nil {
		return Operation{}, classifyAccess(err)
	}
	if !decision.validFor(accessRequest, coordinator.now()) ||
		decision.DeploymentMode != policy.Mode() {
		return Operation{}, ErrForbidden
	}
	target, err := coordinator.dependencies.Store.LoadSetupReissueTarget(
		ctx, command.TenantID,
	)
	if err != nil {
		return Operation{}, classifyStore(err)
	}
	if !validSetupReissueTarget(target, command.TenantID) {
		return Operation{}, ErrConflict
	}
	reissueID := stableID(
		target.Operation.OperationID,
		"initial-admin-setup-reissue:"+command.IdempotencyKey,
	)
	if target.Operation.SetupDeliveryState == "reissued" &&
		target.ReissueOperationID != reissueID {
		return Operation{}, ErrConflict
	}
	err = coordinator.dependencies.Setup.ReissueInitialAdmin(ctx, SetupReissueRequest{
		OperationID: target.Operation.OperationID, ReissueOperationID: reissueID,
		TenantID: command.TenantID, UserID: target.UserID,
		SchemaName: target.SchemaName, Email: target.Email, CreatedAt: coordinator.now(),
	})
	if err != nil {
		return Operation{}, ErrUnavailable
	}
	target.Operation.SetupDeliveryState = "reissued"
	target.Operation.UpdatedAt = coordinator.now()
	return target.Operation, nil
}

func validSetupReissueTarget(target SetupReissueTarget, tenantID uuid.UUID) bool {
	return target.Operation.valid() && target.Operation.TenantID == tenantID &&
		target.Operation.State == "succeeded" && target.SchemaName != "" &&
		target.UserID != uuid.Nil && target.Email != "" &&
		(target.Operation.SetupDeliveryState == "permanent_failure" ||
			target.Operation.SetupDeliveryState == "outcome_unknown" ||
			target.Operation.SetupDeliveryState == "reissued")
}
