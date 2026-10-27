package iam

import (
	"context"
	"time"

	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type profileRecoverySource interface {
	InspectParent(
		context.Context, platformtx.TenantScope, uuid.UUID,
		profilebackfillapp.Evidence,
	) (profilebackfillpostgres.RecoveryContract, error)
	OpenInventory(
		context.Context, platformtx.TenantScope, uuid.UUID,
		profilebackfillapp.Evidence,
	) (profilebackfillpostgres.RecoveryInventoryCursor, error)
	LookupReceipt(
		context.Context, platformtx.TenantScope, profilebackfillapp.Evidence,
	) (profilebackfillapp.RecoveryReceipt, bool, error)
}

type profileRecoveryStarter interface {
	Start(
		context.Context, platformtx.TenantScope,
		profilebackfillapp.RecoveryCreateCommand,
	) (profilebackfillapp.RecoveryResult, error)
}

type ProfileBackfillRecoveryRuntime struct {
	source  profileRecoverySource
	starter profileRecoveryStarter
	clock   func() time.Time
}

type ProfileBackfillRecoveryResult struct {
	PlanID             uuid.UUID
	RecoveryOfPlanID   uuid.UUID
	RecoveryRootPlanID uuid.UUID
	RecoverySequence   int64
	Revision           int64
	AuthorizationClock int64
	State              profilebackfillapp.RecoveryState
	Replayed           bool
}

func NewProfileBackfillRecoveryRuntime(
	catalogs platformasync.WriteCatalogs,
	clock func() time.Time,
) (*ProfileBackfillRecoveryRuntime, error) {
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, profilebackfillapp.ErrEffects
	}
	source, err := profilebackfillpostgres.NewRecoveryReader()
	if err != nil {
		return nil, err
	}
	store, err := profilebackfillpostgres.NewRecoveryStore()
	if err != nil {
		return nil, err
	}
	starter, err := profilebackfillapp.NewRecoveryService(
		store, profileBackfillEffects{catalogs: catalogs},
	)
	if err != nil {
		return nil, err
	}
	return newProfileBackfillRecoveryRuntime(source, starter, clock)
}

func newProfileBackfillRecoveryRuntime(
	source profileRecoverySource,
	starter profileRecoveryStarter,
	clock func() time.Time,
) (*ProfileBackfillRecoveryRuntime, error) {
	if nilProfileHTTPDependency(source) || nilProfileHTTPDependency(starter) ||
		clock == nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	return &ProfileBackfillRecoveryRuntime{
		source: source, starter: starter, clock: clock,
	}, nil
}

func recoveryResultFromReceipt(
	receipt profilebackfillapp.RecoveryReceipt,
) ProfileBackfillRecoveryResult {
	return ProfileBackfillRecoveryResult{
		PlanID: receipt.PlanID, RecoveryOfPlanID: receipt.RecoveryOfPlanID,
		RecoveryRootPlanID: receipt.RecoveryRootPlanID,
		RecoverySequence:   receipt.RecoverySequence, Revision: receipt.Revision,
		AuthorizationClock: receipt.AuthorizationClock, State: receipt.State,
		Replayed: true,
	}
}

func recoveryResultFromPlan(
	plan profilebackfillapp.RecoveryPlan,
	result profilebackfillapp.RecoveryResult,
) ProfileBackfillRecoveryResult {
	return ProfileBackfillRecoveryResult{
		PlanID: result.PlanID, RecoveryOfPlanID: plan.RecoveryOfPlanID(),
		RecoveryRootPlanID: plan.RecoveryRootPlanID(),
		RecoverySequence:   plan.RecoverySequence(), Revision: result.Revision,
		AuthorizationClock: result.AuthorizationClock, State: result.State,
		Replayed: result.Replayed,
	}
}
