package tenancy

import (
	"net/http"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

func (handler *TenantSettingsHTTP) testTenantEmail(w http.ResponseWriter, r *http.Request) {
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
	var document tenantSettingsTestEmailDocument
	if err := httpinput.DecodeJSON(w, r, tenantSettingsMaximumBodyBytes, &document); err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	access, err := handler.accessFor(r, TenantSettingsPermissionManage,
		TenantSettingsOperationTestEmail)
	if err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	result, err := handler.application.TestEmail(r.Context(), access,
		settings.TestEmailCommand{To: document.To, ExpectedRevision: revision,
			IdempotencyKey: key})
	if err != nil {
		handler.writeTenantSettingsFailure(w, r, err)
		return
	}
	handler.writeTenantSettingsSuccess(w, r, http.StatusAccepted,
		"Test email accepted", tenantSettingsTestEmailView(result))
}
