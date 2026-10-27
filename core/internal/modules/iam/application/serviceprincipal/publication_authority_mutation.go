package serviceprincipal

import (
	"context"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *PublicationAuthorityApplication) createPrincipal(
	ctx context.Context,
	scope transaction.TenantScope,
	request ProvisionPublicationAuthority,
	permissionID uuid.UUID,
) (uuid.UUID, error) {
	evidence := authorityMutationEvidence(request, permissionID,
		authorityOperationID(request, "principal"),
		authorityRequestDigest(request, publicationPrincipalID, "principal"))
	result, err := application.commands.Create(ctx, scope, CreateCommand{
		ID: publicationPrincipalID, Key: publicationPrincipalKey,
		DisplayName: publicationPrincipalName, Evidence: evidence,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return result.ID, nil
}

func (application *PublicationAuthorityApplication) createGrant(
	ctx context.Context,
	scope transaction.TenantScope,
	request ProvisionPublicationAuthority,
	permissionID, principalID, grantID uuid.UUID,
) (PublicationAuthority, error) {
	capability, err := ParseCapabilityKey("workflow.instance.execute")
	if err != nil {
		return PublicationAuthority{}, ErrEvidenceMissing
	}
	evidence := authorityMutationEvidence(request, permissionID,
		authorityOperationID(request, "grant"),
		authorityRequestDigest(request, grantID, "grant"))
	result, err := application.commands.Grant(ctx, scope, GrantCommand{
		GrantID: grantID, PrincipalID: principalID, ExpectedRevision: 0,
		Capability: capability, Scope: ResourceScope{Kind: ScopeResource,
			Owner: "workflow", ResourceKind: "definition",
			ResourceID: request.DefinitionID},
		ValidFrom: time.Unix(0, 0).UTC(), ManifestSHA256: request.ManifestSHA256,
		Evidence: evidence,
	})
	if err != nil {
		return PublicationAuthority{}, err
	}
	return PublicationAuthority{ServicePrincipalID: principalID,
		GrantID: result.ID, GrantRevision: result.Revision}, nil
}

func authorityMutationEvidence(
	request ProvisionPublicationAuthority,
	permissionID, operationID uuid.UUID,
	requestDigest [32]byte,
) MutationEvidence {
	evidence := request.MutationEvidence
	return MutationEvidence{
		OperationID: operationID, RequestSHA256: requestDigest,
		ActorID: evidence.ActorID, ManagePermissionID: permissionID,
		ExpectedTokenVersion:         evidence.ExpectedTokenVersion,
		ExpectedAuthorizationVersion: evidence.ExpectedAuthorizationVersion,
		PurposeSHA256:                evidence.PurposeSHA256, AuthTime: evidence.AuthenticatedAt,
		AMRPassword: evidence.AMRPassword, AMRWebAuthn: evidence.AMRWebAuthn,
		AMRTOTP: evidence.AMRTOTP,
	}
}

func validPublicationAuthorityRequest(
	ctx context.Context,
	scope transaction.TenantScope,
	request ProvisionPublicationAuthority,
) bool {
	evidence := request.MutationEvidence
	expectedPurpose, purposeErr := PublicationAuthorityStepUpPurpose(
		scope.TenantID(), request.DefinitionID,
	)
	mutation := MutationEvidence{
		OperationID: publicationPrincipalID, RequestSHA256: request.ManifestSHA256,
		ActorID: evidence.ActorID, ManagePermissionID: publicationPrincipalID,
		ExpectedTokenVersion:         evidence.ExpectedTokenVersion,
		ExpectedAuthorizationVersion: evidence.ExpectedAuthorizationVersion,
		PurposeSHA256:                evidence.PurposeSHA256, AuthTime: evidence.AuthenticatedAt,
		AMRPassword: evidence.AMRPassword, AMRWebAuthn: evidence.AMRWebAuthn,
		AMRTOTP: evidence.AMRTOTP,
	}
	return ctx != nil && ctx.Err() == nil && request.DefinitionID != uuid.Nil &&
		request.DefinitionID.Version() == 7 && request.ManifestSHA256 != [32]byte{} &&
		validAuthorityIdempotencyKey(request.IdempotencyKey) &&
		purposeErr == nil && evidence.PurposeSHA256 == expectedPurpose &&
		scope.OperationID() == publicationAuthorityOperation &&
		mutationScopeUsable(ctx, scope, mutation)
}

func validAuthorityIdempotencyKey(value string) bool {
	if len(value) < 16 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if !(character >= 'A' && character <= 'Z') &&
			!(character >= 'a' && character <= 'z') &&
			!(character >= '0' && character <= '9') &&
			character != '_' && character != '.' && character != ':' &&
			character != '-' {
			return false
		}
	}
	return true
}

func authorityPortNil(value PublicationAuthorityPort) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}
