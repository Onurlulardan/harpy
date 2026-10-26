package iam

import "net/http"

func (handler *LifecycleHTTP) revokeSession(w http.ResponseWriter, r *http.Request) {
	handler.identifierRevision(w, r, handler.application.RevokeSession)
}
