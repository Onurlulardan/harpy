package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	subscriptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/subscription"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const subscriptionBoundaryJobKind = "billing.subscription.boundary"
const subscriptionChangeSchemaSHA256 = "fc511ef78cedb23785e362d889271374bc30f9446e221b4fdb742130ecba9dc3"
const entitlementChangeSchemaSHA256 = "ae7a1c1cb43f0d58d7cf292f5d92e099ba3534eac709fa54da6c4f6bf6e2403e"

type subscriptionEffects struct {
	tenantID  uuid.UUID
	authority platformtx.OwnerAuthority
	jobs      jobsv1.Catalog
	producers eventsv1.ProducerCatalog
	writer    asyncpostgres.Writer
}

func newSubscriptionEffects(
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
	catalogs platformasync.WriteCatalogs,
) (*subscriptionEffects, error) {
	jobs, producers, err := catalogs.Snapshot()
	if err != nil || !authority.OwnsScope(scope) || authority.Domain() != "billing" {
		return nil, subscriptionapp.ErrConfiguration
	}
	writer, err := asyncpostgres.Bind(scope, authority, catalogs)
	if err != nil {
		return nil, fmt.Errorf("bind billing subscription async writer: %w", err)
	}
	return &subscriptionEffects{
		tenantID: scope.TenantID(), authority: authority,
		jobs: jobs, producers: producers, writer: writer,
	}, nil
}

func (effects *subscriptionEffects) EnqueueBoundary(
	ctx context.Context,
	intent subscriptionapp.JobIntent,
) (subscriptionpostgres.BoundaryArtifact, error) {
	artifact, err := subscriptionBoundaryArtifact(effects.tenantID, intent)
	if err != nil {
		return subscriptionpostgres.BoundaryArtifact{}, err
	}
	policy, err := effects.jobs.Resolve(subscriptionBoundaryJobKind, intent.PayloadRevision)
	if err != nil {
		return subscriptionpostgres.BoundaryArtifact{}, err
	}
	tenantID := effects.tenantID.String()
	job, err := jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: intent.ID.String(), Kind: subscriptionBoundaryJobKind,
		PayloadRevision: intent.PayloadRevision, TenantID: &tenantID,
		Authority:        subscriptionJobAuthority(effects.authority),
		PayloadReference: artifact.PayloadReference,
		PayloadSHA256:    hex.EncodeToString(artifact.PayloadSHA256[:]),
		RunAt:            intent.RunAt.Format(time.RFC3339Nano), Priority: 100,
		DedupeKey:       "billing.subscription.boundary:" + intent.EffectID.String(),
		MaximumAttempts: policy.MaximumAttempts,
	}, effects.jobs)
	if err != nil {
		return subscriptionpostgres.BoundaryArtifact{}, err
	}
	result, err := effects.writer.EnqueueJob(ctx, job)
	if err != nil || result.ID != intent.ID.String() {
		return subscriptionpostgres.BoundaryArtifact{}, fmt.Errorf(
			"enqueue billing subscription boundary %s: %w", intent.ID, err,
		)
	}
	return artifact, nil
}

func (effects *subscriptionEffects) CancelBoundary(
	ctx context.Context,
	cancellation subscriptionapp.JobCancellation,
) error {
	digest := sha256.Sum256([]byte(cancellation.Reason))
	request, err := asyncpostgres.NewOwnedJobCancellation(
		cancellation.ID, cancellation.CancelledAt,
		"BILLING_BOUNDARY_CANCELLED", hex.EncodeToString(digest[:]),
	)
	if err != nil {
		return err
	}
	_, err = effects.writer.CancelOwnedJob(ctx, request)
	return err
}

func (effects *subscriptionEffects) PublishSubscription(
	ctx context.Context,
	intent subscriptionapp.OutboxIntent,
	snapshot subscriptionapp.SubscriptionSnapshot,
) error {
	envelope, err := effects.subscriptionEnvelope(intent, snapshot)
	if err != nil {
		return err
	}
	result, err := effects.writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != intent.ID.String() {
		return fmt.Errorf("enqueue billing subscription event %s: %w", intent.ID, err)
	}
	return nil
}

var _ subscriptionpostgres.Effects = (*subscriptionEffects)(nil)
