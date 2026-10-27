package iam

import "net/http"

func (handler *IdentityHTTP) listOrganizations(w http.ResponseWriter, r *http.Request) {
	handler.executeGETQuery(w, r, IdentityOrganizations)
}

func (handler *IdentityHTTP) queryOrganizations(w http.ResponseWriter, r *http.Request) {
	handler.executePOSTQuery(w, r, IdentityOrganizations)
}
