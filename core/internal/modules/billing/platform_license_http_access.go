package billing

import "net/http"

func (handler *PlatformLicenseHTTPHandler) onPremiseEnvironmentFor(
	r *http.Request,
) (PlatformLicenseHTTPEnvironment, error) {
	if handler == nil || r == nil {
		return PlatformLicenseHTTPEnvironment{}, ErrPlatformLicenseHTTPUnavailable
	}
	environment, err := handler.environment.Current(r.Context())
	if err != nil || !environment.valid() {
		return PlatformLicenseHTTPEnvironment{}, ErrPlatformLicenseHTTPUnavailable
	}
	if environment.DeploymentMode != "onpremise" {
		return PlatformLicenseHTTPEnvironment{}, ErrPlatformLicenseHTTPForbidden
	}
	return environment, nil
}

func (handler *PlatformLicenseHTTPHandler) accessFor(
	r *http.Request,
	environment PlatformLicenseHTTPEnvironment,
	capability string,
) (PlatformLicenseHTTPAccess, error) {
	access, err := handler.access.Verify(
		r.Context(),
		PlatformLicenseHTTPAccessRequest{
			AuthorizationValues: append(
				[]string(nil), r.Header.Values("Authorization")...,
			),
			TenantIDValues: append(
				[]string(nil), r.Header.Values("X-Tenant-ID")...,
			),
			TenantSlugValues: append(
				[]string(nil), r.Header.Values("X-Tenant-Slug")...,
			),
			Capability: capability, ExpectedMode: environment.DeploymentMode,
		},
	)
	if err != nil {
		return PlatformLicenseHTTPAccess{}, err
	}
	if !access.valid() || access.TenantID != environment.TenantID {
		return PlatformLicenseHTTPAccess{}, ErrPlatformLicenseHTTPForbidden
	}
	return access, nil
}
