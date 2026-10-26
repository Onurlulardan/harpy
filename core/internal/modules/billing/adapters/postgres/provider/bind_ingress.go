package provider

import (
	"context"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type IngressAcceptance func(
	context.Context,
	providerapp.WebhookAcceptance,
	JobWriter,
) (providerapp.WebhookAcceptDecision, error)

type ingressRepository struct{ database ingressDatabase }

func BindIngress(
	scope platformtx.ProviderTenantScope,
) (IngressAcceptance, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *ingressRepository {
			return &ingressRepository{database: ingressDatabase{
				queryRow: func(
					ctx context.Context, statement string, arguments ...any,
				) databaseRow {
					return database.QueryRow(ctx, statement, arguments...)
				},
				execute: func(
					ctx context.Context, statement string, arguments ...any,
				) (int64, error) {
					result, err := database.Execute(ctx, statement, arguments...)
					return result.RowsAffected(), err
				},
				notFound: platformpostgres.RowNotFound,
			}}
		},
	)
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	bound, err := platformtx.BindProvider(scope, binder)
	if err != nil || bound == nil {
		return nil, providerapp.ErrWebhookDecision
	}
	target := providerTarget{
		tenantID: scope.TenantID(), mappingID: scope.MappingID(),
		providerAccountID: scope.ProviderAccountID(),
	}
	return func(
		ctx context.Context,
		acceptance providerapp.WebhookAcceptance,
		jobs JobWriter,
	) (providerapp.WebhookAcceptDecision, error) {
		if !platformtx.ContextOwnsProviderScope(ctx, scope) ||
			platformtx.NetworkAllowed(ctx) {
			return providerapp.WebhookAcceptDecision{}, providerapp.ErrWebhookDecision
		}
		return bound.accept(ctx, target, acceptance, jobs)
	}, nil
}
