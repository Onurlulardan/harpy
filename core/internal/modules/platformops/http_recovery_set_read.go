package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (handler *HTTPHandler) listRecoverySets(w http.ResponseWriter, r *http.Request) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err == nil {
		_, err = handler.accessFor(r, environment, "platform.recovery_sets.read", "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	sets, err := handler.recoverySets.ListRecoverySets(r.Context(), environment.EnvironmentID)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	items := make([]recoverySetView, len(sets))
	for index, set := range sets {
		items[index] = representRecoverySet(set)
	}
	handler.writeSuccess(w, r, http.StatusOK, struct {
		Items []recoverySetView `json:"items"`
	}{Items: items})
}

func (handler *HTTPHandler) getRecoverySet(w http.ResponseWriter, r *http.Request) {
	recoverySetID := r.PathValue("recovery_set_id")
	if !canonicalHTTPUUID(recoverySetID) {
		handler.writeFailure(w, r, platformopsapi.ErrInvalid)
		return
	}
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err == nil {
		_, err = handler.accessFor(r, environment, "platform.recovery_sets.read", "")
	}
	var set platformopsapi.RecoverySet
	if err == nil {
		set, err = handler.recoverySets.GetRecoverySet(r.Context(), recoverySetID)
	}
	if err == nil && set.EnvironmentID != environment.EnvironmentID {
		err = platformopsapi.ErrNotFound
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, representRecoverySet(set))
}
