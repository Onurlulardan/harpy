package iam

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

func (handler *PostRestoreHTTP) issueRestoredOwnerSession(
	r *http.Request,
	result postrestore.ReactivateResult,
	password []byte,
	origin string,
) (authHTTPGrantResponse, []string, error) {
	if handler == nil || handler.auth == nil || result.TenantSlug == "" ||
		result.Identifier == "" || len(password) == 0 {
		return authHTTPGrantResponse{}, nil, postrestore.ErrUnavailable
	}
	passwordCopy := append([]byte(nil), password...)
	grant, err := handler.auth.application.Login(
		withAuthHTTPPeer(r.Context(), r.RemoteAddr),
		AuthHTTPLoginInput{Mode: AuthHTTPLoginOperator, TenantSlug: result.TenantSlug,
			Identifier: result.Identifier, Password: passwordCopy},
	)
	if err != nil || nilAuthHTTPGrant(grant) {
		if !nilAuthHTTPGrant(grant) {
			grant.Destroy()
		}
		return authHTTPGrantResponse{}, nil, fmt.Errorf(
			"%w: owner session issuance failed", postrestore.ErrUnavailable)
	}
	defer grant.Destroy()
	now, valid := handler.auth.authHTTPNow()
	if !valid {
		return authHTTPGrantResponse{}, nil, postrestore.ErrUnavailable
	}
	response, cookies, err := handler.auth.prepareGrant(r.Context(), grant, origin, now)
	if err != nil {
		return authHTTPGrantResponse{}, nil, fmt.Errorf(
			"%w: owner session presentation failed", postrestore.ErrUnavailable)
	}
	return response, cookies, nil
}
