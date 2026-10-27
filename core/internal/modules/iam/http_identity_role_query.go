package iam

import "net/http"

func (handler *IdentityHTTP) listRoles(w http.ResponseWriter, r *http.Request) {
	handler.executeGETQuery(w, r, IdentityRoles)
}

func (handler *IdentityHTTP) queryRoles(w http.ResponseWriter, r *http.Request) {
	handler.executePOSTQuery(w, r, IdentityRoles)
}
