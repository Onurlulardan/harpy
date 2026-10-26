package config

import (
	"os"
	"time"
)

type API struct {
	longRunning           LongRunning
	publicListenAddress   string
	http                  HTTP
	restQueryCursorKeys   CursorKeyring
	analyticsProxies      AnalyticsTrustedProxies
	deploymentMode        DeploymentMode
	authority             APIAuthority
	iam                   IAMAPI
	formsExport           FormsExportArtifacts
	formsImportSource     FormsImportSourceArtifacts
	formsImportStaging    AbsolutePath
	analyticsExports      AnalyticsDashboardExportArtifacts
	iamTransferArtifacts  IAMSystemTransferArtifacts
	catalogPackageKeys    CatalogPackageKeyring
	documents             DocumentsAPI
	notificationConsumers NotificationsConsumerKeyring
}

func LoadAPI() (API, error) {
	return loadAPI(os.Environ())
}

func loadAPI(entries []string) (API, error) {
	return loadAPIAt(entries, time.Now().UTC())
}

func loadAPIAt(entries []string, now time.Time) (API, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileAPI)
	if err != nil {
		return API{}, err
	}
	longRunning, err := loadLongRunning(values, runtimeConfig)
	if err != nil {
		return API{}, err
	}
	publicAddress, publicPort, err := parseListenAddress(values, "HARPY_PUBLIC_LISTEN_ADDR", "0.0.0.0:8000")
	if err != nil {
		return API{}, err
	}
	_, managementPort, err := parseListenAddress(values, "HARPY_MANAGEMENT_LISTEN_ADDR", "0.0.0.0:9000")
	if err != nil {
		return API{}, err
	}
	if publicPort == managementPort {
		return API{}, configError("HARPY_PUBLIC_LISTEN_ADDR", "must use a different port from management")
	}
	httpConfig, err := loadHTTP(values)
	if err != nil {
		return API{}, err
	}
	cursorKeys, err := loadCursorKeyring(values)
	if err != nil {
		return API{}, err
	}
	analyticsProxies, err := loadAnalyticsTrustedProxies(values)
	if err != nil {
		return API{}, err
	}
	deploymentMode, err := loadDeploymentMode(values)
	if err != nil {
		return API{}, err
	}
	authority, err := loadAPIAuthority(values, deploymentMode)
	if err != nil {
		return API{}, err
	}
	iamConfig, err := loadIAMAPI(values, now, runtimeConfig.Environment())
	if err != nil {
		return API{}, err
	}
	formsExport, err := loadFormsExportArtifacts(values, runtimeConfig.Environment())
	if err != nil {
		return API{}, err
	}
	formsImport, err := loadFormsImportSourceArtifacts(values,
		runtimeConfig.Environment(), formsExport)
	if err != nil {
		return API{}, err
	}
	formsImportStaging, err := loadAbsolutePath(values, formsImportSourceStagingKey)
	if err != nil {
		return API{}, err
	}
	analyticsExports, err := loadAnalyticsDashboardExportArtifacts(
		values, runtimeConfig.Environment(),
	)
	if err != nil {
		return API{}, err
	}
	iamTransferArtifacts, err := loadIAMSystemTransferArtifacts(values,
		runtimeConfig.Environment(), formsExport, formsImport)
	if err != nil {
		return API{}, err
	}
	if err = validateAnalyticsDashboardExportSeparation(
		analyticsExports, formsExport, formsImport, iamTransferArtifacts,
	); err != nil {
		return API{}, err
	}
	catalogKeys, err := loadCatalogPackageKeyring(values)
	if err != nil {
		return API{}, err
	}
	documents, err := loadDocumentsAPI(values, runtimeConfig.Environment())
	if err != nil {
		return API{}, err
	}
	notificationConsumers, err := loadNotificationsConsumerKeyring(values)
	if err != nil {
		return API{}, err
	}
	return API{
		longRunning: longRunning, publicListenAddress: publicAddress,
		http: httpConfig, restQueryCursorKeys: cursorKeys,
		analyticsProxies: analyticsProxies,
		deploymentMode:   deploymentMode, authority: authority,
		iam: iamConfig, formsExport: formsExport,
		formsImportSource: formsImport, formsImportStaging: formsImportStaging,
		analyticsExports:      analyticsExports,
		iamTransferArtifacts:  iamTransferArtifacts,
		catalogPackageKeys:    catalogKeys,
		documents:             documents,
		notificationConsumers: notificationConsumers,
	}, nil
}
