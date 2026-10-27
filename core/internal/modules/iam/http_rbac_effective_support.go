package iam

import (
	"net/http"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
)

func (handler *RBACHTTP) executeEffective(
	w http.ResponseWriter,
	r *http.Request,
	target rbacdomain.AssignmentTarget,
) {
	claims, valid := handler.claims(r)
	id, idValid := rbacPathID(r)
	limit, limitValid := rbacEffectiveLimit(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if !idValid || !limitValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	page, err := handler.application.Effective(
		r.Context(), claims, RBACEffectiveRead{
			Target: target, TargetID: id, Limit: limit,
		},
	)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writePage(w, r, page.Items, false, nil)
}
