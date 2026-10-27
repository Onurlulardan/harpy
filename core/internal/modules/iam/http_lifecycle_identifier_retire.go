package iam

import "net/http"

func (handler *LifecycleHTTP) retireIdentifier(w http.ResponseWriter, r *http.Request) {
	handler.identifierRevision(w, r, handler.application.RetireIdentifier)
}
