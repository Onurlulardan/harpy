package iam

import (
	"context"
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	publicapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) Get(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.GetInput,
) (publicresource.ReadResult, error) {
	evidence, resource, id, err := identityReadInput(ctx, scope, actor, input)
	if err != nil {
		return publicresource.ReadResult{}, err
	}
	authority, err := newIdentitySystemResourceAuthorization(
		port.application, ctx, scope, evidence, resource,
	)
	if err != nil {
		return publicresource.ReadResult{}, err
	}
	permissionID, err := authority.require("read")
	if err != nil {
		return publicresource.ReadResult{}, err
	}
	catalog, published, err := port.resolveCatalogSnapshot(
		ctx, scope, resource, authority,
	)
	if err != nil {
		return publicresource.ReadResult{}, err
	}
	record, err := port.readRecord(
		ctx, scope, evidence, permissionID, resource, id, catalog,
	)
	if err != nil {
		return publicresource.ReadResult{}, identitySystemResourceError(err)
	}
	item, err := port.relationReadItem(
		ctx, scope, published, resource, catalog, id,
	)
	if err != nil {
		return publicresource.ReadResult{}, identitySystemResourceError(err)
	}
	relations, err := port.relationProjections(
		ctx, scope, evidence, resource, []json.RawMessage{item},
	)
	if err != nil {
		return publicresource.ReadResult{}, identitySystemResourceError(err)
	}
	if len(relations) != 1 {
		return publicresource.ReadResult{}, publicresource.ErrUnavailable
	}
	return publicresource.NewReadResultWithRelations(record, catalog, relations[0])
}

func identityReadInput(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.GetInput,
) (identityPublicEvidence, publicresource.Resource, uuid.UUID, error) {
	evidence, evidenceErr := newIdentityPublicEvidence(actor)
	resource, resourceErr := input.Resource()
	id, idErr := input.ID()
	if evidenceErr != nil || resourceErr != nil || idErr != nil ||
		identitySystemResourceScope(
			ctx, scope, evidence, platformtx.SnapshotQuery, "restquery.execute",
		) != nil {
		return identityPublicEvidence{}, "", uuid.Nil, publicresource.ErrContract
	}
	return evidence, resource, uuid.UUID(id), nil
}

func (port *identitySystemResourcePort) readRecord(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	permissionID uuid.UUID,
	resource publicresource.Resource,
	id uuid.UUID,
	catalog publicresource.FieldCatalog,
) (publicresource.Record, error) {
	queryEvidence := identityQueryEvidence(evidence, permissionID)
	switch resource {
	case publicresource.Users:
		user, err := port.application.identities.users.GetUser(
			ctx, scope, userapp.GetUserQuery{
				ID: identityfacade.IdentityID(id), Evidence: queryEvidence,
			},
		)
		if err != nil {
			return publicresource.Record{}, err
		}
		return publicapp.UserRecord(user, catalog)
	case publicresource.Roles:
		role, err := port.application.identities.roles.GetRole(
			ctx, scope, roleapp.GetRoleQuery{
				ID: identityfacade.IdentityID(id), Evidence: queryEvidence,
			},
		)
		if err != nil {
			return publicresource.Record{}, err
		}
		return publicapp.RoleRecord(role, catalog)
	case publicresource.Organizations:
		organization, err := port.application.identities.organizations.GetOrganization(
			ctx, scope, organizationapp.GetOrganizationQuery{
				ID: identityfacade.IdentityID(id), Evidence: queryEvidence,
			},
		)
		if err != nil {
			return publicresource.Record{}, err
		}
		return publicapp.OrganizationRecord(organization, catalog)
	default:
		return publicresource.Record{}, publicresource.ErrContract
	}
}
