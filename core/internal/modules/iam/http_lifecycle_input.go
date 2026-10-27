package iam

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func lifecyclePathUUID(r *http.Request, name string) (uuid.UUID, bool) {
	if r == nil {
		return uuid.Nil, false
	}
	wire := r.PathValue(name)
	id, err := uuid.Parse(wire)
	return id, err == nil && id != uuid.Nil && id.String() == wire
}

func lifecyclePage(r *http.Request) (LifecyclePageInput, bool) {
	if r == nil {
		return LifecyclePageInput{}, false
	}
	query := r.URL.Query()
	if len(query) > 3 {
		return LifecyclePageInput{}, false
	}
	limit := 50
	if values, found := query["limit"]; found {
		if len(values) != 1 {
			return LifecyclePageInput{}, false
		}
		parsed, err := strconv.Atoi(values[0])
		if err != nil || parsed < 1 || parsed > 100 {
			return LifecyclePageInput{}, false
		}
		limit = parsed
	}
	var afterTime time.Time
	var afterID uuid.UUID
	times, hasTime := query["after_time"]
	ids, hasID := query["after_id"]
	if hasTime != hasID {
		return LifecyclePageInput{}, false
	}
	if hasTime {
		if len(times) != 1 || len(ids) != 1 {
			return LifecyclePageInput{}, false
		}
		var err error
		afterTime, err = time.Parse(time.RFC3339Nano, times[0])
		var valid bool
		afterID, valid = decodeUUID(ids[0])
		if err != nil || afterTime.IsZero() || !valid {
			return LifecyclePageInput{}, false
		}
	}
	return LifecyclePageInput{Limit: limit, AfterTime: afterTime, AfterID: afterID}, true
}

func lifecycleAuthenticated(handler *LifecycleHTTP, w http.ResponseWriter,
	r *http.Request) (claimsValid bool) {
	_, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
	}
	return valid
}
