package actorbridge

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type coordinatorStub struct {
	order     *[]string
	committed bool
	err       error
}

func (coordinator *coordinatorStub) InTenant(
	ctx context.Context,
	_ platformtx.ScopeRequest,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	if coordinator.err != nil {
		return coordinator.err
	}
	err := execute(ctx, platformtx.TenantScope{})
	if err != nil {
		return err
	}
	*coordinator.order = append(*coordinator.order, "commit")
	coordinator.committed = true
	return nil
}

type authorizationStub struct {
	order  *[]string
	result Authorization
	err    error
}

func (stub authorizationStub) AuthorizeActorBridge(
	context.Context,
	platformtx.TenantScope,
	string,
	string,
	int64,
	int64,
) (Authorization, error) {
	*stub.order = append(*stub.order, "authorize")
	return stub.result, stub.err
}

type authorityStub struct {
	order  *[]string
	result Authority
	err    error
}

func (stub authorityStub) ResolveScopedIssueAuthority(
	context.Context,
	platformtx.TenantScope,
	actorassertionv1.Domain,
) (Authority, error) {
	*stub.order = append(*stub.order, "authority")
	return stub.result, stub.err
}

func TestIssuerOrdersCurrentIAMAuthorityAndReturnsOnlyAfterCommit(t *testing.T) {
	now := time.Unix(2_000_000_000, 0).UTC()
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{17}, ed25519.SeedSize))
	publicKey := append(ed25519.PublicKey(nil), privateKey[ed25519.SeedSize:]...)
	signer, err := actorassertionv1.NewSigner(
		"edge-2026-07", privateKey, bytes.NewReader(bytes.Repeat([]byte{3}, 16)),
	)
	if err != nil {
		t.Fatal(err)
	}
	authorization, _ := NewAuthorization(
		uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		uuid.MustParse("30000000-0000-4000-8000-000000000003"), 7, 19, 23,
	)
	authority, _ := NewAuthority(11, 12, 13)
	target, _ := actorassertionv1.TargetBindingSHA256("POST", "/records", "record-17")
	input := IssueInput{TargetDomain: actorassertionv1.DomainForms,
		ResourceKey: "forms.records", ActionKey: "create", TargetBinding: target,
		TokenVersion: 7, AuthorizationVersion: 19}
	order := []string{}
	coordinator := &coordinatorStub{order: &order}
	issuer, err := NewIssuer(coordinator,
		authorizationStub{order: &order, result: authorization},
		authorityStub{order: &order, result: authority}, signer,
		func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	credential, err := issuer.Issue(context.Background(), platformtx.ScopeRequest{}, input)
	if err != nil || !coordinator.committed || len(order) != 3 ||
		order[0] != "authorize" || order[1] != "authority" || order[2] != "commit" {
		t.Fatalf("issue order=%v committed=%t err=%v", order, coordinator.committed, err)
	}
	wire, _ := credential.Wire()
	verifier, _ := actorassertionv1.NewVerifier(map[string]ed25519.PublicKey{
		"edge-2026-07": publicKey,
	})
	_, err = verifier.Verify(wire, actorassertionv1.Expectation{
		TenantID: authorization.TenantID(), TargetDomain: input.TargetDomain,
		ResourceKey: input.ResourceKey, ActionKey: input.ActionKey,
		TargetBindingSHA256: input.TargetBinding, IAMGroupGeneration: 11,
		IAMGeneration: 12, TargetGeneration: 13,
	}, now)
	if err != nil {
		t.Fatalf("issued credential does not bind evidence: %v", err)
	}
	issuer.Close()
	if issuer.Ready() {
		t.Fatal("closed issuer remained ready")
	}
}

func TestIssuerStopsBeforeAuthorityWhenIAMDenies(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{18}, ed25519.SeedSize))
	signer, _ := actorassertionv1.NewProductionSigner("edge-2026-07", privateKey)
	order := []string{}
	issuer, _ := NewIssuer(&coordinatorStub{order: &order},
		authorizationStub{order: &order, err: NewFault(FaultUnauthorized)},
		authorityStub{order: &order}, signer, time.Now)
	target, _ := actorassertionv1.TargetBindingSHA256("POST", "/records")
	_, err := issuer.Issue(context.Background(), platformtx.ScopeRequest{}, IssueInput{
		TargetDomain: actorassertionv1.DomainForms, ResourceKey: "forms.records",
		ActionKey: "create", TargetBinding: target, TokenVersion: 1,
		AuthorizationVersion: 1,
	})
	if !FaultIs(err, FaultUnauthorized) || len(order) != 1 || order[0] != "authorize" {
		t.Fatalf("denied issue order=%v err=%v", order, err)
	}
}
