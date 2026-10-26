package tenancy

import (
	"io"
	"net/http"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

func (handler *TenantSettingsHTTP) getTenantSettings(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.URL.RawQuery != "" || !tenantSettingsBodyAbsent(r) {
		handler.writeTenantSettingsFailure(w, r, settings.ErrInvalid)
		return
	}
	access, err := handler.accessFor(r, TenantSettingsPermissionRead,
		TenantSettingsOperationGet)
	if err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	snapshot, err := handler.application.Get(r.Context(), access)
	if err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	w.Header().Set("ETag", tenantSettingsETag(snapshot.Revision))
	handler.writeTenantSettingsSuccess(w, r, http.StatusOK, "", tenantSettingsView(snapshot))
}

func tenantSettingsBodyAbsent(r *http.Request) bool {
	if r == nil || r.ContentLength > 0 || len(r.TransferEncoding) > 0 ||
		len(r.Header.Values("Content-Type")) > 0 || r.Body == nil {
		return false
	}
	if r.Body == http.NoBody {
		return true
	}
	contents, err := io.ReadAll(io.LimitReader(r.Body, 1))
	return err == nil && len(contents) == 0
}
