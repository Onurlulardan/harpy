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
	restQueryCursorKeys  CursorKeyring
	iamTransferArtifacts IAMSystemTransferArtifacts
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
	cursorKeys, err := loadCursorKeyring(values)
	if err != nil {
		return Worker{}, err
	}
	iamTransferArtifacts, err := loadIAMSystemTransferArtifacts(values,
		runtimeConfig.Environment())
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
		restQueryCursorKeys:  cursorKeys,
		iamTransferArtifacts: iamTransferArtifacts,
		documents:            documents,
		notificationsKeys:    notificationsKeys,
	}, nil
}
