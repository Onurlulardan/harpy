package iam

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (handler *PermissionRequestHTTP) listCatalogResources(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, valid := handler.claims(r)
	input, inputValid := permissionRequestCatalogInput(r, false, permissionrequest.CatalogKindAll)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if !inputValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.ListCatalogResources(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeCatalogResourcePage(w, r, view)
}

func (handler *PermissionRequestHTTP) listCatalogActions(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, valid := handler.claims(r)
	input, inputValid := permissionRequestCatalogInput(r, false, permissionrequest.CatalogKindAll)
	_, hasKind := r.URL.Query()["kind"]
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if !inputValid || hasKind || input.Kind != permissionrequest.CatalogKindAll {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.ListCatalogActions(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeCatalogActionPage(w, r, view)
}

func (handler *PermissionRequestHTTP) listMissingCatalogResources(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, valid := handler.claims(r)
	input, inputValid := permissionRequestCatalogInput(r, true, permissionrequest.CatalogKindForm)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if !inputValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "" {
		action = "read"
	}
	view, err := handler.application.ListMissingCatalogResources(
		r.Context(), claims,
		PermissionRequestMissingCatalogList{
			PermissionRequestCatalogList: input, ActionKey: action,
		},
	)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeCatalogResourcePage(w, r, view)
}

func permissionRequestCatalogInput(
	r *http.Request,
	allowAction bool,
	defaultKind permissionrequest.CatalogKind,
) (PermissionRequestCatalogList, bool) {
	if r == nil {
		return PermissionRequestCatalogList{}, false
	}
	query := r.URL.Query()
	allowed := map[string]bool{
		"kind": true, "search": true, "limit": true, "after_id": true,
	}
	if allowAction {
		allowed["action"] = true
	}
	for key, values := range query {
		if !allowed[key] || len(values) != 1 || values[0] == "" {
			return PermissionRequestCatalogList{}, false
		}
	}
	input := PermissionRequestCatalogList{Kind: defaultKind, Limit: 50}
	if raw := query.Get("kind"); raw != "" {
		input.Kind = permissionrequest.CatalogKind(strings.ToLower(strings.TrimSpace(raw)))
	}
	input.Search = strings.TrimSpace(query.Get("search"))
	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return PermissionRequestCatalogList{}, false
		}
		input.Limit = limit
	}
	if raw := query.Get("after_id"); raw != "" {
		id, valid := decodeUUID(raw)
		if !valid {
			return PermissionRequestCatalogList{}, false
		}
		input.After = id
	}
	catalogQuery := permissionRequestCatalogQuery(input)
	if catalogQuery.Validate() != nil {
		return PermissionRequestCatalogList{}, false
	}
	if allowAction {
		action := strings.ToLower(strings.TrimSpace(query.Get("action")))
		if action == "" {
			action = "read"
		}
		if (permissionrequest.MissingCatalogQuery{
			CatalogQuery: catalogQuery, ActionKey: action,
		}).Validate() != nil {
			return PermissionRequestCatalogList{}, false
		}
	}
	return input, true
}
