package tenancy

import (
	"net/http"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const customerProvisioningMaximumBodyBytes int64 = 16 * 1024

func (handler *CustomerProvisioningHTTP) create(w http.ResponseWriter, r *http.Request) {
	key, valid := customerProvisioningIdempotencyKey(r)
	if !valid || r.URL.RawQuery != "" {
		handler.writeCustomerProvisioningFailure(w, r, provisioning.ErrInvalid)
		return
	}
	var document provisioning.CustomerRequest
	if err := httpinput.DecodeJSON(
		w, r, customerProvisioningMaximumBodyBytes, &document,
	); err != nil {
		handler.writeCustomerProvisioningFailure(w, r, err)
		return
	}
	operation, err := handler.application.Create(r.Context(), provisioning.Command{
		Caller: provisioning.CallerHTTP, IdempotencyKey: key,
		Customer: document, Credential: customerProvisioningCredential(r),
	})
	if err != nil {
		handler.writeCustomerProvisioningFailure(w, r, err)
		return
	}
	location := customerProvisioningOperationsPath + operation.OperationID.String()
	w.Header().Set("Location", location)
	handler.writeCustomerProvisioningJSON(w, r, http.StatusAccepted,
		customerProvisioningAcceptedView(operation, location))
}
