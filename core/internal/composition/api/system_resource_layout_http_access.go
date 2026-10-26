package api

import (
	"net/http"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
)

func (handler *systemResourceLayoutHTTPHandler) admissionFor(
	r *http.Request,
	resource layouts.Resource,
	action, operation string,
) (layouts.Admission, error) {
	if handler == nil || nilSystemResourceLayoutHTTPDependency(handler.access) || r == nil {
		return layouts.Admission{}, layouts.ErrLayoutAccessAuthentication
	}
	decision, err := handler.access.Verify(r.Context(), layouts.LayoutAccessRequest{
		AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
		Resource:            resource, Action: action, Operation: operation,
	})
	if err != nil {
		return layouts.Admission{}, err
	}
	admission, err := decision.Admission()
	if err != nil {
		return layouts.Admission{}, layouts.ErrLayoutAccessUnavailable
	}
	return admission, nil
}
