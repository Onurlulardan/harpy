package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	providerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/provider"
	subscriptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/subscription"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type providerIngressStore struct {
	coordinator *platformtx.Coordinator
	resolver    platformtx.ProviderResolverBinder
	catalogs    platformasync.WriteCatalogs
}

func NewProviderIngressStore(
	coordinator *platformtx.Coordinator,
	catalogs platformasync.WriteCatalogs,
) (providerapp.WebhookIngressStore, error) {
	if coordinator == nil {
		return nil, providerapp.ErrConfiguration
	}
	jobs, _, err := catalogs.Snapshot()
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	policy := providerapp.WebhookProcessorPolicy()
	registered, err := jobs.Resolve(policy.Kind, policy.PayloadRevision)
	if err != nil || registered != policy {
		return nil, providerapp.ErrConfiguration
	}
	resolver, err := providerpostgres.NewResolverBinder()
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	return &providerIngressStore{
		coordinator: coordinator, resolver: resolver, catalogs: catalogs,
	}, nil
}

func (storage *providerIngressStore) AcceptWebhook(
	ctx context.Context,
	acceptance providerapp.WebhookAcceptance,
) (decision providerapp.WebhookAcceptDecision, err error) {
	lookup, err := providerLookup(acceptance)
	if storage == nil || storage.coordinator == nil || err != nil {
		return providerapp.WebhookAcceptDecision{}, providerapp.ErrWebhookMapping
	}
	request := platformtx.ProviderScopeRequest{Lookup: lookup, Resolver: storage.resolver}
	err = storage.coordinator.InProviderTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.ProviderTenantScope) error {
			accept, bindErr := providerpostgres.BindIngress(scope)
			if bindErr != nil {
				return bindErr
			}
			authority, bindErr := platformtx.MintProviderOwnerAuthority(scope, "billing")
			if bindErr != nil {
				return bindErr
			}
			writer, bindErr := asyncpostgres.BindProvider(
				scope, authority, storage.catalogs,
			)
			if bindErr != nil {
				return bindErr
			}
			decision, bindErr = accept(
				scoped, acceptance,
				&providerJobWriter{tenantID: scope.TenantID(), authority: authority, writer: writer},
			)
			return bindErr
		},
	)
	if errors.Is(err, platformtx.ErrProviderTarget) ||
		errors.Is(err, platformtx.ErrProviderResolver) {
		return providerapp.WebhookAcceptDecision{}, providerapp.ErrWebhookMapping
	}
	if err != nil {
		return providerapp.WebhookAcceptDecision{}, fmt.Errorf(
			"accept billing webhook: %w", err,
		)
	}
	return decision, nil
}

func providerLookup(
	acceptance providerapp.WebhookAcceptance,
) (platformtx.ProviderLookup, error) {
	verified := acceptance.Verified()
	return platformtx.NewProviderLookup(
		string(acceptance.Provider()), verified.AccountID(), verified.ObjectKind(),
		verified.ObjectID(), verified.EventID(), verified.AuthenticationSHA256(),
	)
}

type providerJobWriter struct {
	tenantID  uuid.UUID
	authority platformtx.OwnerAuthority
	writer    asyncpostgres.Writer
}

func (writer *providerJobWriter) EnqueueProviderJob(
	ctx context.Context,
	jobID uuid.UUID,
	job providerapp.AsyncJob,
) error {
	if writer == nil || writer.tenantID == uuid.Nil || writer.writer == nil ||
		jobID == uuid.Nil || job.Policy().OwnerDomain != "billing" {
		return providerapp.ErrWebhookDecision
	}
	tenantID := writer.tenantID.String()
	value, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: jobID.String(), Kind: job.Policy().Kind,
		PayloadRevision: job.Policy().PayloadRevision, TenantID: &tenantID,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenantID,
			Domain:            writer.authority.Domain(),
			CutoverGeneration: writer.authority.Generation(),
			WriterOrigin:      jobsv1.WriterV2, GroupID: writer.authority.GroupID(),
			GroupGeneration: writer.authority.GroupGeneration(),
		},
		PayloadReference: job.PayloadReference(), PayloadSHA256: job.PayloadSHA256(),
		RunAt: job.RunAt().Format(time.RFC3339Nano), Priority: 100,
		DedupeKey: job.DedupeKey(), EffectIdempotencyKey: optionalProviderText(job.EffectKey()),
		MaximumAttempts: job.Policy().MaximumAttempts,
	}, singleProviderJobCatalog(job.Policy()))
	if err != nil {
		return fmt.Errorf("validate billing provider job: %w", err)
	}
	result, err := writer.writer.EnqueueJob(ctx, value)
	if err != nil {
		return fmt.Errorf("enqueue billing provider job %s: %w", jobID, err)
	}
	if result.ID != jobID.String() {
		return providerapp.ErrWebhookConflict
	}
	return nil
}

func BindProviderWebhookProcessor(
	scope platformtx.JobTenantScope,
	effects subscriptionpostgres.Effects,
	nextID subscriptionapp.IDGenerator,
	clock subscriptionapp.Clock,
	jobs providerpostgres.JobWriter,
) (providerapp.WebhookProcessorStore, error) {
	subscriptions, err := subscriptionpostgres.BindProvider(scope, effects)
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	commands, err := subscriptionapp.NewCommands(subscriptions, nextID, clock)
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	return providerpostgres.BindProcessor(scope, nextID, clock, jobs, commands)
}

func optionalProviderText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func singleProviderJobCatalog(policy jobsv1.KindPolicy) jobsv1.Catalog {
	catalog, _ := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	return catalog
}
