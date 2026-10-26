package provider

import (
	"bytes"
	"context"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func (storage *processorStore) ApplyWebhook(
	ctx context.Context,
	application providerapp.WebhookApplication,
) (providerapp.WebhookProcessDecision, error) {
	if !storage.usable(ctx) || application.ReceiptID() == uuid.Nil ||
		application.TenantID() != storage.tenantID {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrWebhookProcessing
	}
	row, err := storage.loadWebhookRow(ctx, application.ReceiptID())
	if err != nil || !row.matches(application) {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrWebhookMapping
	}
	event := application.Event()
	canonical, err := canonicaljsonv1.CanonicalizeObject(event.CanonicalFact(), 64<<10)
	if err != nil || !bytes.Equal(canonical, event.CanonicalFact()) {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrNormalizedEvent
	}
	existing, found, err := storage.findProviderEvent(
		ctx, row.authorityID, application,
	)
	if err != nil {
		return providerapp.WebhookProcessDecision{}, err
	}
	if found {
		if !existing.exact(event) ||
			(existing.receiptState != "applied" &&
				existing.receiptState != "reconcile_required") {
			return providerapp.WebhookProcessDecision{}, providerapp.ErrWebhookConflict
		}
		return providerapp.WebhookProcessDecision{State: providerapp.WebhookDuplicate}, nil
	}
	if row.state.value != "processing" {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrWebhookProcessing
	}
	eventID, normalizedAt, err := storage.eventIdentity(event)
	if err != nil {
		return providerapp.WebhookProcessDecision{}, err
	}
	ordered, err := storage.orderingReady(ctx, row, event)
	if err != nil {
		return providerapp.WebhookProcessDecision{}, err
	}
	if err := storage.insertProviderEvent(
		ctx, row, application, eventID, normalizedAt,
	); err != nil {
		return providerapp.WebhookProcessDecision{}, err
	}
	if !ordered || event.Schema() != subscriptionSnapshotSchema ||
		event.Revision() != subscriptionSnapshotRevision {
		return storage.queueReconciliation(ctx, row, application, normalizedAt)
	}
	return storage.applySubscriptionFact(
		ctx, row, application, eventID, normalizedAt,
	)
}

func (storage *processorStore) eventIdentity(
	event providerapp.NormalizedEvent,
) (uuid.UUID, time.Time, error) {
	eventID, err := storage.nextID()
	if err != nil || eventID == uuid.Nil || eventID.Version() != uuid.Version(7) {
		return uuid.Nil, time.Time{}, providerapp.ErrProviderDependency
	}
	now := storage.clock()
	if now.IsZero() || now.Location() != time.UTC ||
		now.Nanosecond()%int(time.Microsecond) != 0 || now.Before(event.OccurredAt()) {
		return uuid.Nil, time.Time{}, providerapp.ErrProviderDependency
	}
	return eventID, now, nil
}

func (row webhookRow) matches(application providerapp.WebhookApplication) bool {
	verified := application.Verified()
	return row.authorityID != uuid.Nil && row.mappingID != uuid.Nil &&
		row.providerAccountID != uuid.Nil && row.providerKey == string(application.Provider()) &&
		row.accountID == verified.AccountID() && row.objectKind == verified.ObjectKind() &&
		row.objectID == verified.ObjectID() && row.eventID == verified.EventID()
}
