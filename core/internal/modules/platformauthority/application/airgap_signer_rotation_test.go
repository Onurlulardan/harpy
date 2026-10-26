package application

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func TestAirGapSignerRotationExportImportActivatesExactlyOnce(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	initializeFixtureAuthority(t, fixture)
	request := rotationRequest(t, fixture)
	request.Policy = api.ExpiryAirGap
	airgap := &AirGapCoordinator{base: fixture.service, authorizer: allowAckImport{}}
	exported, err := airgap.PrepareSignerRotationExport(context.Background(), request)
	if err != nil || len(exported.Bytes) == 0 {
		t.Fatalf("rotation export = %+v, err=%v", exported, err)
	}
	operation, err := fixture.store.LoadOperation(context.Background(), exported.OperationID)
	if err != nil {
		t.Fatal(err)
	}
	ack, err := fixture.registry.ExchangeProposal(context.Background(), operation.Proposal)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := airgap.ImportAcknowledgement(context.Background(), api.AirGapAckImport{
		OperationID: exported.OperationID, AckBytes: ack,
	})
	if err != nil || receipt.Generation != 2 {
		t.Fatalf("rotation import = %+v, err=%v", receipt, err)
	}
	replayed, err := airgap.ImportAcknowledgement(context.Background(), api.AirGapAckImport{
		OperationID: exported.OperationID, AckBytes: ack,
	})
	if err != nil || replayed.AckSHA256 != receipt.AckSHA256 {
		t.Fatalf("rotation import replay = %+v, err=%v", replayed, err)
	}
}
