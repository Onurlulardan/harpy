package iam

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func rbacListRequest(r *http.Request, allowAfter bool) (RBACList, bool) {
	input := RBACList{State: "active", Limit: 100}
	if r == nil {
		return RBACList{}, false
	}
	values := r.URL.Query()
	allowed := map[string]bool{"state": true, "limit": true}
	if allowAfter {
		allowed["after_id"] = true
	}
	for key, candidates := range values {
		if !allowed[key] || len(candidates) != 1 || candidates[0] == "" {
			return RBACList{}, false
		}
	}
	if state := values.Get("state"); state != "" {
		input.State = state
	}
	if raw := values.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return RBACList{}, false
		}
		input.Limit = limit
	}
	if raw := values.Get("after_id"); raw != "" {
		id, valid := decodeUUID(raw)
		if !valid {
			return RBACList{}, false
		}
		input.AfterID = id
	}
	return input, rbacListValid(input) && (allowAfter || input.AfterID == uuid.Nil)
}

func rbacEffectiveLimit(r *http.Request) (int, bool) {
	if r == nil {
		return 0, false
	}
	values := r.URL.Query()
	if len(values) == 0 {
		return 100, true
	}
	if len(values) != 1 || len(values["limit"]) != 1 || values.Get("limit") == "" {
		return 0, false
	}
	limit, err := strconv.Atoi(values.Get("limit"))
	return limit, err == nil && limit >= 1 && limit <= 100
}
