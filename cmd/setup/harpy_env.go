package main

import (
	"path/filepath"
	"strconv"
)

const localManagementListenAddress = "0.0.0.0:9090"

func harpyEnvLines(root string, in answers) ([]string, error) {
	mode, err := deploymentMode(in.Profile)
	if err != nil {
		return nil, err
	}
	dsn, err := postgresDSN(in)
	if err != nil {
		return nil, err
	}
	rabbit, err := rabbitURL(in)
	if err != nil {
		return nil, err
	}
	secrets := filepath.Join(root, "deploy", in.Profile, "secrets")
	lines := []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN=" + dsn,
		"HARPY_LOG_LEVEL=info",
		"DEPLOYMENT_MODE=" + mode,
		"HARPY_PUBLIC_LISTEN_ADDR=127.0.0.1:" + strconv.Itoa(in.HTTPPort),
		"HARPY_MANAGEMENT_LISTEN_ADDR=" + localManagementListenAddress,
		"HARPY_RABBITMQ_URL=" + rabbit,
		"GOMEMLIMIT=512MiB",
		"HARPY_INITIAL_ADMIN_PASSWORD_FILE=" + filepath.Join(secrets, "initial_admin_password"),
	}
	if in.S3PathStyle {
		access, secret := documentsCredentialFiles(in, secrets)
		lines = append(lines,
			"HARPY_DOCUMENTS_S3_ENDPOINT="+in.S3Endpoint,
			"HARPY_DOCUMENTS_S3_REGION="+in.S3Region,
			"HARPY_DOCUMENTS_S3_BUCKET="+in.S3Bucket,
			"HARPY_DOCUMENTS_S3_PATH_STYLE=true",
			"HARPY_DOCUMENTS_S3_CREDENTIAL_SOURCE=static_files",
			"HARPY_DOCUMENTS_S3_ACCESS_KEY_ID_FILE="+access,
			"HARPY_DOCUMENTS_S3_SECRET_ACCESS_KEY_FILE="+secret,
		)
	}
	return lines, nil
}

func documentsCredentialFiles(in answers, secrets string) (string, string) {
	if in.S3Provider == "rustfs" {
		return filepath.Join(secrets, "rustfs_access_key"), filepath.Join(secrets, "rustfs_secret_key")
	}
	return filepath.Join(secrets, "s3_access_key"), filepath.Join(secrets, "s3_secret_key")
}
