package provider

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type mappingResolver struct {
	database platformpostgres.BoundExecutor
}

func NewResolverBinder() (platformtx.ProviderResolverBinder, error) {
	return platformtx.NewProviderResolverBinder(
		func(database platformpostgres.BoundExecutor) platformtx.ProviderTargetResolver {
			return &mappingResolver{database: database}
		},
	)
}

func (resolver *mappingResolver) ProviderTarget(
	ctx context.Context,
	lookup platformtx.ProviderLookup,
) (platformtx.ProviderTargetEvidence, error) {
	if resolver == nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || platformtx.ProviderLookupBinding(lookup) == [32]byte{} {
		return platformtx.ProviderTargetEvidence{}, fmt.Errorf("provider mapping resolver is unavailable")
	}
	var evidence platformtx.ProviderTargetEvidence
	var subscriptionID, invoiceID *uuid.UUID
	err := resolver.database.QueryRow(
		ctx, resolveProviderTargetSQL, lookup.Provider(), lookup.AccountID(),
		lookup.ObjectKind(), lookup.ObjectID(),
	).Scan(
		&evidence.TenantID, &evidence.MappingID, &evidence.ProviderAccountID,
		&subscriptionID, &invoiceID, &evidence.MappingRevision,
		&evidence.GroupGeneration, &evidence.BillingGeneration,
		&evidence.FormsGeneration,
	)
	if err != nil {
		return platformtx.ProviderTargetEvidence{}, fmt.Errorf("resolve provider mapping: %w", err)
	}
	if subscriptionID != nil {
		evidence.SubscriptionID = *subscriptionID
	}
	if invoiceID != nil {
		evidence.InvoiceID = *invoiceID
	}
	evidence.LookupBindingSHA256 = platformtx.ProviderLookupBinding(lookup)
	return evidence, nil
}

var _ platformtx.ProviderTargetResolver = (*mappingResolver)(nil)
