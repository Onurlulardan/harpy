package iam

import "net/http"

func (handler *PortalUserHTTP) create(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	key, keyValid := identityIdempotencyKey(r)
	if !valid {
		handler.writeFailure(w, r, portalUserFailureAuthentication)
		return
	}
	if !keyValid {
		handler.writeFailure(w, r, portalUserFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[portalUserCreateDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	subjectID, idValid := decodeUUID(document.SubjectID)
	status := normalizePortalStatus(document.Status)
	if !idValid || !portalSubjectKeyPattern.MatchString(document.SubjectType) ||
		!portalSubjectKeyPattern.MatchString(document.ProfileFormSlug) ||
		(status != "" && status != "PENDING" && status != "INACTIVE") {
		handler.writeFailure(w, r, portalUserFailureInvalid)
		return
	}
	view, err := handler.application.CreatePortalUser(r.Context(), claims,
		PortalUserCreateInput{IdempotencyKey: key, Email: document.Email,
			FirstName: document.FirstName, LastName: document.LastName,
			FormSlug: document.ProfileFormSlug, SubjectType: document.SubjectType,
			SubjectID: subjectID, Status: status})
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeView(w, r, http.StatusCreated, view)
}
