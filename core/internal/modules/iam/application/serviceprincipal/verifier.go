package serviceprincipal

import (
	"context"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Clock func() time.Time

type Verifier struct {
	port  EvidencePort
	clock Clock
	valid bool
}

func NewVerifier(port EvidencePort, clock Clock) (Verifier, error) {
	if evidencePortNil(port) || clock == nil {
		return Verifier{}, ErrEvidenceMissing
	}
	return Verifier{port: port, clock: clock, valid: true}, nil
}

func (verifier Verifier) Ready() bool { return verifier.valid }

func (verifier Verifier) Verify(
	ctx context.Context,
	scope transaction.TenantScope,
	request PinnedRequest,
) error {
	_, err := verifier.VerifyEvidence(ctx, scope, request)
	return err
}

func (verifier Verifier) VerifyEvidence(
	ctx context.Context,
	scope transaction.TenantScope,
	request PinnedRequest,
) (VerificationEvidence, error) {
	if verifier.clock == nil {
		return VerificationEvidence{}, ErrInvalidGrant
	}
	return verifier.VerifyEvidenceAt(ctx, scope, request, verifier.clock().UTC())
}

func (verifier Verifier) VerifyEvidenceAt(
	ctx context.Context,
	scope transaction.TenantScope,
	request PinnedRequest,
	verifiedAt time.Time,
) (VerificationEvidence, error) {
	if !verifier.valid || evidencePortNil(verifier.port) || verifier.clock == nil ||
		request.Validate() != nil || verifiedAt.IsZero() || verifiedAt.Location() != time.UTC ||
		!commandScopeUsable(ctx, scope) {
		return VerificationEvidence{}, ErrInvalidGrant
	}
	principal, grant, err := verifier.port.Evidence(
		ctx, scope, request.GrantID, request.GrantRevision,
	)
	if err != nil {
		return VerificationEvidence{}, ErrEvidenceMissing
	}
	if !exactPinnedPrincipal(principal, request) {
		return VerificationEvidence{}, ErrGrantMismatch
	}
	if err := Evaluate(principal, grant, AuthorizationRequest{
		Capability: request.Capability, Scope: request.Scope,
		GrantID: request.GrantID, GrantRevision: request.GrantRevision,
		Manifest: request.Manifest, At: verifiedAt,
	}); err != nil {
		return VerificationEvidence{}, err
	}
	evidence := verifiedEvidence(principal, grant, verifiedAt)
	if evidence.Validate() != nil {
		return VerificationEvidence{}, ErrEvidenceMissing
	}
	return evidence, nil
}

func exactPinnedPrincipal(principal Principal, request PinnedRequest) bool {
	return principal.ID == request.PrincipalID
}

func commandScopeUsable(ctx context.Context, scope transaction.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.Command) == nil
}

func evidencePortNil(port EvidencePort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
