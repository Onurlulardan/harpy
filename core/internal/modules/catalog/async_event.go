package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (port *AsyncPort) Emit(
	ctx context.Context, scope platformtx.TenantScope, event catalogapp.InstallationEvent,
) error {
	if event.Version != "1" || event.Installation.RowVersion < 1 ||
		event.Installation.RowVersion > math.MaxInt64 || !validAsyncDigest(event.PlanSHA256) ||
		!canonicalAsyncTime(event.OccurredAt) || event.IdempotencyKey == "" {
		return catalogapp.ErrInvalid
	}
	writer, authority, err := port.bindAsync(ctx, scope)
	if err != nil {
		return err
	}
	payload, err := catalogEventPayload(event)
	if err != nil {
		return catalogapp.ErrInvalid
	}
	policy, err := port.producers.Resolve(event.Type, 1)
	if err != nil {
		return catalogapp.ErrInvalid
	}
	digest := sha256.Sum256(payload)
	operationID, err := uuid.Parse(event.OperationID.String())
	if err != nil || operationID.Version() != 7 {
		return catalogapp.ErrInvalid
	}
	eventID := catalogEventID(operationID, event.Type == asyncmanifestv1.CatalogMenuPublishedEventType)
	aggregateType, aggregateID := "catalog_installation", event.Installation.ID.String()
	aggregateVersion := int64(event.Installation.RowVersion)
	eventSequence := int64(event.Installation.RowVersion)
	if event.Type == asyncmanifestv1.CatalogMenuPublishedEventType {
		aggregateType, aggregateID, aggregateVersion, eventSequence = "catalog_menu", operationID.String(), 1, 1
	}
	tenantID := scope.TenantID().String()
	envelope, err := eventsv1.NewEnvelope(eventsv1.Envelope{EventID: eventID.String(),
		EventType: event.Type, EventVersion: 1, TenantID: &tenantID,
		ProducerAuthority: catalogEventAuthority(authority), AggregateType: aggregateType,
		AggregateID: aggregateID, AggregateVersion: aggregateVersion, EventSequence: eventSequence,
		OccurredAt: event.OccurredAt.Format(time.RFC3339Nano), ActorID: scope.ActorID().String(),
		CorrelationID: operationID.String(), CausationID: operationID.String(),
		IdempotencyKey: event.IdempotencyKey, BusinessEffectIdempotencyKey: event.IdempotencyKey,
		PayloadSchemaSHA256: policy.PayloadSchemaSHA256,
		PayloadSHA256:       hex.EncodeToString(digest[:]), Payload: payload}, port.producers)
	if err != nil {
		return asyncUnavailable()
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != eventID.String() {
		return asyncUnavailable()
	}
	return nil
}

func catalogEventPayload(event catalogapp.InstallationEvent) ([]byte, error) {
	base := struct{ PackageID, InstallationID, OperationID string }{
		event.Installation.PackageID.String(), event.Installation.ID.String(), event.OperationID.String(),
	}
	if event.Type == asyncmanifestv1.CatalogMenuPublishedEventType {
		return json.Marshal(struct {
			PackageID      string `json:"package_id"`
			InstallationID string `json:"installation_id"`
			OperationID    string `json:"operation_id"`
			MenuVersionID  string `json:"menu_version_id"`
			ContentSHA256  string `json:"content_sha256"`
		}{base.PackageID, base.InstallationID, base.OperationID, base.OperationID, event.PlanSHA256})
	}
	return json.Marshal(map[string]any{"package_id": base.PackageID,
		"installation_id": base.InstallationID, "operation_id": base.OperationID,
		"package_version": event.Installation.ActiveVersion.String(),
		"state":           string(event.Installation.State), "reference_epoch": event.Installation.ReferenceEpoch,
		"plan_sha256": event.PlanSHA256})
}

func catalogEventID(operationID uuid.UUID, menu bool) uuid.UUID {
	if !menu {
		return operationID
	}
	result := operationID
	result[15] ^= 1
	return result
}

func validAsyncDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && hex.EncodeToString(decoded) == value
}

func canonicalAsyncTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
