package tenancy

import (
	"net/http"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const tenantSettingsMaximumBodyBytes int64 = 16 * 1024

func (handler *TenantSettingsHTTP) updateTenantSettings(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.URL.RawQuery != "" {
		handler.writeTenantSettingsFailure(w, r, settings.ErrInvalid)
		return
	}
	key, valid := tenantSettingsIdempotencyKey(r)
	if !valid {
		handler.writeTenantSettingsFailure(w, r, settings.ErrInvalid)
		return
	}
	revision, valid := tenantSettingsExpectedRevision(r)
	if !valid {
		handler.writeTenantSettingsFailure(w, r, settings.ErrPrecondition)
		return
	}
	var document tenantSettingsUpdateDocument
	if err := httpinput.DecodeJSON(w, r, tenantSettingsMaximumBodyBytes, &document); err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	if document.SMTP != nil {
		defer document.SMTP.clear()
	}
	if document.Name == nil && document.SMTP == nil {
		handler.writeTenantSettingsFailure(w, r, settings.ErrInvalid)
		return
	}
	smtp, valid := document.SMTP.mutation()
	if !valid {
		handler.writeTenantSettingsFailure(w, r, settings.ErrInvalid)
		return
	}
	if smtp != nil {
		defer clear(smtp.Password)
	}
	access, err := handler.accessFor(r, TenantSettingsPermissionManage,
		TenantSettingsOperationUpdate)
	if err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	result, err := handler.application.Update(r.Context(), access, settings.UpdateCommand{
		Name: document.Name, SMTP: smtp, ExpectedRevision: revision, IdempotencyKey: key})
	if err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	w.Header().Set("ETag", tenantSettingsETag(result.Revision))
	handler.writeTenantSettingsSuccess(w, r, http.StatusOK,
		"Settings updated successfully", tenantSettingsMutationView(result))
}
