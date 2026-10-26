package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) Query(
	ctx context.Context,
	claims accesstoken.Claims,
	resource IdentityResource,
	request queryapp.Request,
) (IdentityQueryView, error) {
	return application.queryWithEvidence(ctx, claims, resource, request)
}

func (application *identityHTTPApplication) queryWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	resource IdentityResource,
	request queryapp.Request,
) (IdentityQueryView, error) {
	var view IdentityQueryView
	err := application.inSnapshot(
		ctx, evidence, "restquery.execute", resource,
		func(scoped context.Context, scope platformtx.TenantScope, _ uuid.UUID) error {
			result, err := application.queries.Execute(
				scoped, scope, string(resource), request,
			)
			if err != nil {
				return err
			}
			view = IdentityQueryView{
				Items: result.Items(), HasMore: result.HasMore(),
				NextCursor: result.NextCursor(), NextTransport: result.NextTransport(),
				ExactCount: result.ExactCount(),
			}
			return nil
		},
	)
	return view, err
}
