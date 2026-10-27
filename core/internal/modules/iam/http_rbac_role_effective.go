package iam

import (
	"net/http"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
)

func (handler *RBACHTTP) roleEffective(w http.ResponseWriter, r *http.Request) {
	handler.executeEffective(w, r, rbacdomain.AssignmentRole)
}
