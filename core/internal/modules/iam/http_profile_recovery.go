package iam

import "net/http"

func (handler *ProfileHTTP) recoverBackfill(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errProfileAuthentication)
		return
	}
	resource, resourceValid := profilePathResource(r)
	parentID, parentValid := profilePathUUID(r, "plan_id")
	key, keyValid := identityIdempotencyKey(r)
	if !resourceValid || !parentValid || !keyValid ||
		r.ContentLength != 0 || len(r.TransferEncoding) != 0 ||
		len(r.Header.Values("Content-Type")) != 0 ||
		len(r.Header.Values("Content-Encoding")) != 0 {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	view, err := handler.application.RecoverBackfill(
		r.Context(), claims, ProfileBackfillRecoveryInput{
			IdempotencyKey: key, Resource: resource, ParentPlanID: parentID,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeRecovery(w, r, http.StatusCreated, view)
}
