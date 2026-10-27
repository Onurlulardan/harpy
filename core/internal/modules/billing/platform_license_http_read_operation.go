package billing

import (
	"net/http"

	"github.com/google/uuid"
)

func (handler *PlatformLicenseHTTPHandler) readRefreshOperation(
	w http.ResponseWriter,
	r *http.Request,
) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	if _, err = handler.accessFor(
		r, environment, "platform.license.read",
	); err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	operationID, err := uuid.Parse(r.PathValue("operation_id"))
	if err != nil || operationID.Version() != uuid.Version(7) {
		handler.writeFailure(w, r, platformLicenseRequestInvalid)
		return
	}
	environmentID, err := uuid.Parse(environment.EnvironmentID)
	if err != nil {
		handler.writeFailure(w, r, ErrPlatformLicenseHTTPUnavailable)
		return
	}
	value, err := handler.application.Get(
		r.Context(), environmentID, operationID,
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK,
		platformLicenseRefreshOperationView{
			OperationID: value.OperationID.String(), State: value.State,
			SnapshotSequence: value.SnapshotSequence,
			SnapshotSHA256:   value.SnapshotSHA256, FailureCode: value.FailureCode,
			CreatedAt: value.CreatedAt, CompletedAt: value.CompletedAt,
		})
}
