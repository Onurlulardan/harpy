package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *profileHTTPApplication) GetSelf(
	ctx context.Context,
	claims accesstoken.Claims,
) (ProfileSelfView, error) {
	if application == nil || claims.Subject() == uuid.Nil {
		return ProfileSelfView{}, ErrProfileHTTPInvalid
	}
	request, err := application.snapshotScopeRequest(ctx, claims)
	if err != nil {
		return ProfileSelfView{}, err
	}
	var view ProfileSelfView
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			record, document, schema, readErr := application.loadSelfProfile(
				scoped, scope, claims,
			)
			if readErr != nil {
				return readErr
			}
			view, readErr = newProfileSelfView(record, schema, document)
			return readErr
		},
	)
	return view, profileApplicationError(err)
}
