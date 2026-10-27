package workhttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/httpdecode"
)

func (handler *Handler) serve(writer http.ResponseWriter, request *http.Request, spec routeSpec) {
	access, err := handler.access.Verify(request.Context(), AccessRequest{
		AuthorizationValues: request.Header.Values("Authorization"),
		TenantIDValues:      request.Header.Values("X-Tenant-ID"),
		TenantSlugValues:    request.Header.Values("X-Tenant-Slug"),
		ConsumerIDValues:    request.Header.Values("X-Harpy-Consumer-ID"),
		BuildIDValues:       request.Header.Values("X-Harpy-Build-ID"),
		Authority:           spec.authority, Operation: spec.operation,
		Compatibility: spec.compatibility,
	})
	if err != nil {
		handler.writeFailure(writer, request, err)
		return
	}
	if !access.ValidFor(spec.authority, spec.operation, spec.compatibility) {
		handler.writeFailure(writer, request, ErrInvalid)
		return
	}
	ids, err := pathIDs(request, spec.pattern)
	if err != nil {
		handler.writeFailure(writer, request, err)
		return
	}
	if spec.compatibility {
		_ = handler.application.RecordCompatibilityHit(request.Context(), CompatibilityHit{
			Access: access, RouteContract: "work-v1-http.v1:" + spec.operation,
			ConsumerID: access.ConsumerID(), BuildID: access.BuildID(),
		})
	}
	if spec.list {
		handler.serveList(writer, request, spec, access, ids)
		return
	}
	if request.Method == http.MethodGet {
		handler.serveGet(writer, request, spec, access, ids)
		return
	}
	handler.serveWrite(writer, request, spec, access, ids)
}

func (handler *Handler) serveList(writer http.ResponseWriter, request *http.Request,
	spec routeSpec, access Access, ids map[string]string,
) {
	query, err := httpdecode.GET().Decode(request.URL.RawQuery)
	var legacy map[string]string
	if spec.compatibility {
		legacy, err = decodeCompatibilityQuery(spec.operation, request.URL.Query())
		if err == nil {
			query, err = httpdecode.GET().Decode("")
		}
	}
	if err != nil {
		handler.writeFailure(writer, request, errRequestInvalid)
		return
	}
	var document []byte
	if query.Validate() == nil {
		document, err = handler.application.List(request.Context(), ListInput{
			Access: access, Operation: spec.operation, Resource: spec.resource,
			IDs: ids, Query: query,
			LegacyQuery: legacy, Compatibility: spec.compatibility,
		})
	} else {
		err = errRequestInvalid
	}
	handler.complete(writer, request, spec, http.StatusOK, document, err)
}

func (handler *Handler) serveGet(writer http.ResponseWriter, request *http.Request,
	spec routeSpec, access Access, ids map[string]string,
) {
	if request.URL.RawQuery != "" {
		handler.writeFailure(writer, request, errRequestInvalid)
		return
	}
	document, err := handler.application.Get(request.Context(), GetInput{
		Access: access, Operation: spec.operation, Resource: spec.resource,
		IDs: ids, Compatibility: spec.compatibility,
	})
	handler.complete(writer, request, spec, http.StatusOK, document, err)
}

func (handler *Handler) serveWrite(writer http.ResponseWriter, request *http.Request,
	spec routeSpec, access Access, ids map[string]string,
) {
	body, err := readDocument(request, spec.compatibility && request.Method == http.MethodDelete)
	var key string
	if err == nil {
		key, err = idempotencyKey(request, spec, access, ids, body)
	}
	var document []byte
	if err == nil {
		document, err = handler.application.Write(request.Context(), WriteInput{
			Access: access, Resource: spec.resource, Operation: spec.operation,
			IDs: ids, Body: body, IdempotencyKey: key, Compatibility: spec.compatibility,
		})
	}
	status := http.StatusOK
	if spec.created {
		status = http.StatusCreated
	}
	handler.complete(writer, request, spec, status, document, err)
}
