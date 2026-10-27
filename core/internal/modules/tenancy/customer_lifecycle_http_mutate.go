package tenancy

import (
	"net/http"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const customerLifecycleMaximumBodyBytes int64 = 16 * 1024

type customerLifecycleUpdateDocument struct {
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

func (handler *CustomerLifecycleHTTP) update(w http.ResponseWriter, r *http.Request) {
	handler.applyUpdate(w, r, false)
}

func (handler *CustomerLifecycleHTTP) legacyUpdate(w http.ResponseWriter, r *http.Request) {
	handler.applyUpdate(w, r, true)
}

func (handler *CustomerLifecycleHTTP) applyUpdate(w http.ResponseWriter, r *http.Request,
	legacy bool,
) {
	tenantID, tenantValid := customerLifecycleTenantID(r)
	key, keyValid := customerLifecycleIdempotencyKey(r)
	revision, revisionValid := customerLifecycleExpectedRevision(r)
	if !tenantValid || !keyValid || r.URL.RawQuery != "" {
		handler.writeCustomerLifecycleFailure(w, r, lifecycle.ErrInvalid)
		return
	}
	if !revisionValid {
		handler.writeCustomerLifecycleFailure(w, r, errCustomerLifecyclePrecondition)
		return
	}
	var document customerLifecycleUpdateDocument
	if err := httpinput.DecodeJSON(w, r, customerLifecycleMaximumBodyBytes, &document); err != nil {
		handler.writeCustomerLifecycleFailure(w, r, err)
		return
	}
	if document.Status != "" {
		if legacy {
			w.Header().Set("Link", "</api/platform/tenants/"+tenantID.String()+
				"/suspend>; rel=\"successor-version\"")
		}
		handler.writeCustomerLifecycleFailure(w, r, errCustomerLifecycleStatusMoved)
		return
	}
	tenant, err := handler.application.Update(r.Context(), lifecycle.MutationCommand{
		Credential: customerLifecycleCredential(r), TenantID: tenantID,
		ExpectedRevision: revision, IdempotencyKey: key, Name: document.Name,
	})
	if err != nil {
		handler.writeCustomerLifecycleFailure(w, r, err)
		return
	}
	w.Header().Set("ETag", customerLifecycleETag(tenant.Revision))
	handler.writeCustomerLifecycleTenantSuccess(w, r, tenant, "Tenant updated")
}

func (handler *CustomerLifecycleHTTP) suspend(w http.ResponseWriter, r *http.Request) {
	handler.applyStatus(w, r, lifecycle.MutationSuspend)
}

func (handler *CustomerLifecycleHTTP) reactivate(w http.ResponseWriter, r *http.Request) {
	handler.applyStatus(w, r, lifecycle.MutationReactivate)
}

func (handler *CustomerLifecycleHTTP) applyStatus(w http.ResponseWriter, r *http.Request,
	kind string,
) {
	command, valid, precondition := customerLifecycleMutationCommand(r)
	if !valid {
		failure := error(lifecycle.ErrInvalid)
		if precondition {
			failure = errCustomerLifecyclePrecondition
		}
		handler.writeCustomerLifecycleFailure(w, r, failure)
		return
	}
	var tenant lifecycle.Tenant
	var err error
	if kind == lifecycle.MutationSuspend {
		tenant, err = handler.application.Suspend(r.Context(), command)
	} else {
		tenant, err = handler.application.Reactivate(r.Context(), command)
	}
	if err != nil {
		handler.writeCustomerLifecycleFailure(w, r, err)
		return
	}
	w.Header().Set("ETag", customerLifecycleETag(tenant.Revision))
	message := "Tenant reactivated"
	if kind == lifecycle.MutationSuspend {
		message = "Tenant suspended"
	}
	handler.writeCustomerLifecycleTenantSuccess(w, r, tenant, message)
}

func customerLifecycleMutationCommand(r *http.Request) (lifecycle.MutationCommand, bool, bool) {
	tenantID, tenantValid := customerLifecycleTenantID(r)
	key, keyValid := customerLifecycleIdempotencyKey(r)
	revision, revisionValid := customerLifecycleExpectedRevision(r)
	if !tenantValid || !keyValid || r.URL.RawQuery != "" || !emptyRequestBody(r) {
		return lifecycle.MutationCommand{}, false, false
	}
	if !revisionValid {
		return lifecycle.MutationCommand{}, false, true
	}
	return lifecycle.MutationCommand{Credential: customerLifecycleCredential(r),
		TenantID: tenantID, ExpectedRevision: revision, IdempotencyKey: key}, true, false
}
