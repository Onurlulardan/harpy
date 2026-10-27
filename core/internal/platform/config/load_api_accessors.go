package config

func (c API) NotificationsConsumerKeyring() (NotificationsConsumerKeyring, error) {
	return c.notificationConsumers, c.notificationConsumers.Validate()
}

func (c API) Documents() (DocumentsAPI, error) {
	return c.documents, c.documents.Validate()
}

func (c API) CatalogPackageKeyring() (CatalogPackageKeyring, error) {
	return c.catalogPackageKeys, c.catalogPackageKeys.Validate()
}

func (c API) IAM() (IAMAPI, error) {
	if err := c.iam.Validate(); err != nil {
		return IAMAPI{}, err
	}
	return c.iam, nil
}

func (c API) IAMHTTPProfile() (IAMHTTPProfile, error) {
	if err := c.iam.http.Validate(); err != nil {
		return IAMHTTPProfile{}, err
	}
	profile := c.iam.http
	profile.origins = append([]string(nil), profile.origins...)
	return profile, nil
}

func (c API) LongRunning() LongRunning    { return c.longRunning }
func (c API) PublicListenAddress() string { return c.publicListenAddress }
func (c API) HTTP() HTTP                  { return c.http }
func (c API) DeploymentMode() (DeploymentMode, error) {
	return c.deploymentMode, c.deploymentMode.Validate()
}
func (c API) Authority() APIAuthority { return c.authority }
func (c API) RestQueryCursorKeyring() (CursorKeyring, error) {
	if err := c.restQueryCursorKeys.Validate(); err != nil {
		return CursorKeyring{}, err
	}
	return c.restQueryCursorKeys, nil
}

func (c API) AnalyticsTrustedProxies() AnalyticsTrustedProxies {
	return c.analyticsProxies
}

func (c API) AnalyticsDashboardExportArtifacts() (
	AnalyticsDashboardExportArtifacts, error,
) {
	if err := c.analyticsExports.Validate(); err != nil {
		return AnalyticsDashboardExportArtifacts{}, err
	}
	return c.analyticsExports, nil
}

func (c API) IAMSystemTransferArtifacts() (IAMSystemTransferArtifacts, error) {
	if err := c.iamTransferArtifacts.Validate(); err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	return c.iamTransferArtifacts, nil
}
