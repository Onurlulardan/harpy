package iam

import (
	"context"
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *invitationHTTPApplication) acceptInvitation(
	ctx context.Context,
	tenantID uuid.UUID,
	presented *invitation.PresentedCredential,
	verified invitation.CredentialRecord,
	input InvitationAcceptInput,
	credential password.Argon2idCredential,
	directoryHMAC [32]byte,
	operationID uuid.UUID,
	requestSHA [32]byte,
	ids invitationRegistrationIDSet,
) error {
	selector, _ := presented.Selector()
	binding, _ := invitation.InvitationBindingSHA256(tenantID, selector)
	lookup, err := platformtx.NewTenantRouteCredentialLookup(
		"iam.invitation.accept", platformtx.CredentialInvitation, tenantID, binding,
	)
	if err != nil {
		return ErrInvitationHTTPUnavailable
	}
	err = application.coordinator.InCredentialTenant(
		ctx, platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			record, lookupErr := application.runtime.service.LookupCredential(
				scoped, scope, invitation.CredentialLookup{
					Selector: selector, BindingSHA256: binding,
				},
			)
			if lookupErr != nil || record.Invitation.ID != verified.Invitation.ID {
				return invitation.ErrDenied
			}
			proof, proofErr := presented.Proof(
				tenantID, record.Proof.ID, record.Proof.Revision,
				record.Proof.HashKeyID, record.Proof.SecretHMAC, application.keys,
			)
			if proofErr != nil {
				return invitation.ErrDenied
			}
			profile, profileErr := application.profiles(
				record.ProfileCatalog, input.FirstName, input.LastName,
				input.CustomAttributes,
			)
			if profileErr != nil {
				return profileErr
			}
			_, acceptErr := application.runtime.service.Accept(
				scoped, scope, invitation.AcceptCommand{
					OperationID: operationID, RequestSHA256: requestSHA,
					ID: record.Invitation.ID, ExpectedRevision: record.Invitation.Revision,
					Token: proof, AcceptedUserID: ids.user,
					AcceptedIdentifierID: ids.identifier, AccessModeID: ids.accessMode,
					RoleAssignmentID:         ids.assignment,
					OrganizationMembershipID: ids.membership,
					Profile: invitation.RegistrationProfile{
						SchemaID: profile.schemaID, Values: profile.values, SHA256: profile.digest,
					},
					Password: invitation.RegistrationPassword{
						Salt: credential.Salt(), Tag: credential.Tag(),
					},
					Directory: invitation.DirectoryProjection{
						LookupKeyID: application.directory.id, HMAC: directoryHMAC,
						SourceSHA256: sha256.Sum256([]byte(record.Invitation.NormalizedValue)),
					},
					ExpectedAuthorizationVersion: 1,
				},
			)
			clear(profile.values)
			return acceptErr
		},
	)
	if err != nil {
		return mapInvitationCredentialError(err)
	}
	return nil
}
