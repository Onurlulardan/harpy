package billing

import (
	"context"
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	invoicepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/invoice"
	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const invoiceFinalizedSchemaSHA256 = "b55f45f8a7321fd9a94a5436782eb8a116b75f5cbfb72d9dbc4d735c51eb5264"

type invoiceEffects struct {
	tenantID  uuid.UUID
	authority platformtx.OwnerAuthority
	producers eventsv1.ProducerCatalog
	writer    asyncpostgres.Writer
}

func newInvoiceEffects(
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
	catalogs platformasync.WriteCatalogs,
) (*invoiceEffects, error) {
	_, producers, err := catalogs.Snapshot()
	if err != nil || !authority.OwnsScope(scope) || authority.Domain() != "billing" {
		return nil, invoiceapp.ErrConfiguration
	}
	writer, err := asyncpostgres.Bind(scope, authority, catalogs)
	if err != nil {
		return nil, fmt.Errorf("bind billing invoice async writer: %w", err)
	}
	return &invoiceEffects{
		tenantID: scope.TenantID(), authority: authority,
		producers: producers, writer: writer,
	}, nil
}

func (effects *invoiceEffects) PublishFinalized(
	ctx context.Context,
	event invoiceapp.FinalizedEventIntent,
	finalization invoiceapp.Finalization,
) error {
	if effects == nil || !event.Valid() || event.InvoiceID != finalization.InvoiceID ||
		event.InvoiceRevision != finalization.ResultRevision {
		return invoiceapp.ErrConfiguration
	}
	payload, digest, err := encodeInvoiceFinalizedEvent(event)
	if err != nil {
		return err
	}
	tenantID, causationID := effects.tenantID.String(), event.CausationID.String()
	envelope, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: event.ID.String(), EventType: invoiceapp.FinalizedEventType,
		EventVersion: 1, TenantID: &tenantID,
		ProducerAuthority: billingEventAuthority(effects.authority),
		AggregateType:     "billing.invoice", AggregateID: event.InvoiceID.String(),
		AggregateVersion: event.InvoiceRevision, EventSequence: event.InvoiceRevision,
		OccurredAt: event.OccurredAt.Format(time.RFC3339Nano), ActorID: event.ActorID.String(),
		CorrelationID: causationID, CausationID: causationID,
		IdempotencyKey: event.ID.String(),
		BusinessEffectIdempotencyKey: fmt.Sprintf(
			"billing.invoice.finalized:%s:%d", event.InvoiceID, event.InvoiceRevision,
		),
		PayloadSchemaSHA256: invoiceFinalizedSchemaSHA256,
		PayloadSHA256:       digest, Payload: payload,
	}, effects.producers)
	if err != nil {
		return err
	}
	result, err := effects.writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != event.ID.String() {
		return fmt.Errorf("enqueue billing invoice event %s: %w", event.ID, err)
	}
	return nil
}

var _ invoicepostgres.Effects = (*invoiceEffects)(nil)
