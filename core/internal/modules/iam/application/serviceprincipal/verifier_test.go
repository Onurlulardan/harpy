package serviceprincipal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type evidenceStub struct {
	principal Principal
	grant     GrantRevision
}

func (stub evidenceStub) Evidence(
	context.Context,
	transaction.TenantScope,
	uuid.UUID,
	int64,
) (Principal, GrantRevision, error) {
	return stub.principal, stub.grant, nil
}

func TestVerifierRejectsUnboundTenantScope(t *testing.T) {
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	principalID, grantID := uuid.New(), uuid.New()
	capability, err := ParseCapabilityKey("forms.record.update")
	if err != nil {
		t.Fatal(err)
	}
	scope := ResourceScope{
		Kind: ScopeTenant, Owner: "forms", ResourceKind: "record",
	}
	verifier, err := NewVerifier(evidenceStub{
		principal: Principal{
			ID: principalID, Kind: KindWorkflow,
			State:                PrincipalActive,
			AuthorizationVersion: 1, Revision: 1,
		},
		grant: GrantRevision{
			GrantID: grantID, Revision: 1, PrincipalID: principalID,
			Capability: capability, Scope: scope, ValidFrom: now.Add(-time.Minute),
			State: GrantActive, ManifestSHA256: [32]byte{1},
		},
	}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	err = verifier.Verify(context.Background(), transaction.TenantScope{},
		PinnedRequest{
			PrincipalID: principalID, Capability: capability, Scope: scope, GrantID: grantID,
			GrantRevision: 1, Manifest: [32]byte{1},
		})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("unbound scope error = %v", err)
	}
}

func TestVerifierRequiresExactPinnedPrincipal(t *testing.T) {
	principal := Principal{ID: uuid.New()}
	if exactPinnedPrincipal(principal, PinnedRequest{PrincipalID: uuid.New()}) {
		t.Fatal("mismatched principal was accepted")
	}
	if !exactPinnedPrincipal(principal, PinnedRequest{PrincipalID: principal.ID}) {
		t.Fatal("exact pinned principal was rejected")
	}
}

func TestPinnedRequestRequiresPrincipalIdentity(t *testing.T) {
	capability, err := ParseCapabilityKey("workflow.instance.execute")
	if err != nil {
		t.Fatal(err)
	}
	request := PinnedRequest{
		Capability: capability,
		Scope: ResourceScope{
			Kind: ScopeTenant, Owner: "workflow", ResourceKind: "instance",
		},
		GrantID: uuid.New(), GrantRevision: 1, Manifest: [32]byte{1},
	}
	if !errors.Is(request.Validate(), ErrInvalidGrant) {
		t.Fatal("principal-less pinned request was accepted")
	}
	request.PrincipalID = uuid.New()
	if err := request.Validate(); err != nil {
		t.Fatalf("exact pinned request rejected: %v", err)
	}
}
