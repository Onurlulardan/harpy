package main

import (
	"path/filepath"
	"strconv"
	"strings"
)

func harpyEnvLines(root string, in answers) ([]string, error) {
	mode, err := deploymentMode(in.Profile)
	if err != nil {
		return nil, err
	}
	prev, err := loadPrev(root, in.Profile)
	if err != nil {
		return nil, err
	}
	listen, err := resolveManagementListenAddress(
		"local", requestedManagementListen, prev["HARPY_MANAGEMENT_LISTEN_ADDR"],
	)
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
		"HARPY_MANAGEMENT_LISTEN_ADDR=" + listen,
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
	runtime, err := runtimeMaterial(root, in.Profile)
	if err != nil {
		return nil, err
	}
	release, err := releaseMaterial(root, in.Profile)
	if err != nil {
		return nil, err
	}
	return mergeEnvLines(lines, runtime, release), nil
}

func mergeEnvLines(groups ...[]string) []string {
	index := map[string]int{}
	merged := make([]string, 0)
	for _, group := range groups {
		for _, line := range group {
			key, _, ok := strings.Cut(line, "=")
			if !ok || key == "" {
				merged = append(merged, line)
				continue
			}
			if at, found := index[key]; found {
				merged[at] = line
				continue
			}
			index[key] = len(merged)
			merged = append(merged, line)
		}
	}
	return merged
}

func documentsCredentialFiles(in answers, secrets string) (string, string) {
	if in.S3Provider == "rustfs" {
		return filepath.Join(secrets, "rustfs_access_key"), filepath.Join(secrets, "rustfs_secret_key")
	}
	return filepath.Join(secrets, "s3_access_key"), filepath.Join(secrets, "s3_secret_key")
}
