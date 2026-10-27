package iam

import (
	"context"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *profileHTTPApplication) RetireField(
	ctx context.Context,
	claims accesstoken.Claims,
	input ProfileSchemaRetire,
) (ProfileSchemaMutationView, error) {
	if input.IdempotencyKey == "" || input.SchemaID == uuid.Nil ||
		!input.Resource.Valid() || input.ExpectedActiveSchemaID == uuid.Nil ||
		input.FieldID == uuid.Nil || input.ExpectedFieldRevision < 1 {
		return ProfileSchemaMutationView{}, ErrProfileHTTPInvalid
	}
	requestSHA, err := profileRetireRequestSHA(input)
	if err != nil {
		return ProfileSchemaMutationView{}, err
	}
	operationID := identityOperationID(
		claims, "iam.profile.retire", input.IdempotencyKey,
	)
	var view ProfileSchemaMutationView
	err = application.inProfileCommand(
		ctx, claims, "iam.profile.retire", input.Resource,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			reader, bindErr := iampostgres.BindProfileReader(scope)
			if bindErr != nil {
				return bindErr
			}
			active, bindErr := reader.ActiveSchema(scoped, input.Resource)
			if bindErr != nil {
				return bindErr
			}
			if active.ID() != input.ExpectedActiveSchemaID {
				return ErrProfileHTTPConflict
			}
			change, bindErr := profileRetirementChange(active, input)
			if bindErr != nil {
				return bindErr
			}
			result, retireErr := application.commands.Retire(
				scoped, scope, profileapp.RetireCommand{
					Change: change,
					Target: profileapp.RetirementTarget{
						Resource: input.Resource, FieldID: input.FieldID,
						ExpectedFieldRevision: input.ExpectedFieldRevision,
					},
					Evidence: profileMutationEvidence(
						claims, permissionID, operationID, requestSHA,
					),
				},
			)
			if retireErr == nil {
				view = profileMutationView(result)
			}
			return retireErr
		},
	)
	return view, err
}
