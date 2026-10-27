package serviceprincipal

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const (
	publicationAuthorityOperation = "workflow.publication_authority.provision"
	publicationPrincipalKey       = "workflow.publication"
	publicationPrincipalName      = "Workflow publication authority"
)

var publicationPrincipalID = uuid.MustParse("019f9e00-0000-7000-8000-000000000001")

type PublicationAuthorityPort interface {
	ManagePermissionID(context.Context, transaction.TenantScope) (uuid.UUID, error)
	PublicationPrincipal(context.Context, transaction.TenantScope) (
		PublicationAuthorityPrincipal, error,
	)
	PublicationGrant(context.Context, transaction.TenantScope, uuid.UUID) (
		PublicationAuthorityGrant, error,
	)
}

type PublicationAuthorityPrincipal struct {
	ID    uuid.UUID
	Found bool
	Kind  string
	State string
}

type PublicationAuthorityGrant struct {
	PublicationAuthority
	Found          bool
	State          string
	Capability     CapabilityKey
	Scope          ResourceScope
	ManifestSHA256 [32]byte
}

type PublicationAuthorityEvidence struct {
	ActorID                      uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
	PurposeSHA256                [32]byte
	AuthenticatedAt              time.Time
	AMRPassword                  bool
	AMRWebAuthn                  bool
	AMRTOTP                      bool
}

type ProvisionPublicationAuthority struct {
	DefinitionID     uuid.UUID
	ManifestSHA256   [32]byte
	IdempotencyKey   string
	MutationEvidence PublicationAuthorityEvidence
}

type PublicationAuthority struct {
	ServicePrincipalID uuid.UUID
	GrantID            uuid.UUID
	GrantRevision      int64
}

type PublicationAuthorityApplication struct {
	commands *Commands
	port     PublicationAuthorityPort
}

func NewPublicationAuthorityApplication(
	commands *Commands,
	port PublicationAuthorityPort,
) (*PublicationAuthorityApplication, error) {
	if commands == nil || authorityPortNil(port) {
		return nil, ErrEvidenceMissing
	}
	return &PublicationAuthorityApplication{commands: commands, port: port}, nil
}

func (application *PublicationAuthorityApplication) Provision(
	ctx context.Context,
	scope transaction.TenantScope,
	request ProvisionPublicationAuthority,
) (PublicationAuthority, error) {
	if !validPublicationAuthorityRequest(ctx, scope, request) {
		return PublicationAuthority{}, ErrInvalidGrant
	}
	permissionID, err := application.port.ManagePermissionID(ctx, scope)
	if err != nil || permissionID == uuid.Nil {
		return PublicationAuthority{}, ErrEvidenceMissing
	}
	principal, err := application.port.PublicationPrincipal(ctx, scope)
	if err != nil {
		return PublicationAuthority{}, ErrEvidenceMissing
	}
	principalID := principal.ID
	if principal.Found && (principal.ID == uuid.Nil || principal.Kind != "workflow" ||
		principal.State != "active") {
		return PublicationAuthority{}, ErrGrantInactive
	}
	if !principal.Found {
		principalID, err = application.createPrincipal(ctx, scope, request, permissionID)
		if err != nil {
			return PublicationAuthority{}, err
		}
	}
	grantID := publicationAuthorityID(
		request.DefinitionID, request.ManifestSHA256, "grant",
	)
	grant, err := application.port.PublicationGrant(ctx, scope, grantID)
	if err != nil {
		return PublicationAuthority{}, ErrEvidenceMissing
	}
	if grant.Found {
		expectedScope := ResourceScope{Kind: ScopeResource, Owner: "workflow",
			ResourceKind: "definition", ResourceID: request.DefinitionID}
		if grant.ServicePrincipalID != principalID || grant.GrantID != grantID ||
			grant.GrantRevision != 1 || grant.State != "active" ||
			grant.Capability != CapabilityKey("workflow.instance.execute") ||
			!grant.Scope.Equal(expectedScope) ||
			grant.ManifestSHA256 != request.ManifestSHA256 {
			return PublicationAuthority{}, ErrGrantMismatch
		}
		return grant.PublicationAuthority, nil
	}
	return application.createGrant(
		ctx, scope, request, permissionID, principalID, grantID,
	)
}
