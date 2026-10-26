package config

import (
	"os"
	"time"
)

type Worker struct {
	longRunning          LongRunning
	deploymentMode       DeploymentMode
	authority            WorkerAuthority
	memoryLimit          WorkerMemoryLimit
	rabbitMQ             RabbitMQ
	iam                  IAMWorker
	formsExport          FormsExportArtifacts
	formsExportSpool     AbsolutePath
	formsImportSource    FormsImportSourceArtifacts
	analyticsExports     AnalyticsDashboardExportArtifacts
	analyticsExportSpool AbsolutePath
	restQueryCursorKeys  CursorKeyring
	iamTransferArtifacts IAMSystemTransferArtifacts
	catalogPackageKeys   CatalogPackageKeyring
	documents            DocumentsWorker
	notificationsKeys    NotificationsKEKKeyring
}

func LoadWorker() (Worker, error) {
	return loadWorker(os.Environ())
}

func loadWorker(entries []string) (Worker, error) {
	return loadWorkerAt(entries, time.Now().UTC())
}

func loadWorkerAt(entries []string, now time.Time) (Worker, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileWorker)
	if err != nil {
		return Worker{}, err
	}
	longRunning, err := loadLongRunning(values, runtimeConfig)
	if err != nil {
		return Worker{}, err
	}
	deploymentMode, err := loadDeploymentMode(values)
	if err != nil {
		return Worker{}, err
	}
	authority, err := loadWorkerAuthority(values, runtimeConfig.Environment(), deploymentMode)
	if err != nil {
		return Worker{}, err
	}
	memoryLimit, err := loadWorkerMemoryLimit(values)
	if err != nil {
		return Worker{}, err
	}
	rabbitMQ, err := loadRabbitMQ(values, runtimeConfig.Environment())
	if err != nil {
		return Worker{}, err
	}
	iamConfig, err := loadIAMWorker(values, runtimeConfig.Environment(), now)
	if err != nil {
		return Worker{}, err
	}
	formsExport, err := loadFormsExportArtifacts(values, runtimeConfig.Environment())
	if err != nil {
		return Worker{}, err
	}
	spool, err := loadAbsolutePath(values, "HARPY_FORMS_EXPORT_SPOOL_DIR")
	if err != nil {
		return Worker{}, err
	}
	formsImport, err := loadFormsImportSourceArtifacts(values,
		runtimeConfig.Environment(), formsExport)
	if err != nil {
		return Worker{}, err
	}
	analyticsExports, err := loadAnalyticsDashboardExportArtifacts(
		values, runtimeConfig.Environment(),
	)
	if err != nil {
		return Worker{}, err
	}
	analyticsExportSpool, err := loadAbsolutePath(values, analyticsExportSpoolKey)
	if err != nil {
		return Worker{}, err
	}
	cursorKeys, err := loadCursorKeyring(values)
	if err != nil {
		return Worker{}, err
	}
	iamTransferArtifacts, err := loadIAMSystemTransferArtifacts(values,
		runtimeConfig.Environment(), formsExport, formsImport)
	if err != nil {
		return Worker{}, err
	}
	if err = validateAnalyticsDashboardExportSeparation(
		analyticsExports, formsExport, formsImport, iamTransferArtifacts,
	); err != nil {
		return Worker{}, err
	}
	catalogKeys, err := loadCatalogPackageKeyring(values)
	if err != nil {
		return Worker{}, err
	}
	documents, err := loadDocumentsWorker(values, runtimeConfig.Environment())
	if err != nil {
		return Worker{}, err
	}
	notificationsKeys, err := loadNotificationsKEKKeyring(values)
	if err != nil {
		return Worker{}, err
	}
	return Worker{
		longRunning: longRunning, deploymentMode: deploymentMode,
		authority: authority, memoryLimit: memoryLimit,
		rabbitMQ: rabbitMQ, iam: iamConfig,
		formsExport: formsExport, formsExportSpool: spool,
		formsImportSource: formsImport,
		analyticsExports:  analyticsExports, analyticsExportSpool: analyticsExportSpool,
		restQueryCursorKeys:  cursorKeys,
		iamTransferArtifacts: iamTransferArtifacts,
		catalogPackageKeys:   catalogKeys,
		documents:         documents,
		notificationsKeys: notificationsKeys,
	}, nil
}
