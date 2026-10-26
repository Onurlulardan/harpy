package iam

import (
	"context"

	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *ProfileBackfillRecoveryRuntime) plan(
	ctx context.Context,
	scope platformtx.TenantScope,
	contract profilebackfillpostgres.RecoveryContract,
	evidence profilebackfillapp.Evidence,
) (profilebackfillapp.RecoveryPlan, error) {
	classifier, err := NewProfileRecoveryInventoryClassifier(
		ProfileRecoveryInventoryInput{
			Parent: contract.Parent, Source: contract.Source, Target: contract.Target,
			Normalizer: profileHTTPNormalizer{}, Clock: runtime.clock,
		},
	)
	if err != nil {
		return profilebackfillapp.RecoveryPlan{}, err
	}
	inventory, err := runtime.scanRecoveryInventory(
		ctx, scope, contract, evidence, classifier,
	)
	if err != nil {
		return profilebackfillapp.RecoveryPlan{}, err
	}
	runAt, err := runtime.recoveryRunAt()
	if err != nil {
		return profilebackfillapp.RecoveryPlan{}, err
	}
	planID := uuid.NewSHA1(
		evidence.OperationID, []byte("iam.profile.backfill.recovery.plan.v1"),
	)
	return profilebackfillapp.PlanRecovery(
		planID, contract.Parent, inventory,
		profilebackfillapp.MaximumPageSize, runAt,
	)
}
