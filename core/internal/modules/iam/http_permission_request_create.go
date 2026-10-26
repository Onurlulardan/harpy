package iam

import (
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (handler *PermissionRequestHTTP) create(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	key, valid := identityIdempotencyKey(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[permissionRequestCreateDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, document.ExpiresAt)
	if err != nil {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	input := PermissionRequestCreate{
		Mode: permissionrequest.Mode(document.Mode), ExpiresAt: expiresAt,
		Items:          make([]PermissionRequestCreateItem, len(document.Items)),
		IdempotencyKey: key,
	}
	for index, item := range document.Items {
		targetID, targetValid := decodeUUID(item.TargetID)
		permissionID, permissionValid := decodeUUID(item.PermissionID)
		if !targetValid || !permissionValid {
			handler.writeFailure(w, r, identityFailureInvalid)
			return
		}
		input.Items[index] = PermissionRequestCreateItem{
			TargetKind: permissionrequest.TargetKind(item.TargetKind),
			TargetID:   targetID, PermissionID: permissionID,
		}
	}
	view, err := handler.application.Create(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusCreated, view.Revision, view)
}
