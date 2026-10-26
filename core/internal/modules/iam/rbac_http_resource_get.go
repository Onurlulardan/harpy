package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) GetResource(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (RBACResourceView, error) {
	if id == uuid.Nil {
		return RBACResourceView{}, ErrRBACHTTPInvalid
	}
	value, err := application.readResource(ctx, claims, id)
	if err != nil {
		return RBACResourceView{}, err
	}
	view := rbacResourceView(value)
	if !validRBACResourceView(view) {
		return RBACResourceView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
