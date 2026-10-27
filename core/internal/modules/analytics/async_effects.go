package analytics

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type analyticsAsyncWriter interface {
	EnqueueEvent(context.Context, eventsv1.Envelope) (asyncpostgres.EnqueueResult, error)
	EnqueueJob(context.Context, jobsv1.Enqueue) (asyncpostgres.EnqueueResult, error)
}

type AsyncEffects struct {
	catalogs platformasync.WriteCatalogs
	ids      httpPlatformValues
}

func NewAsyncEffects(catalogs platformasync.WriteCatalogs) (*AsyncEffects, error) {
	jobs, producers, err := catalogs.Snapshot()
	if err != nil || !analyticsAsyncPoliciesValid(jobs, producers) {
		return nil, fmt.Errorf("Analytics async policies are invalid")
	}
	return &AsyncEffects{catalogs: catalogs, ids: httpPlatformValues{}}, nil
}

func (effects *AsyncEffects) bind(
	scope platformtx.TenantScope,
) (analyticsAsyncWriter, platformtx.OwnerAuthority, error) {
	if effects == nil {
		return nil, platformtx.OwnerAuthority{}, fmt.Errorf("Analytics effects are unavailable")
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "analytics")
	if err != nil {
		return nil, platformtx.OwnerAuthority{}, err
	}
	writer, err := asyncpostgres.Bind(scope, authority, effects.catalogs)
	if err != nil {
		return nil, platformtx.OwnerAuthority{}, err
	}
	return writer, authority, nil
}

func analyticsAsyncPoliciesValid(
	jobs jobsv1.Catalog, producers eventsv1.ProducerCatalog,
) bool {
	for _, kind := range []string{
		"analytics.query.revalidate", "analytics.cache.warm",
		"analytics.publication.repair",
	} {
		policy, err := jobs.Resolve(kind, 1)
		if err != nil || policy.OwnerDomain != "analytics" ||
			policy.EffectClass != jobsv1.EffectDatabaseTransactional {
			return false
		}
	}
	for _, eventType := range []string{
		"analytics.query.published", "analytics.dashboard.published",
		"analytics.publication.activated", "analytics.publication.revoked",
	} {
		policy, err := producers.Resolve(eventType, 1)
		if err != nil || policy.OwnerDomain != "analytics" {
			return false
		}
	}
	return true
}
