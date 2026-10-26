package platformops

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

const permissionCacheFlushDomain = "harpy/iam/permission-cache-flush/v1"

func (handler *HTTPHandler) flushPermissionCache(w http.ResponseWriter, r *http.Request) {
	environment, err := handler.environmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.operations.permission_cache.flush", environment.TenantID)
	}
	key, valid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !valid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	operationID := recoveryOperationID(uuid.MustParse(environment.EnvironmentID),
		uuid.MustParse(access.ActorID), "permission-cache-flush", key)
	digest := sha256.Sum256([]byte(permissionCacheFlushDomain))
	result, err := handler.permissionCache.FlushPermissionCache(
		r.Context(), platformopsapi.PermissionCacheFlush{
			TenantID: environment.TenantID, ActorID: access.ActorID,
			OperationID: operationID.String(), RequestSHA256: hex.EncodeToString(digest[:]),
			TokenVersion:         access.TokenVersion,
			AuthorizationVersion: access.AuthorizationVersion,
			PlatformClock:        access.PlatformClock})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	handler.writeSuccess(w, r, http.StatusOK, struct {
		Strategy           string `json:"strategy"`
		AuthorizationClock int64  `json:"authorization_clock"`
		Revision           int64  `json:"revision"`
		Replayed           bool   `json:"replayed"`
	}{Strategy: "versioned_transactional", AuthorizationClock: result.AuthorizationClock,
		Revision: result.Revision, Replayed: result.Replayed})
}
