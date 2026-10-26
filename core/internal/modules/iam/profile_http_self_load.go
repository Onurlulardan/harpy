package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *profileHTTPApplication) loadSelfProfile(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims accesstoken.Claims,
) (profileapp.SelfRecord, profileapp.Document, profileapp.Schema, error) {
	evidence := profileapp.SelfEvidence{
		ActorID:                      claims.Subject(),
		ExpectedTokenVersion:         claims.TokenVersion(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
	}
	record, err := application.commands.GetSelf(ctx, scope, evidence)
	if err != nil {
		return profileapp.SelfRecord{}, profileapp.Document{}, profileapp.Schema{}, err
	}
	schema, document, err := application.selfWriter.SourceForSelf(
		ctx, scope, claims.Subject(),
	)
	if err != nil {
		return profileapp.SelfRecord{}, profileapp.Document{}, profileapp.Schema{}, err
	}
	return record, document, schema, nil
}
