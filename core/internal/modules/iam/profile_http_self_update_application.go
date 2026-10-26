package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *profileHTTPApplication) UpdateSelf(
	ctx context.Context,
	claims accesstoken.Claims,
	input ProfileSelfUpdate,
) (ProfileSelfView, error) {
	if application == nil || input.IdempotencyKey == "" || input.Values == nil {
		return ProfileSelfView{}, ErrProfileHTTPInvalid
	}
	requestSHA, err := identityRequestSHA(input.Values)
	if err != nil {
		return ProfileSelfView{}, ErrProfileHTTPInvalid
	}
	operationID := identityOperationID(claims, "iam.profile.patch", input.IdempotencyKey)
	request, err := application.commandScopeRequest(ctx, claims, "iam.profile.patch")
	if err != nil {
		return ProfileSelfView{}, err
	}
	var view ProfileSelfView
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			record, document, schema, loadErr := application.loadSelfProfile(
				scoped, scope, claims,
			)
			if loadErr != nil {
				return loadErr
			}
			patch, patchErr := profileSelfPatch(schema, input.Values)
			if patchErr != nil {
				return patchErr
			}
			prepared, patchErr := application.selfWriter.PrepareSelfPatch(
				scoped, scope, claims.Subject(), patch,
			)
			if patchErr != nil {
				return patchErr
			}
			result, patchErr := application.commands.UpdateSelf(
				scoped, scope, profileapp.SelfUpdateCommand{
					OperationID: operationID, RequestSHA256: requestSHA,
					ExpectedSourceSchemaID: prepared.ExpectedSourceSchemaID,
					ExpectedRevision:       record.Revision, Target: prepared.Target,
					ActorID:                      claims.Subject(),
					ExpectedTokenVersion:         claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
				},
			)
			if patchErr != nil {
				return patchErr
			}
			if !result.Replayed {
				change := changeapp.NewChange(
					identityfacade.MutationEvidence{
						OperationID:   identityfacade.IdentityID(operationID),
						ActorID:       identityfacade.IdentityID(claims.Subject()),
						RequestSHA256: requestSHA,
					},
					identityfacade.IdentityID(claims.Subject()), "user",
					"iam.identity.user.update",
				)
				if application.changes.WriteChange(scoped, scope, change) != nil {
					return ErrProfileHTTPUnavailable
				}
			}
			record, document, schema, patchErr = application.loadSelfProfile(
				scoped, scope, claims,
			)
			if patchErr != nil || record.Revision != result.Revision {
				if patchErr != nil {
					return patchErr
				}
				return ErrProfileHTTPUnavailable
			}
			view, patchErr = newProfileSelfView(record, schema, document)
			return patchErr
		},
	)
	return view, profileApplicationError(err)
}
