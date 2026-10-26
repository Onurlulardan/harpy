package iam

import (
	"context"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identityBulkDeleteCommand func(
	context.Context, platformtx.TenantScope, identityfacade.MutationEvidence,
	IdentityMutation,
) (identityfacade.MutationResult, error)

func (application *identityHTTPApplication) DeleteUsers(
	ctx context.Context, claims accesstoken.Claims, input IdentityBulkDelete,
) (int, error) {
	return application.deleteIdentityCollection(
		ctx, claims, input, IdentityUsers, "user", "iam.identity.user.delete",
		func(scoped context.Context, scope platformtx.TenantScope,
			evidence identityfacade.MutationEvidence, item IdentityMutation,
		) (identityfacade.MutationResult, error) {
			return application.identities.users.DeleteUser(
				scoped, scope, userapp.UserStateCommand{
					ID:               identityfacade.IdentityID(item.ID),
					ExpectedRevision: item.ExpectedRevision, Evidence: evidence,
				},
			)
		},
	)
}

func (application *identityHTTPApplication) DeleteRoles(
	ctx context.Context, claims accesstoken.Claims, input IdentityBulkDelete,
) (int, error) {
	return application.deleteIdentityCollection(
		ctx, claims, input, IdentityRoles, "role", "iam.identity.role.retire",
		func(scoped context.Context, scope platformtx.TenantScope,
			evidence identityfacade.MutationEvidence, item IdentityMutation,
		) (identityfacade.MutationResult, error) {
			return application.identities.roles.RetireRole(
				scoped, scope, roleapp.RetireRoleCommand{
					ID:               identityfacade.IdentityID(item.ID),
					ExpectedRevision: item.ExpectedRevision, Evidence: evidence,
				},
			)
		},
	)
}

func (application *identityHTTPApplication) DeleteOrganizations(
	ctx context.Context, claims accesstoken.Claims, input IdentityBulkDelete,
) (int, error) {
	return application.deleteIdentityCollection(
		ctx, claims, input, IdentityOrganizations, "organization",
		"iam.identity.organization.delete",
		func(scoped context.Context, scope platformtx.TenantScope,
			evidence identityfacade.MutationEvidence, item IdentityMutation,
		) (identityfacade.MutationResult, error) {
			return application.identities.organizations.DeleteOrganization(
				scoped, scope, organizationapp.OrganizationStateCommand{
					ID:               identityfacade.IdentityID(item.ID),
					ExpectedRevision: item.ExpectedRevision, Evidence: evidence,
				},
			)
		},
	)
}

func (application *identityHTTPApplication) deleteIdentityCollection(
	ctx context.Context,
	claims accesstoken.Claims,
	input IdentityBulkDelete,
	resource IdentityResource,
	resourceKind string,
	commandKey string,
	command identityBulkDeleteCommand,
) (int, error) {
	items, valid := canonicalIdentityBulkDelete(input)
	if !valid || application == nil || command == nil ||
		application.identities == nil || application.identities.bulkDeletes == nil {
		return 0, ErrIdentityHTTPInvalid
	}
	requestSHA, err := identityRequestSHA(struct {
		Resource string
		Items    []IdentityMutation
	}{resourceKind, items})
	if err != nil {
		return 0, ErrIdentityHTTPInvalid
	}
	operationID := identityOperationID(
		claims, commandKey+".collection", input.IdempotencyKey,
	)
	err = application.inCommand(
		ctx, claims, commandKey, resource, "delete",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			replayed, reserveErr := application.identities.bulkDeletes.ReserveBulkDelete(
				scoped, scope, identityfacade.BulkDeleteReceiptCommand{
					OperationID:  identityfacade.IdentityID(operationID),
					ActorID:      identityfacade.IdentityID(claims.Subject()),
					ResourceKind: resourceKind, RequestSHA256: requestSHA,
					DeletedCount: len(items),
				},
			)
			if reserveErr != nil || replayed {
				return reserveErr
			}
			for _, item := range items {
				childSHA, hashErr := identityRequestSHA(struct {
					BulkSHA          [32]byte
					ID               uuid.UUID
					ExpectedRevision int64
				}{requestSHA, item.ID, item.ExpectedRevision})
				if hashErr != nil {
					return ErrIdentityHTTPInvalid
				}
				childID := identityChildID(operationID, "delete-"+item.ID.String())
				evidence := identityMutationEvidence(
					claims, permissionID, childID, childSHA,
				)
				result, commandErr := command(scoped, scope, evidence, item)
				if commandErr != nil {
					return commandErr
				}
				if finishErr := application.finishIdentityMutation(
					scoped, scope, evidence, item.ID, resourceKind,
					commandKey, result.Replayed,
				); finishErr != nil {
					return finishErr
				}
			}
			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func canonicalIdentityBulkDelete(input IdentityBulkDelete) ([]IdentityMutation, bool) {
	if input.IdempotencyKey == "" || len(input.Items) < 1 ||
		len(input.Items) > identityfacade.MaximumBulkDeleteItems {
		return nil, false
	}
	items := append([]IdentityMutation(nil), input.Items...)
	sort.Slice(items, func(left, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})
	for index, item := range items {
		if item.ID == uuid.Nil || item.ExpectedRevision < 1 ||
			index > 0 && item.ID == items[index-1].ID {
			return nil, false
		}
	}
	return items, true
}
