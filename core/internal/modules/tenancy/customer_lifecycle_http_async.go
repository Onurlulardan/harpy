package tenancy

import (
	"net/http"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

func (handler *CustomerLifecycleHTTP) offboard(w http.ResponseWriter, r *http.Request) {
	handler.requestLifecycleOperation(w, r, lifecycle.OperationOffboard)
}

func (handler *CustomerLifecycleHTTP) purge(w http.ResponseWriter, r *http.Request) {
	handler.requestLifecycleOperation(w, r, lifecycle.OperationPurge)
}

func (handler *CustomerLifecycleHTTP) requestLifecycleOperation(w http.ResponseWriter,
	r *http.Request, kind string,
) {
	mutation, valid, precondition := customerLifecycleMutationCommand(r)
	if !valid {
		failure := error(lifecycle.ErrInvalid)
		if precondition {
			failure = errCustomerLifecyclePrecondition
		}
		handler.writeCustomerLifecycleFailure(w, r, failure)
		return
	}
	command := lifecycle.AsyncCommand{Credential: mutation.Credential,
		TenantID: mutation.TenantID, ExpectedRevision: mutation.ExpectedRevision,
		IdempotencyKey: mutation.IdempotencyKey}
	var operation lifecycle.Operation
	var err error
	if kind == lifecycle.OperationOffboard {
		operation, err = handler.application.RequestOffboard(r.Context(), command)
	} else {
		operation, err = handler.application.RequestPurge(r.Context(), command)
	}
	if err != nil {
		handler.writeCustomerLifecycleFailure(w, r, err)
		return
	}
	location := customerLifecycleOperationsPath + operation.ID.String()
	w.Header().Set("Location", location)
	handler.writeCustomerLifecycleJSON(w, r, http.StatusAccepted,
		customerLifecycleOperationView(operation, location))
}
