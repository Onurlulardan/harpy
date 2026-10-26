package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (handler *HTTPHandler) environmentFor(r *http.Request) (HTTPEnvironment, error) {
	if handler == nil || r == nil {
		return HTTPEnvironment{}, ErrHTTPUnavailable
	}
	environment, err := handler.environment.Current(r.Context())
	if err != nil || !environment.valid() {
		return HTTPEnvironment{}, ErrHTTPUnavailable
	}
	return environment, nil
}

func (handler *HTTPHandler) onPremiseEnvironmentFor(
	r *http.Request,
) (HTTPEnvironment, error) {
	environment, err := handler.environmentFor(r)
	if err != nil {
		return HTTPEnvironment{}, err
	}
	if environment.DeploymentMode != "onpremise" {
		return HTTPEnvironment{}, ErrHTTPForbidden
	}
	return environment, nil
}

func (handler *HTTPHandler) accessFor(
	r *http.Request,
	environment HTTPEnvironment,
	capability string,
	stepUpTarget string,
) (HTTPAccess, error) {
	request := HTTPAccessRequest{
		AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
		Capability:          capability,
		ExpectedMode:        environment.DeploymentMode,
	}
	if stepUpTarget != "" {
		purpose, purposeErr := platformopsapi.StepUpPurpose(
			environment.TenantID,
			capability,
			stepUpTarget,
		)
		if purposeErr != nil {
			return HTTPAccess{}, ErrHTTPUnavailable
		}
		request.RequireStepUp, request.StepUpPurpose = true, purpose
	}
	access, err := handler.access.Verify(r.Context(), request)
	if err != nil {
		return HTTPAccess{}, err
	}
	if !access.valid() || access.TenantID != environment.TenantID {
		return HTTPAccess{}, ErrHTTPForbidden
	}
	return access, nil
}
