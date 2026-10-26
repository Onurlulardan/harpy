package actorbridge

import (
	"sync/atomic"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type Clock func() time.Time

type Issuer struct {
	coordinator TenantCoordinator
	authorizer  AuthorizationPort
	authority   ScopedIssueAuthorityPort
	signer      *actorassertionv1.Signer
	now         Clock
	closed      atomic.Bool
}

func NewIssuer(
	coordinator TenantCoordinator,
	authorizer AuthorizationPort,
	authority ScopedIssueAuthorityPort,
	signer *actorassertionv1.Signer,
	now Clock,
) (*Issuer, error) {
	if interfaceIsNil(coordinator) || interfaceIsNil(authorizer) ||
		interfaceIsNil(authority) || signer == nil || now == nil {
		return nil, NewFault(FaultInvalid)
	}
	return &Issuer{coordinator: coordinator, authorizer: authorizer,
		authority: authority, signer: signer, now: now}, nil
}

func (issuer *Issuer) Ready() bool {
	return issuer != nil && !issuer.closed.Load() &&
		!interfaceIsNil(issuer.coordinator) && !interfaceIsNil(issuer.authorizer) &&
		!interfaceIsNil(issuer.authority) && issuer.signer != nil && issuer.now != nil
}

func (issuer *Issuer) Close() {
	if issuer == nil || !issuer.closed.CompareAndSwap(false, true) {
		return
	}
	if issuer.signer != nil {
		issuer.signer.Close()
	}
}
