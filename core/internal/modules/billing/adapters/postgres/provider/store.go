package provider

import (
	"context"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type databaseRow interface{ Scan(...any) error }

type ingressDatabase struct {
	queryRow func(context.Context, string, ...any) databaseRow
	execute  func(context.Context, string, ...any) (int64, error)
	notFound func(error) bool
}

func (database ingressDatabase) usable() bool {
	return database.queryRow != nil && database.execute != nil && database.notFound != nil
}

type providerTarget struct {
	tenantID          uuid.UUID
	mappingID         uuid.UUID
	providerAccountID uuid.UUID
}

type databaseAccess struct {
	queryRow func(context.Context, string, ...any) databaseRow
	execute  func(context.Context, string, ...any) error
	allowed  func(context.Context) bool
	notFound func(error) bool
}

type JobWriter interface {
	EnqueueProviderJob(context.Context, uuid.UUID, providerapp.AsyncJob) error
}

type processorStore struct {
	tenantID uuid.UUID
	database databaseAccess
	nextID   subscriptionapp.IDGenerator
	clock    subscriptionapp.Clock
	jobs     JobWriter
	update   SubscriptionUpdater
}

type SubscriptionUpdater interface {
	UpdateFromProvider(
		context.Context,
		subscriptionapp.UpdateFromProviderRequest,
	) (subscriptionapp.Result, error)
}

var _ providerapp.WebhookProcessorStore = (*processorStore)(nil)

func (storage *processorStore) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil &&
		storage.database.queryRow != nil && storage.database.execute != nil &&
		storage.database.allowed != nil && storage.database.notFound != nil &&
		storage.nextID != nil && storage.clock != nil && storage.jobs != nil &&
		storage.update != nil && ctx != nil && ctx.Err() == nil &&
		storage.database.allowed(ctx)
}
