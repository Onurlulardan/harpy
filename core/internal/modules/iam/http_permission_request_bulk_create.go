package iam

import "net/http"

// bulkCreate exposes the canonical grouped create command on the legacy bulk
// route. The application persists every item through one tenant transaction
// and one idempotent group operation.
func (handler *PermissionRequestHTTP) bulkCreate(
	w http.ResponseWriter,
	r *http.Request,
) {
	handler.create(w, r)
}
