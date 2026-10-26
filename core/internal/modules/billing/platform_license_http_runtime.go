package billing

func NewPlatformLicenseHTTPHandler(
	dependencies PlatformLicenseHTTPDependencies,
) (*PlatformLicenseHTTPHandler, error) {
	if nilPort(dependencies.Access) || nilPort(dependencies.Environment) ||
		nilPort(dependencies.Application) || dependencies.Report == nil {
		return nil, ErrPlatformLicenseHTTPUnavailable
	}
	return &PlatformLicenseHTTPHandler{
		access: dependencies.Access, environment: dependencies.Environment,
		application: dependencies.Application, report: dependencies.Report,
	}, nil
}
