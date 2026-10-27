package iam

import "net/http"

func (handler *RBACHTTP) createAction(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	key, valid := rbacCreateKey(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[rbacCatalogDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	if document.Name == nil {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	description := ""
	if document.Description != nil {
		description = *document.Description
	}
	slug := ""
	if document.Slug != nil {
		slug = *document.Slug
	} else {
		slug = rbacCompatibilitySlug(*document.Name)
	}
	view, err := handler.application.CreateAction(
		r.Context(), claims, RBACCatalogCreate{
			Key: slug, DisplayName: *document.Name,
			Description: description, IdempotencyKey: key,
		},
	)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusCreated, view.Revision, view)
}
