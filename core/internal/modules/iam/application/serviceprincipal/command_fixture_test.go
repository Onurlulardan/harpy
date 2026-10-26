package serviceprincipal

import (
	"context"
	"testing"
	"time"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type commandPortStub struct{}
type changeSinkStub struct{}

func (changeSinkStub) WriteChange(
	context.Context, transaction.TenantScope, rbacdomain.Change,
) error {
	return nil
}

func (commandPortStub) Create(
	context.Context, transaction.TenantScope, CreateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (commandPortStub) Disable(
	context.Context, transaction.TenantScope, StateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (commandPortStub) Enable(
	context.Context, transaction.TenantScope, StateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (commandPortStub) Retire(
	context.Context, transaction.TenantScope, StateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (commandPortStub) Grant(
	context.Context, transaction.TenantScope, GrantCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (commandPortStub) Revoke(
	context.Context, transaction.TenantScope, RevokeCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}

func commandFixtureEvidence() MutationEvidence {
	return MutationEvidence{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1}, ActorID: uuid.New(),
		ManagePermissionID: uuid.New(), ExpectedTokenVersion: 1,
		ExpectedAuthorizationVersion: 1, PurposeSHA256: [32]byte{2},
		AuthTime: time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC), AMRWebAuthn: true,
	}
}

func commandFixtureCommands(t *testing.T) *Commands {
	t.Helper()
	capability, err := ParseCapabilityKey("forms.record.update")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCapabilityCatalog([]CapabilityDefinition{{
		Key: capability, Owner: "forms", ResourceKind: "record",
		TenantScope: true, ResourceScope: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	commands, err := NewCommands(commandPortStub{}, catalog, changeSinkStub{})
	if err != nil {
		t.Fatal(err)
	}
	return commands
}

func commandFixtureState() StateCommand {
	return StateCommand{ID: uuid.New(), ExpectedRevision: 1, Evidence: commandFixtureEvidence()}
}
