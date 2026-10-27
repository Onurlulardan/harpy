package iam

import "net/http"

func (handler *LifecycleHTTP) suspendAuthenticator(w http.ResponseWriter, r *http.Request) {
	handler.identifierRevision(w, r, handler.application.SuspendAuthenticator)
}
