package iam

import "net/http"

func (handler *RBACHTTP) updateAction(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	mutation, valid := rbacRevisionedMutation(r)
	if !valid {
		handler.writeFailure(w, r, identityFailurePrecondition)
		return
	}
	document, err := decodeIdentityDocument[rbacCatalogDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	if document.Name == nil && document.Slug == nil && document.Description == nil {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	current, err := handler.application.GetAction(r.Context(), claims, mutation.ID)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	name, slug, description := current.Name, current.Slug, current.Description
	if document.Name != nil {
		name = *document.Name
	}
	if document.Slug != nil {
		slug = *document.Slug
	}
	if document.Description != nil {
		description = *document.Description
	}
	view, err := handler.application.UpdateAction(
		r.Context(), claims, RBACCatalogUpdate{
			RBACMutation: mutation, Key: slug,
			DisplayName: name, Description: description,
		},
	)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
