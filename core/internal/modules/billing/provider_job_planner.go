package billing

import (
	"context"
	"fmt"
	"strings"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var providerEffectNamespace = uuid.MustParse("9463be3b-8245-57a1-b80f-781556074d11")

type providerJobPlanner struct {
	kind            string
	referencePrefix string
	effectPrefix    string
	operation       string
}

func (planner providerJobPlanner) Plan(
	_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	tenantID, tenantScoped := claim.TenantID()
	provider, causationID, err := parseProviderJobReference(
		claim.PayloadReference(), planner.referencePrefix,
	)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	if err != nil || !tenantScoped || tenantID == uuid.Nil || claim.Kind() != planner.kind ||
		claim.PayloadRevision() != 1 || !hasEffectKey ||
		!strings.HasPrefix(effectKey, planner.effectPrefix) ||
		!validProviderEffectKey(planner.kind, provider, causationID, effectKey) {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("billing provider job claim is invalid")
	}
	attemptID := claim.AttemptID()
	if attemptID == uuid.Nil {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("billing provider job claim is invalid")
	}
	effectID := uuid.NewSHA1(providerEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(
		effectID, provider, planner.operation, claim.JobID().String(), causationID,
	)
}

func providerCommandPlanner() providerJobPlanner {
	return providerJobPlanner{
		kind:            providerapp.ProviderCommandJobKind,
		referencePrefix: "billing-provider-command",
		effectPrefix:    "billing:provider:command:",
		operation:       providerapp.CommandOperation,
	}
}

func providerReconciliationPlanner() providerJobPlanner {
	return providerJobPlanner{
		kind:            providerapp.ProviderReconciliationJobKind,
		referencePrefix: "billing-provider-reconciliation",
		effectPrefix:    "billing:provider:reconcile:",
		operation:       providerapp.ReconciliationOperation,
	}
}
