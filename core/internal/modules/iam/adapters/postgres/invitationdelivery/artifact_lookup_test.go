package invitationdelivery

import (
	"context"
	"errors"
	"testing"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

var errInvitationDeliveryRow = errors.New("invitation delivery row failed")

type failingDeliveryDatabase struct{}

func (failingDeliveryDatabase) Execute(
	context.Context, string, ...any,
) (platformpostgres.CommandResult, error) {
	return platformpostgres.CommandResult{}, errInvitationDeliveryRow
}

func (failingDeliveryDatabase) QueryRow(
	context.Context, string, ...any,
) platformpostgres.Row {
	return failingDeliveryRow{}
}

type failingDeliveryRow struct{}

func (failingDeliveryRow) Scan(...any) error { return errInvitationDeliveryRow }

func TestArtifactLookupPropagatesDatabaseFailure(t *testing.T) {
	value := store{database: failingDeliveryDatabase{}}
	_, err := value.claimArtifact(context.Background(), invitationdeliveryapp.ClaimCommand{})
	if !errors.Is(err, errInvitationDeliveryRow) {
		t.Fatalf("read artifact error = %v", err)
	}
}
