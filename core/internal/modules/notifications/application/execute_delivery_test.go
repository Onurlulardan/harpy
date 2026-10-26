package application

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestExecuteDeliveryAcceptsAndClearsSensitiveBuffers(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	result := ports.ProviderResult{Outcome: ports.ProviderAccepted, Code: "ACCEPTED",
		ResponseClass: "smtp_250", StatusCode: "250", CompletedAt: now,
		ProviderReferenceDigest: sha256.Sum256([]byte("provider-message"))}
	worker, store, sender, tenantID, deliveryID := workerFixture(t, result)
	processed, err := worker.Execute(t.Context(), tenantID, deliveryID, "worker-a")
	if err != nil || processed.Outcome != ports.ProviderAccepted ||
		store.finalized.Delivery.State != domain.DeliveryAccepted ||
		store.finalized.Attempt.ProviderReferenceDigest == [32]byte{} ||
		store.finalized.ProviderLocator == nil {
		t.Fatalf("Execute() = %+v, %v; final = %+v", processed, err, store.finalized)
	}
	if !zeroed(sender.target) || !zeroed(sender.credential) {
		t.Fatal("recipient or credential buffer was not cleared")
	}
}

func TestExecuteDeliveryRequeuesDefinitiveTransient(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	worker, store, _, tenantID, deliveryID := workerFixture(t,
		ports.ProviderResult{Outcome: ports.ProviderTransient, Code: "CONNECT_FAILED",
			ResponseClass: "pre_bytes", CompletedAt: now})
	processed, err := worker.Execute(t.Context(), tenantID, deliveryID, "worker-a")
	if err != nil || processed.Outcome != ports.ProviderTransient ||
		store.finalized.Delivery.State != domain.DeliveryPending ||
		!store.finalized.Delivery.NotBefore.After(now) {
		t.Fatalf("Execute() = %+v, %v; final = %+v", processed, err, store.finalized)
	}
}

func TestExecuteDeliverySuppressesDisabledProviderBeforeEffects(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	worker, store, sender, tenantID, deliveryID := workerFixture(t,
		ports.ProviderResult{Outcome: ports.ProviderAccepted, Code: "ACCEPTED",
			ResponseClass: "smtp_250", CompletedAt: now,
			ProviderReferenceDigest: sha256.Sum256([]byte("provider-message"))})
	store.work.Provider.Enabled = false
	processed, err := worker.Execute(t.Context(), tenantID, deliveryID, "worker-a")
	if err != nil || processed.Outcome != ports.ProviderPermanent ||
		processed.Code != "PROVIDER_DISABLED" ||
		store.finalized.Delivery.State != domain.DeliveryFailed ||
		store.finalized.Attempt.ErrorCode != "PROVIDER_DISABLED" {
		t.Fatalf("Execute() = %+v, %v; final = %+v", processed, err, store.finalized)
	}
	if sender.target != nil || sender.credential != nil {
		t.Fatal("disabled provider performed an external effect")
	}
}

func TestExecuteDeliveryMakesInvalidOrUnpersistedResultUnknown(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	worker, store, _, tenantID, deliveryID := workerFixture(t,
		ports.ProviderResult{Outcome: ports.ProviderAccepted, Code: "ACCEPTED",
			ResponseClass: "smtp_250", CompletedAt: now})
	processed, err := worker.Execute(t.Context(), tenantID, deliveryID, "worker-a")
	if !errors.Is(err, ErrOutcomeUnknown) || processed.Outcome != ports.ProviderUnknown ||
		!store.finalized.Delivery.OutcomeUnknown {
		t.Fatalf("invalid Execute() = %+v, %v; final = %+v", processed, err, store.finalized)
	}
	worker, store, _, tenantID, deliveryID = workerFixture(t,
		ports.ProviderResult{Outcome: ports.ProviderPermanent, Code: "AUTH_REJECTED",
			ResponseClass: "smtp_535", CompletedAt: now})
	store.finalizeErr = errFinalizeFake
	_, err = worker.Execute(t.Context(), tenantID, deliveryID, "worker-a")
	if !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("unpersisted result error = %v", err)
	}
}
