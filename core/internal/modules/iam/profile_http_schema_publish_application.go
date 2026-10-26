package iam

import (
	"context"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *profileHTTPApplication) PublishSchema(
	ctx context.Context,
	claims accesstoken.Claims,
	input ProfileSchemaPublish,
) (ProfileSchemaMutationView, error) {
	if input.IdempotencyKey == "" {
		return ProfileSchemaMutationView{}, ErrProfileHTTPInvalid
	}
	mappings, valid := canonicalProfileMigrationMappings(input.MigrationMappings)
	if !valid {
		return ProfileSchemaMutationView{}, ErrProfileHTTPInvalid
	}
	input.MigrationMappings = mappings
	schema, err := profilePublishedSchema(input)
	if err != nil {
		return ProfileSchemaMutationView{}, err
	}
	requestSHA, err := profilePublishRequestSHA(input)
	if err != nil {
		return ProfileSchemaMutationView{}, err
	}
	operationID := identityOperationID(
		claims, "iam.profile.publish", input.IdempotencyKey,
	)
	var view ProfileSchemaMutationView
	err = application.inProfileCommand(
		ctx, claims, "iam.profile.publish", input.Resource,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			evidence := profileMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			replay, found, replayErr := application.backfills.Replay(
				scoped, scope, evidence, schema.ID(),
			)
			if replayErr != nil {
				return replayErr
			}
			if found {
				view = profileBackfillMutationView(replay)
				return nil
			}
			reader, bindErr := iampostgres.BindProfileReader(scope)
			if bindErr != nil {
				return bindErr
			}
			active, bindErr := reader.ActiveSchema(scoped, input.Resource)
			if bindErr != nil {
				return bindErr
			}
			if active.ID() != input.ExpectedActiveSchemaID ||
				input.Version != active.Version()+1 {
				return ErrProfileHTTPConflict
			}
			publication, planned, planErr := application.backfills.Originate(
				scoped, scope, active, schema, evidence, input.MigrationMappings,
			)
			if planErr != nil {
				return planErr
			}
			if planned {
				view = profileBackfillMutationView(publication)
				return nil
			}
			result, publishErr := application.commands.Publish(
				scoped, scope, profileapp.PublishCommand{
					Change: profileapp.SchemaChange{
						Schema: schema, ExpectedActiveSchemaID: active.ID(),
					},
					Evidence: evidence,
				},
			)
			if publishErr == nil {
				view = profileMutationView(result)
			}
			return publishErr
		},
	)
	return view, err
}
