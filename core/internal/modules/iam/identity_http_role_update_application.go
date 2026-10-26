package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) UpdateRole(
	ctx context.Context,
	claims accesstoken.Claims,
	input IdentityRoleUpdate,
) (IdentityRoleView, error) {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 || input.IdempotencyKey == "" {
		return IdentityRoleView{}, ErrIdentityHTTPInvalid
	}
	current, err := application.getRole(ctx, claims, input.ID, "update")
	if err != nil {
		return IdentityRoleView{}, err
	}
	name, description, defaultValue := current.Name, current.Description, current.Default
	if input.Name != nil {
		name = *input.Name
	}
	if input.Description != nil {
		description = *input.Description
	}
	if input.Default != nil {
		defaultValue = *input.Default
	}
	patch, err := identityProfileValues(input.CustomAttributes)
	if err != nil {
		return IdentityRoleView{}, err
	}
	operationID := identityOperationID(
		claims, "iam.identity.role.update", input.IdempotencyKey,
	)
	requestSHA, err := identityRequestSHA(struct {
		ID                uuid.UUID
		ExpectedRevision  int64
		Name, Description string
		Default           bool
		Patch             []byte
	}{input.ID, input.ExpectedRevision, name, description, defaultValue, patch})
	if err != nil {
		return IdentityRoleView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.role.update", IdentityRoles, "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			storedPatch, profileErr := application.profileDocumentValues(
				scoped, scope, identityfacade.ProfileRole, patch,
			)
			if profileErr != nil {
				return profileErr
			}
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, updateErr := application.identities.roles.UpdateRole(
				scoped, scope, roleapp.UpdateRoleCommand{
					ID:               identityfacade.IdentityID(input.ID),
					ExpectedRevision: input.ExpectedRevision,
					Name:             name, Description: description, Default: defaultValue,
					ProfilePatch: storedPatch,
					Evidence:     evidence,
				},
			)
			if updateErr != nil {
				return updateErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "role",
				"iam.identity.role.update", result.Replayed,
			)
		},
	)
	if err != nil {
		return IdentityRoleView{}, err
	}
	return application.getRole(ctx, claims, input.ID, "update")
}
