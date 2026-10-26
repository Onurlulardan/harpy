package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetSelfRequiresActorBoundSnapshot(t *testing.T) {
	_, err := fixtureCommands(t).GetSelf(
		context.Background(), transaction.TenantScope{}, SelfEvidence{
			ActorID: fixtureActorID, ExpectedTokenVersion: 2,
			ExpectedAuthorizationVersion: 3,
		},
	)
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("unbound self read error = %v", err)
	}
}

func TestSelfEvidenceRejectsMissingVersions(t *testing.T) {
	err := (SelfEvidence{ActorID: fixtureActorID}).Validate()
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("missing version evidence error = %v", err)
	}
}
