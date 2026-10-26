package tenancy

import (
	"io"
	"net/http"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func (handler *CustomerProvisioningHTTP) reissueInitialAdminSetup(
	w http.ResponseWriter,
	r *http.Request,
) {
	tenantID, tenantValid := customerProvisioningOperationID(r.PathValue("tenant_id"))
	key, keyValid := customerProvisioningIdempotencyKey(r)
	if !tenantValid || !keyValid || r.URL.RawQuery != "" || !emptyRequestBody(r) {
		handler.writeCustomerProvisioningFailure(w, r, provisioning.ErrInvalid)
		return
	}
	operation, err := handler.application.ReissueInitialAdminSetup(
		r.Context(), provisioning.ReissueCommand{
			Caller: provisioning.CallerHTTP, TenantID: tenantID,
			IdempotencyKey: key, Credential: customerProvisioningCredential(r),
		},
	)
	if err != nil {
		handler.writeCustomerProvisioningFailure(w, r, err)
		return
	}
	location := customerProvisioningOperationsPath + operation.OperationID.String()
	w.Header().Set("Location", location)
	handler.writeCustomerProvisioningJSON(w, r, http.StatusAccepted,
		customerProvisioningAcceptedView(operation, location))
}

func emptyRequestBody(r *http.Request) bool {
	if r == nil || r.Body == nil {
		return r != nil
	}
	var buffer [1]byte
	read, err := r.Body.Read(buffer[:])
	return read == 0 && err == io.EOF
}
