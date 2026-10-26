package tenancy

import (
	"net/http"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func (handler *CustomerProvisioningHTTP) read(w http.ResponseWriter, r *http.Request) {
	operationID, valid := customerProvisioningOperationID(r.PathValue("operation_id"))
	if !valid || r.URL.RawQuery != "" {
		handler.writeCustomerProvisioningFailure(w, r, provisioning.ErrInvalid)
		return
	}
	operation, err := handler.application.Read(r.Context(), provisioning.ReadQuery{
		Caller: provisioning.CallerHTTP, OperationID: operationID,
		Credential: customerProvisioningCredential(r),
	})
	if err != nil {
		handler.writeCustomerProvisioningFailure(w, r, err)
		return
	}
	handler.writeCustomerProvisioningJSON(w, r, http.StatusOK,
		customerProvisioningOperationView(operation))
}
