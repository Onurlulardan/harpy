package iam

import "net/http"

func (handler *IdentityHTTP) listUsers(w http.ResponseWriter, r *http.Request) {
	handler.executeGETQuery(w, r, IdentityUsers)
}

func (handler *IdentityHTTP) queryUsers(w http.ResponseWriter, r *http.Request) {
	handler.executePOSTQuery(w, r, IdentityUsers)
}
