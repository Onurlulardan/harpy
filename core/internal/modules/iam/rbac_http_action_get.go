package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) GetAction(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (RBACActionView, error) {
	if id == uuid.Nil {
		return RBACActionView{}, ErrRBACHTTPInvalid
	}
	value, err := application.readAction(ctx, claims, id)
	if err != nil {
		return RBACActionView{}, err
	}
	view := rbacActionView(value)
	if !validRBACActionView(view) {
		return RBACActionView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
