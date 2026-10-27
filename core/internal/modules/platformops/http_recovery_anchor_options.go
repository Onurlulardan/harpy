package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

func (handler *HTTPHandler) createRecoveryAnchorOptions(
	w http.ResponseWriter,
	r *http.Request,
) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.recovery.anchor.rotate", environment.EnvironmentID)
	}
	key, valid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !valid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	digest, err := recoveryRequestSHA256(struct{}{})
	if err != nil {
		handler.writeFailure(w, r, ErrHTTPUnavailable)
		return
	}
	environmentID := uuid.MustParse(environment.EnvironmentID)
	actorID := uuid.MustParse(access.ActorID)
	options, err := handler.anchors.CreateRecoveryAnchorOptions(
		r.Context(), platformopsapi.RecoveryAnchorOptionsCommand{
			RecoveryAnchorSubject: platformopsapi.RecoveryAnchorSubject{
				EnvironmentID: environment.EnvironmentID,
				TenantID:      environment.TenantID, ActorID: access.ActorID},
			OperationID: recoveryOperationID(environmentID, actorID,
				"anchor-registration", key).String(),
			AnchorID: recoveryOperationID(environmentID, actorID,
				"recovery-anchor", key).String(),
			IdempotencyKey: key, RequestSHA256: digest})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, recoveryAnchorOptionsView(options))
}
