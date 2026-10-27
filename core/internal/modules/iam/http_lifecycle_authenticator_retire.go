package iam

import "net/http"

func (handler *LifecycleHTTP) retireAuthenticator(w http.ResponseWriter, r *http.Request) {
	handler.identifierRevision(w, r, handler.application.RetireAuthenticator)
}
