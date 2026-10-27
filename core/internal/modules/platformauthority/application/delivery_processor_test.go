package application

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type deliveryFinalizer struct {
	calls    int
	failOnce bool
}

func (finalizer *deliveryFinalizer) FinalizeAuthorityOperation(
	context.Context,
	ports.OperationRecord,
) error {
	finalizer.calls++
	if finalizer.failOnce {
		finalizer.failOnce = false
		return fmt.Errorf("owner transaction interrupted")
	}
	return nil
}

func TestDeliveryProcessorResumesAfterOwnerFinalizationCrash(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	initial, err := fixture.service.PrepareInitial(
		context.Background(),
		api.InitialAuthorityRequest{
			OperationID: testOperationID,
			State:       fixture.state,
			IssuedAt:    testNow,
			Policy:      api.ExpiryOnline,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fixture.service.AcknowledgePrepared(
		context.Background(), initial.OperationID,
	); err != nil {
		t.Fatal(err)
	}
	changed := fixture.state
	anchorDigest := strings.Repeat("b", 64)
	changed.RecoveryAnchorSetSHA256 = &anchorDigest
	const changeOperationID = "55555555-5555-4555-8555-555555555555"
	if _, err = fixture.service.PrepareChange(
		context.Background(),
		api.ProtectedChangeRequest{
			OperationID: changeOperationID,
			Purpose:     api.ChangeRecoveryAnchor,
			State:       changed,
			IssuedAt:    testNow,
			Policy:      api.ExpiryOnline,
		},
	); err != nil {
		t.Fatal(err)
	}
	finalizer := &deliveryFinalizer{failOnce: true}
	processor, err := NewDeliveryProcessor(
		fixture.store, fixture.registry, testClock{},
		fixture.service.trust, fixture.service.registryKind, finalizer, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err = processor.Process(
		context.Background(), changeOperationID,
	); err == nil {
		t.Fatal("owner finalization crash was hidden")
	}
	if err = processor.Process(
		context.Background(), changeOperationID,
	); err != nil {
		t.Fatal(err)
	}
	operation, err := fixture.store.LoadOperation(
		context.Background(), changeOperationID,
	)
	if err != nil || operation.Status != ports.StatusCompleted {
		t.Fatalf("operation status = %s, err=%v", operation.Status, err)
	}
	if fixture.registry.exchanges != 2 || finalizer.calls != 2 {
		t.Fatalf("exchanges=%d finalizations=%d",
			fixture.registry.exchanges, finalizer.calls)
	}
}

func TestAirGapDeliveryWaitsForImportedAcknowledgement(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	prepared, err := fixture.service.PrepareInitial(
		context.Background(),
		api.InitialAuthorityRequest{
			OperationID: testOperationID,
			State:       fixture.state,
			IssuedAt:    testNow,
			Policy:      api.ExpiryAirGap,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	processor, err := NewDeliveryProcessor(
		fixture.store, nil, testClock{},
		fixture.service.trust, "", &deliveryFinalizer{}, true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err = processor.Process(
		context.Background(), prepared.OperationID,
	); err != ErrDeliveryAwaitingImport {
		t.Fatalf("air-gap delivery error = %v", err)
	}
	if fixture.registry.exchanges != 0 {
		t.Fatal("air-gap delivery contacted a registry")
	}
}
