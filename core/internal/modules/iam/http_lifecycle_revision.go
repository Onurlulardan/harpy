package iam

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

type lifecycleHTTPRevisionOperation func(context.Context, accesstoken.Claims,
	LifecycleRevisionInput) error

func (handler *LifecycleHTTP) identifierRevision(w http.ResponseWriter, r *http.Request,
	execute lifecycleHTTPRevisionOperation) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	id, idValid := lifecyclePathUUID(r, "id")
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	if !idValid || !keyValid || !revisionValid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	if err := execute(r.Context(), claims, LifecycleRevisionInput{ID: id,
		ExpectedRevision: revision, IdempotencyKey: key}); err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
