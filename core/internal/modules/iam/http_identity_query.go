package iam

import (
	"encoding/json"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

func (handler *IdentityHTTP) executeGETQuery(
	w http.ResponseWriter,
	r *http.Request,
	resource IdentityResource,
) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureAuthentication)
		return
	}
	request, err := queryapp.DecodeGET(r.URL.RawQuery)
	if err != nil {
		handler.writeQueryFailure(w, r, err)
		return
	}
	handler.executeQuery(w, r, claims, resource, request)
}

func (handler *IdentityHTTP) executePOSTQuery(
	w http.ResponseWriter,
	r *http.Request,
	resource IdentityResource,
) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureAuthentication)
		return
	}
	var document json.RawMessage
	if err := httpinput.DecodeJSON(
		w, r, identityHTTPBodyLimit, &document,
	); err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	defer clear(document)
	request, err := queryapp.DecodePOST(document)
	if err != nil {
		handler.writeQueryFailure(w, r, err)
		return
	}
	handler.executeQuery(w, r, claims, resource, request)
}

func (handler *IdentityHTTP) executeQuery(
	w http.ResponseWriter,
	r *http.Request,
	claims accesstoken.Claims,
	resource IdentityResource,
	request queryapp.Request,
) {
	view, err := handler.application.Query(r.Context(), claims, resource, request)
	if err != nil {
		handler.writeQueryFailure(w, r, err)
		return
	}
	handler.writeQuery(w, r, view)
}
