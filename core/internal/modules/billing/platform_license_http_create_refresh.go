package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

const platformLicenseRefreshDigestDomain = "harpy/platform-license-refresh/v1"

func (handler *PlatformLicenseHTTPHandler) createRefresh(
	w http.ResponseWriter,
	r *http.Request,
) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	access, err := handler.accessFor(
		r, environment, "platform.license.refresh",
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	key, valid := platformLicenseIdempotencyKey(
		r.Header.Values("Idempotency-Key"),
	)
	if !valid {
		handler.writeFailure(w, r, platformLicenseRequestInvalid)
		return
	}
	digest := platformLicenseRefreshDigest(environment, access)
	environmentID, environmentErr := uuid.Parse(environment.EnvironmentID)
	tenantID, tenantErr := uuid.Parse(environment.TenantID)
	actorID, actorErr := uuid.Parse(access.ActorID)
	if environmentErr != nil || tenantErr != nil || actorErr != nil {
		handler.writeFailure(w, r, ErrPlatformLicenseHTTPUnavailable)
		return
	}
	value, err := handler.application.Create(
		r.Context(),
		platformlicenseapp.RefreshCommand{
			EnvironmentID: environmentID,
			TenantID:      tenantID, ActorID: actorID,
			IdempotencyKey: key, RequestSHA256: digest,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	location := "/api/platform/license/refresh-operations/" +
		value.OperationID.String()
	w.Header().Set("Location", location)
	handler.writeSuccess(w, r, http.StatusAccepted,
		platformLicenseRefreshAcceptedView{
			OperationID: value.OperationID.String(), State: "accepted",
		})
}

func platformLicenseRefreshDigest(
	environment PlatformLicenseHTTPEnvironment,
	access PlatformLicenseHTTPAccess,
) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(platformLicenseRefreshDigestDomain))
	_, _ = hash.Write([]byte(environment.EnvironmentID))
	_, _ = hash.Write([]byte(environment.TenantID))
	_, _ = hash.Write([]byte(access.ActorID))
	return hex.EncodeToString(hash.Sum(nil))
}
