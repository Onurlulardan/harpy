package lifecycle

import (
	"context"
	"time"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type purgeOrderOwner struct {
	componentowner.Owner
	blocked                    bool
	beginWaiting               bool
	preflight, begin, finalize int
}

func (owner *purgeOrderOwner) RetireComponent(
	context.Context, platformtx.TenantScope, componentowner.LifecycleCommand,
) (componentowner.OwnerResult, error) {
	return componentowner.OwnerResult{}, ErrUnavailable
}

func (owner *purgeOrderOwner) RestoreComponent(
	context.Context, platformtx.TenantScope, componentowner.LifecycleCommand,
) (componentowner.OwnerResult, error) {
	return componentowner.OwnerResult{}, ErrUnavailable
}

func (owner *purgeOrderOwner) PreflightComponentPurge(
	_ context.Context, _ platformtx.TenantScope, command componentowner.PurgeCommand,
) (componentowner.OwnerResult, error) {
	owner.preflight++
	state := catalogcomponentv2.StateReady
	if owner.blocked {
		state = catalogcomponentv2.StateBlocked
	}
	return purgeOrderResult(command, state), nil
}

func (owner *purgeOrderOwner) BeginComponentPurge(
	_ context.Context, _ platformtx.TenantScope, command componentowner.PurgeCommand,
) (componentowner.OwnerResult, error) {
	owner.begin++
	if owner.beginWaiting {
		result := purgeOrderResult(command, catalogcomponentv2.StateWaiting)
		result.Ready = false
		result.CompletedAt = time.Time{}
		result.RetryAt = command.Authority.RequestedAt.Add(time.Minute)
		result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
		return result, nil
	}
	return purgeOrderResult(command, catalogcomponentv2.StatePurgePending), nil
}

func (owner *purgeOrderOwner) FinalizeComponentPurge(
	_ context.Context, _ platformtx.TenantScope, command componentowner.PurgeCommand,
) (componentowner.OwnerResult, error) {
	owner.finalize++
	return purgeOrderResult(command, catalogcomponentv2.StatePurged), nil
}

type purgeOrderDirectory map[string]*purgeOrderOwner

func (directory purgeOrderDirectory) Resolve(name string) (componentowner.Owner, error) {
	owner := directory[name]
	if owner == nil {
		return nil, ErrOwnerNotV2
	}
	return owner, nil
}

type purgeOrderStore struct {
	Store
	changes int
}

func (store *purgeOrderStore) ChangeOperation(
	_ context.Context, mutation OperationMutation,
) (catalogdomain.Operation, error) {
	store.changes++
	return mutation.Operation, nil
}

const digestA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func purgeOrderAuthority() PurgeAuthority {
	requested := time.Date(2026, 7, 31, 9, 0, 0, 0, time.UTC)
	value := PurgeAuthority{
		TenantID:    "019fa200-0000-7000-8000-000000000001",
		PackageID:   "019fa200-0000-7000-8000-000000000002",
		PackageSlug: "purge-order", OperationID: "019fa200-0000-7000-8000-000000000003",
		ActorID:           "019fa200-0000-7000-8000-000000000004",
		RequestedBy:       "019fa200-0000-7000-8000-000000000004",
		ApprovedBy:        "019fa200-0000-7000-8000-000000000005",
		ApprovalReceiptID: "019fa200-0000-7000-8000-000000000006",
		ApprovalSHA256:    digestA, BackupReceiptID: "019fa200-0000-7000-8000-000000000007",
		BackupState: catalogcomponentv2.BackupReady, BackupSHA256: digestA,
		ConfirmationSlug: "purge-order", Reason: "ordered destructive cleanup",
		IdempotencyKey: "purge-order-0001", ReferenceEpoch: 3,
		AuthorizationTokenVersion: 3, AuthorizationVersion: 5,
		AuthorizationClock:          7,
		AuthorizationPermissionID:   "019fa200-0000-7000-8000-000000000008",
		AuthorizationEvidenceSHA256: digestA, StepUpPurposeSHA256: digestA,
		StepUpAuthenticatedAt: requested.Add(-time.Minute),
		ApprovalStepUpAt:      requested.Add(-3 * time.Minute),
		ApprovedAt:            requested.Add(-2 * time.Minute),
		ApprovalExpiresAt:     requested.Add(28 * time.Minute),
		RequestedAt:           requested, CoolingUntil: requested.Add(24 * time.Hour),
	}
	value.AuthoritySHA256 = catalogcomponentv2.AuthoritySHA256(value)
	return value
}

func purgeOrderResult(
	command componentowner.PurgeCommand, state string,
) componentowner.OwnerResult {
	value := componentowner.OwnerResult{
		ReceiptID: "019fa200-0000-7000-8000-000000000020",
		TenantID:  command.Authority.TenantID, PackageID: command.Authority.PackageID,
		Component: command.Component, OperationID: command.Authority.OperationID,
		ReferenceEpoch: command.Authority.ReferenceEpoch, State: state,
		AuthoritySHA256: command.Authority.AuthoritySHA256,
	}
	if command.Component.Owner == "forms" {
		value.ResourceID = command.Component.ID
		value.VersionID = "019fa200-0000-7000-8000-000000000021"
		value.HeadID = "019fa200-0000-7000-8000-000000000022"
	}
	switch state {
	case catalogcomponentv2.StateReady, catalogcomponentv2.StatePurgePending,
		catalogcomponentv2.StatePurged:
		value.Ready, value.CompletedAt = true, command.Authority.CoolingUntil
	case catalogcomponentv2.StateBlocked:
		value.Blockers = []componentowner.Blocker{{
			Owner: command.Component.Owner, Code: "ACTIVE_REFERENCE", Digest: digestA,
		}}
	}
	value.ResultSHA256 = catalogcomponentv2.ResultSHA256(value)
	return value
}
