package billing

import (
	"net/http"

	"github.com/google/uuid"
)

func (handler *PlatformLicenseHTTPHandler) readLicense(
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
	environmentID, err := uuid.Parse(environment.EnvironmentID)
	if err != nil {
		handler.writeFailure(w, r, ErrPlatformLicenseHTTPUnavailable)
		return
	}
	value, err := handler.application.Read(r.Context(), environmentID)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, platformLicenseView{
		LicenseID: value.LicenseID.String(), EnvironmentID: value.EnvironmentID.String(),
		SnapshotSequence: value.SnapshotSequence, SnapshotSHA256: value.SnapshotSHA256,
		State: value.State, ReleaseChannel: value.ReleaseChannel,
		Features: value.Features, Limits: value.Limits, DatabaseNow: value.DatabaseNow,
	})
}
