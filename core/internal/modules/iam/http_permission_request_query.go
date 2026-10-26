package iam

import (
	"net/http"
	"strconv"
)

func (handler *PermissionRequestHTTP) get(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	id, idValid := identityPathID(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if !idValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.Get(r.Context(), claims, id)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}

func (handler *PermissionRequestHTTP) list(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	input, valid := permissionRequestListInput(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.List(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writePage(w, r, view)
}

func permissionRequestListInput(r *http.Request) (PermissionRequestList, bool) {
	query := r.URL.Query()
	for key := range query {
		if key != "limit" && key != "after_id" {
			return PermissionRequestList{}, false
		}
	}
	input := PermissionRequestList{Limit: 50}
	if values, found := query["limit"]; found {
		if len(values) != 1 || values[0] == "" {
			return PermissionRequestList{}, false
		}
		limit, err := strconv.Atoi(values[0])
		if err != nil {
			return PermissionRequestList{}, false
		}
		input.Limit = limit
	}
	if values, found := query["after_id"]; found {
		if len(values) != 1 {
			return PermissionRequestList{}, false
		}
		id, valid := decodeUUID(values[0])
		if !valid {
			return PermissionRequestList{}, false
		}
		input.After = id
	}
	return input, input.Limit >= 1 && input.Limit <= 100
}
