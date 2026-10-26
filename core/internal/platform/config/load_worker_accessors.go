package config

func (c Worker) NotificationsKEKKeyring() (NotificationsKEKKeyring, error) {
	return c.notificationsKeys, c.notificationsKeys.Validate()
}

func (c Worker) Documents() (DocumentsWorker, error) {
	return c.documents, c.documents.Validate()
}

func (c Worker) CatalogPackageKeyring() (CatalogPackageKeyring, error) {
	return c.catalogPackageKeys, c.catalogPackageKeys.Validate()
}

func (c Worker) LongRunning() LongRunning {
	return c.longRunning
}

func (c Worker) DeploymentMode() (DeploymentMode, error) {
	return c.deploymentMode, c.deploymentMode.Validate()
}

func (c Worker) Authority() WorkerAuthority { return c.authority }

func (c Worker) RabbitMQ() RabbitMQ { return c.rabbitMQ }

func (c Worker) MemoryLimit() WorkerMemoryLimit { return c.memoryLimit }

func (c Worker) IAM() (IAMWorker, error) {
	if err := c.iam.Validate(); err != nil {
		return IAMWorker{}, err
	}
	return c.iam, nil
}

func (c Worker) FormsExportArtifacts() (FormsExportArtifacts, error) {
	if err := c.formsExport.Validate(); err != nil {
		return FormsExportArtifacts{}, err
	}
	return c.formsExport, nil
}

func (c Worker) FormsExportSpoolDirectory() AbsolutePath {
	return c.formsExportSpool
}

func (c Worker) FormsImportSourceArtifacts() (FormsImportSourceArtifacts, error) {
	if err := c.formsImportSource.Validate(); err != nil {
		return FormsImportSourceArtifacts{}, err
	}
	return c.formsImportSource, nil
}

func (c Worker) AnalyticsDashboardExportArtifacts() (
	AnalyticsDashboardExportArtifacts, error,
) {
	if err := c.analyticsExports.Validate(); err != nil {
		return AnalyticsDashboardExportArtifacts{}, err
	}
	return c.analyticsExports, nil
}

func (c Worker) AnalyticsDashboardExportSpoolDirectory() AbsolutePath {
	return c.analyticsExportSpool
}

func (c Worker) RestQueryCursorKeyring() (CursorKeyring, error) {
	if err := c.restQueryCursorKeys.Validate(); err != nil {
		return CursorKeyring{}, err
	}
	return c.restQueryCursorKeys, nil
}

func (c Worker) IAMSystemTransferArtifacts() (IAMSystemTransferArtifacts, error) {
	if err := c.iamTransferArtifacts.Validate(); err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	return c.iamTransferArtifacts, nil
}
