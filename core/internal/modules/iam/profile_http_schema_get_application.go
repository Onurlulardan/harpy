package iam

import (
	"context"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *profileHTTPApplication) GetSchema(
	ctx context.Context,
	claims accesstoken.Claims,
	resource profileapp.ResourceKind,
	schemaID uuid.UUID,
) (ProfileSchemaView, error) {
	if !resource.Valid() || schemaID == uuid.Nil {
		return ProfileSchemaView{}, ErrProfileHTTPInvalid
	}
	var view ProfileSchemaView
	err := application.inProfileSnapshot(
		ctx, claims, resource,
		func(scoped context.Context, scope platformtx.TenantScope, _ uuid.UUID) error {
			reader, err := iampostgres.BindProfileReader(scope)
			if err != nil {
				return err
			}
			schema, err := reader.HistoricalSchema(scoped, schemaID)
			if err == nil && schema.Resource() != resource {
				return ErrProfileHTTPNotFound
			}
			if err == nil {
				view, err = newProfileSchemaView(schema)
			}
			return err
		},
	)
	return view, err
}
