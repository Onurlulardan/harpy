package lifecycle

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type PortalSetupQuery struct {
	BindingID, ActorID, PermissionID                   uuid.UUID
	ExpectedTokenVersion, ExpectedAuthorizationVersion int64
}

type PortalSetup struct {
	BindingID, UserID                         uuid.UUID
	SetupState, ApprovalState                 string
	Revision                                  int64
	SetupCompletedAt, SubmittedAt, DecisionAt time.Time
	UpdatedAt                                 time.Time
}

func (value PortalSetup) Validate() error {
	if value.BindingID == uuid.Nil || value.UserID == uuid.Nil ||
		!validPortalSetupState(value.SetupState) ||
		!validPortalApprovalState(value.ApprovalState) ||
		value.Revision < 1 || value.UpdatedAt.IsZero() ||
		(value.SetupState == "completed") != !value.SetupCompletedAt.IsZero() ||
		(value.ApprovalState == "not_required" ||
			value.ApprovalState == "not_submitted") &&
			(!value.SubmittedAt.IsZero() || !value.DecisionAt.IsZero()) ||
		value.ApprovalState == "pending" && !value.DecisionAt.IsZero() ||
		portalSetupTimeAfter(value.SetupCompletedAt, value.UpdatedAt) ||
		portalSetupTimeAfter(value.SubmittedAt, value.UpdatedAt) ||
		portalSetupTimeAfter(value.DecisionAt, value.UpdatedAt) ||
		(!value.SubmittedAt.IsZero() && !value.DecisionAt.IsZero() &&
			value.DecisionAt.Before(value.SubmittedAt)) {
		return ErrInvalid
	}
	return nil
}

func validPortalSetupState(value string) bool {
	return value == "not_required" || value == "required" || value == "completed"
}

func validPortalApprovalState(value string) bool {
	switch value {
	case "not_required", "not_submitted", "pending",
		"revision_required", "approved", "rejected":
		return true
	default:
		return false
	}
}

func portalSetupTimeAfter(value, upperBound time.Time) bool {
	return !value.IsZero() && value.After(upperBound)
}

func (service *Service) ResolvePortalSetup(
	ctx context.Context, scope platformtx.TenantScope, query PortalSetupQuery,
) (PortalSetup, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.BindingID == uuid.Nil || query.ActorID != scope.ActorID() ||
		query.PermissionID == uuid.Nil || query.ExpectedTokenVersion < 1 ||
		query.ExpectedAuthorizationVersion < 1 {
		return PortalSetup{}, ErrInvalid
	}
	result, err := service.store.ResolvePortalSetup(ctx, scope, query)
	if err != nil {
		return PortalSetup{}, err
	}
	if result.Validate() != nil || result.BindingID != query.BindingID {
		return PortalSetup{}, ErrUnavailable
	}
	return result, nil
}
