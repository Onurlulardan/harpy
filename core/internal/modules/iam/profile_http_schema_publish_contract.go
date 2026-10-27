package iam

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func profilePublishRequestSHA(input ProfileSchemaPublish) ([32]byte, error) {
	return identityRequestSHA(struct {
		SchemaID               uuid.UUID
		Resource               profileapp.ResourceKind
		Version                int64
		ExpectedActiveSchemaID uuid.UUID
		Fields                 []ProfileFieldInput
		MigrationMappings      []ProfileMigrationMappingInput
	}{
		input.SchemaID, input.Resource, input.Version,
		input.ExpectedActiveSchemaID, input.Fields, input.MigrationMappings,
	})
}

func profileMutationEvidence(
	claims accesstoken.Claims,
	permissionID uuid.UUID,
	operationID uuid.UUID,
	requestSHA [32]byte,
) profileapp.MutationEvidence {
	return profileapp.MutationEvidence{
		OperationID: operationID, RequestSHA256: requestSHA,
		ActorID: claims.Subject(), ManagePermissionID: permissionID,
		ExpectedTokenVersion:         claims.TokenVersion(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
	}
}

func profileMutationView(result profileapp.MutationResult) ProfileSchemaMutationView {
	return ProfileSchemaMutationView{
		ID: result.ID, Revision: result.Revision,
		AuthorizationClock: result.AuthorizationClock, Replayed: result.Replayed,
	}
}

func profileBackfillMutationView(
	result ProfileBackfillPublicationResult,
) ProfileSchemaMutationView {
	view := ProfileSchemaMutationView{
		ID: result.SchemaID, Revision: result.Revision,
		AuthorizationClock: result.AuthorizationClock, Replayed: result.Replayed,
	}
	if result.PlanID != uuid.Nil {
		planID := result.PlanID
		view.BackfillPlanID = &planID
		view.PublicationState = "planned"
	}
	return view
}
