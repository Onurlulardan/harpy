//go:build integration

package subscription

import (
	"context"
	"crypto/sha256"
	"database/sql"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type subscriptionActorEffects struct {
	transaction *sql.Tx
	tenantID    uuid.UUID
}

func (effects *subscriptionActorEffects) EnqueueBoundary(
	ctx context.Context,
	intent subscriptionapp.JobIntent,
) (BoundaryArtifact, error) {
	reference := "urn:test:billing:subscription-boundary:" + intent.ID.String()
	digest := sha256.Sum256([]byte(reference))
	_, err := effects.transaction.ExecContext(
		ctx, `
INSERT INTO platform.jobs(
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,
 payload_revision,payload_ref,payload_sha256,state
) VALUES($1,$2,'tenant',$2,'billing_forms',1,'billing',1,$3,$4,
 pg_catalog.encode($5,'hex'),'queued')
`, intent.ID, effects.tenantID, intent.PayloadRevision, reference, digest[:],
	)
	return BoundaryArtifact{PayloadReference: reference, PayloadSHA256: digest}, err
}

func (effects *subscriptionActorEffects) CancelBoundary(
	context.Context,
	subscriptionapp.JobCancellation,
) error {
	return nil
}

func (effects *subscriptionActorEffects) PublishSubscription(
	ctx context.Context,
	intent subscriptionapp.OutboxIntent,
	_ subscriptionapp.SubscriptionSnapshot,
) error {
	_, err := effects.transaction.ExecContext(
		ctx, `INSERT INTO platform.test_subscription_events(event_id) VALUES($1)`,
		intent.ID,
	)
	return err
}

var _ Effects = (*subscriptionActorEffects)(nil)
