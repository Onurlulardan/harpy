package actorbridge

import (
	"context"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (issuer *Issuer) Issue(
	ctx context.Context,
	scopeRequest platformtx.ScopeRequest,
	input IssueInput,
) (actorassertionv1.Credential, error) {
	if !issuer.Ready() || ctx == nil || ctx.Err() != nil || !input.valid() ||
		!platformtx.NetworkAllowed(ctx) {
		return actorassertionv1.Credential{}, NewFault(FaultInvalid)
	}
	var credential actorassertionv1.Credential
	err := issuer.coordinator.InTenant(
		ctx, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			if issuer.closed.Load() {
				return NewFault(FaultUnavailable)
			}
			authorization, err := issuer.authorizer.AuthorizeActorBridge(
				scoped, scope, input.ResourceKey, input.ActionKey,
				input.TokenVersion, input.AuthorizationVersion,
			)
			if err != nil {
				return runtimeFault(err, FaultUnauthorized)
			}
			authority, err := issuer.authority.ResolveScopedIssueAuthority(
				scoped, scope, input.TargetDomain,
			)
			if err != nil {
				return runtimeFault(err, FaultAuthority)
			}
			decision, err := input.decision(authorization, authority)
			if err != nil {
				return err
			}
			timepoint := issuer.now().UTC().Truncate(time.Second)
			if timepoint.IsZero() {
				return NewFault(FaultUnavailable)
			}
			credential, err = issuer.signer.Issue(decision, timepoint)
			return runtimeFault(err, FaultUnavailable)
		},
	)
	if err != nil {
		return actorassertionv1.Credential{}, runtimeFault(err, FaultUnavailable)
	}
	if _, err := credential.Wire(); err != nil {
		return actorassertionv1.Credential{}, NewFault(FaultUnavailable)
	}
	return credential, nil
}
