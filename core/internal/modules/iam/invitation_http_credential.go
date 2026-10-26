package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *invitationHTTPApplication) invitationCredential(
	ctx context.Context,
	tenantID uuid.UUID,
	presented *invitation.PresentedCredential,
) (invitation.CredentialRecord, invitation.TokenProof, error) {
	selector, err := presented.Selector()
	if err != nil {
		return invitation.CredentialRecord{}, invitation.TokenProof{},
			ErrInvitationHTTPInvalidCredential
	}
	binding, err := invitation.InvitationBindingSHA256(tenantID, selector)
	if err != nil {
		return invitation.CredentialRecord{}, invitation.TokenProof{},
			ErrInvitationHTTPInvalidCredential
	}
	lookup, err := platformtx.NewTenantRouteCredentialLookup(
		"iam.invitation.accept", platformtx.CredentialInvitation, tenantID, binding,
	)
	if err != nil {
		return invitation.CredentialRecord{}, invitation.TokenProof{},
			ErrInvitationHTTPUnavailable
	}
	var record invitation.CredentialRecord
	var proof invitation.TokenProof
	err = application.coordinator.InCredentialTenant(
		ctx, platformtx.CredentialScopeRequest{
			Lookup: lookup, Resolver: application.resolver,
		},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			var lookupErr error
			record, lookupErr = application.runtime.service.LookupCredential(
				scoped, scope, invitation.CredentialLookup{
					Selector: selector, BindingSHA256: binding,
				},
			)
			if lookupErr != nil {
				return lookupErr
			}
			proof, lookupErr = presented.Proof(
				tenantID, record.Proof.ID, record.Proof.Revision,
				record.Proof.HashKeyID, record.Proof.SecretHMAC, application.keys,
			)
			return lookupErr
		},
	)
	if err != nil {
		return invitation.CredentialRecord{}, invitation.TokenProof{},
			mapInvitationCredentialError(err)
	}
	return record, proof, nil
}

func parseInvitationCredential(
	wire []byte,
) (*invitation.PresentedCredential, error) {
	presented, err := invitation.ParseCredential(wire)
	if err != nil {
		return nil, ErrInvitationHTTPInvalidCredential
	}
	return presented, nil
}
