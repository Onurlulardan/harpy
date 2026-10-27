package iam

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func (handler *InvitationHTTP) list(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errInvitationAuthentication)
		return
	}
	input, valid := invitationListQuery(r)
	if !valid {
		handler.writeFailure(w, r, ErrInvitationHTTPInvalid)
		return
	}
	view, err := handler.application.List(r.Context(), claims, input)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	next := ""
	if view.Next != uuid.Nil {
		next = view.Next.String()
	}
	handler.writeJSON(w, r, http.StatusOK, invitationListEnvelope{
		Data: view.Items, Total: len(view.Items), NextCursor: next,
	})
}

func invitationListQuery(r *http.Request) (InvitationListInput, bool) {
	if r == nil {
		return InvitationListInput{}, false
	}
	query := r.URL.Query()
	if len(query) > 3 {
		return InvitationListInput{}, false
	}
	input := InvitationListInput{Limit: 100}
	if values := query["limit"]; len(values) > 0 {
		if len(values) != 1 {
			return InvitationListInput{}, false
		}
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit < 1 || limit > 100 {
			return InvitationListInput{}, false
		}
		input.Limit = limit
	}
	if values := query["after"]; len(values) > 0 {
		if len(values) != 1 {
			return InvitationListInput{}, false
		}
		id, err := uuid.Parse(values[0])
		if err != nil || id == uuid.Nil || id.String() != values[0] {
			return InvitationListInput{}, false
		}
		input.After = id
	}
	if values := query["status"]; len(values) > 0 {
		if len(values) != 1 || values[0] == "" {
			return InvitationListInput{}, false
		}
		input.States = strings.Split(values[0], ",")
	}
	return input, true
}
