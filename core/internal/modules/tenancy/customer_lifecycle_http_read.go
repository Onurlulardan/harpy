package tenancy

import (
	"net/http"
	"strconv"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

func (handler *CustomerLifecycleHTTP) list(w http.ResponseWriter, r *http.Request) {
	limit, after, valid := customerLifecyclePage(r)
	if !valid {
		handler.writeCustomerLifecycleFailure(w, r, lifecycle.ErrInvalid)
		return
	}
	page, err := handler.application.List(r.Context(), lifecycle.ListQuery{
		Credential: customerLifecycleCredential(r), Limit: limit, After: after,
	})
	if err != nil {
		handler.writeCustomerLifecycleFailure(w, r, err)
		return
	}
	if customerLifecycleLegacyRequest(r) {
		handler.writeCustomerLifecycleLegacyJSON(w, r, http.StatusOK,
			customerLifecycleLegacyPageView(page))
		return
	}
	handler.writeCustomerLifecycleJSON(w, r, http.StatusOK, customerLifecyclePageView(page))
}

func (handler *CustomerLifecycleHTTP) get(w http.ResponseWriter, r *http.Request) {
	tenantID, valid := customerLifecycleTenantID(r)
	if !valid || r.URL.RawQuery != "" || !emptyRequestBody(r) {
		handler.writeCustomerLifecycleFailure(w, r, lifecycle.ErrInvalid)
		return
	}
	tenant, err := handler.application.Get(r.Context(), lifecycle.GetQuery{
		Credential: customerLifecycleCredential(r), TenantID: tenantID,
	})
	if err != nil {
		handler.writeCustomerLifecycleFailure(w, r, err)
		return
	}
	w.Header().Set("ETag", customerLifecycleETag(tenant.Revision))
	if customerLifecycleLegacyRequest(r) {
		handler.writeCustomerLifecycleLegacyJSON(w, r, http.StatusOK,
			customerLifecycleLegacyTenantView{Data: customerLifecycleTenantView(tenant)})
		return
	}
	handler.writeCustomerLifecycleJSON(w, r, http.StatusOK, customerLifecycleTenantView(tenant))
}

func (handler *CustomerLifecycleHTTP) readOperation(w http.ResponseWriter, r *http.Request) {
	operationID, valid := customerLifecycleOperationID(r)
	if !valid || r.URL.RawQuery != "" || !emptyRequestBody(r) {
		handler.writeCustomerLifecycleFailure(w, r, lifecycle.ErrInvalid)
		return
	}
	operation, err := handler.application.ReadOperation(r.Context(), lifecycle.OperationQuery{
		Credential: customerLifecycleCredential(r), OperationID: operationID,
	})
	if err != nil {
		handler.writeCustomerLifecycleFailure(w, r, err)
		return
	}
	handler.writeCustomerLifecycleJSON(w, r, http.StatusOK,
		customerLifecycleOperationView(operation, ""))
}

func customerLifecyclePage(r *http.Request) (int, uuid.UUID, bool) {
	if r == nil || !emptyRequestBody(r) {
		return 0, uuid.Nil, false
	}
	query := r.URL.Query()
	for key, values := range query {
		if (key != "limit" && key != "after") || len(values) != 1 {
			return 0, uuid.Nil, false
		}
	}
	limit := 0
	if wire := query.Get("limit"); wire != "" {
		parsed, err := strconv.Atoi(wire)
		if err != nil || parsed < 1 || parsed > 100 || strconv.Itoa(parsed) != wire {
			return 0, uuid.Nil, false
		}
		limit = parsed
	}
	after := uuid.Nil
	if wire := query.Get("after"); wire != "" {
		parsed, err := uuid.Parse(wire)
		if err != nil || parsed == uuid.Nil || parsed.String() != wire {
			return 0, uuid.Nil, false
		}
		after = parsed
	}
	return limit, after, true
}
