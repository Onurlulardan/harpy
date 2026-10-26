package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type allowAckImport struct{}

func (allowAckImport) AuthorizeAckImport(context.Context, string, string) error { return nil }

func TestImportAcknowledgementCommitsAndRejectsDifferentReplay(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	state, err := artifact.CanonicalizeState(artifactState(fixture.state))
	if err != nil {
		t.Fatal(err)
	}
	operation, err := fixture.service.prepareOperation(context.Background(), operationCommand{
		operationID: testOperationID, purpose: "initial_environment_authority",
		state: state, generation: 1, issuedAt: testNow, policy: artifact.ExpiryAirGap,
	})
	if err != nil {
		t.Fatal(err)
	}
	ack, err := fixture.registry.ExchangeProposal(context.Background(), operation.Proposal)
	if err != nil {
		t.Fatal(err)
	}
	service := &AirGapCoordinator{base: fixture.service, authorizer: allowAckImport{}}
	receipt, err := service.ImportAcknowledgement(context.Background(), api.AirGapAckImport{
		OperationID: testOperationID, AckBytes: ack,
	})
	if err != nil || receipt.Generation != 1 {
		t.Fatalf("import receipt = %+v, %v", receipt, err)
	}
	_, err = service.ImportAcknowledgement(context.Background(), api.AirGapAckImport{
		OperationID: testOperationID, AckBytes: append(append([]byte(nil), ack...), ' '),
	})
	if !errors.Is(err, api.ErrConflict) {
		t.Fatalf("different replay error = %v", err)
	}
}
