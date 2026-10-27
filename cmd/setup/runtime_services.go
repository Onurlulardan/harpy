package main

import "encoding/json"

type fcartKeyDocument struct {
	KID string `json:"kid"`
	Key string `json:"key"`
}

type fcartKeyringDocument struct {
	Version  int                `json:"version"`
	Current  fcartKeyDocument   `json:"current"`
	Retained []fcartKeyDocument `json:"retained"`
}

func writeServiceMaterial(dir string) ([]string, error) {
	lines := make([]string, 0, 32)
	documents, err := writeDocumentsMaterial(dir)
	if err != nil {
		return nil, err
	}
	analytics, err := writeAnalyticsMaterial(dir)
	if err != nil {
		return nil, err
	}
	transfer, err := writeTransferMaterial(dir)
	if err != nil {
		return nil, err
	}
	notifications, err := writeNotificationsMaterial(dir)
	if err != nil {
		return nil, err
	}
	lines = append(lines, documents...)
	lines = append(lines, analytics...)
	lines = append(lines, transfer...)
	return append(lines, notifications...), nil
}

func writeDocumentsMaterial(dir string) ([]string, error) {
	keyring, err := writeFCARTKeyring(dir, "documents-kek-keyring.json", "documents-local", 11)
	if err != nil {
		return nil, err
	}
	lines := []string{
		runtimeLine("HARPY_DOCUMENTS_KEK_KEYRING_FILE", keyring),
		runtimeLine("HARPY_DOCUMENTS_BILLING_ACTOR_ID", "019b1234-5678-7abc-8def-0123456789ab"),
	}
	lines = append(lines, artifactS3Lines("HARPY_DOCUMENTS_LEGACY_SOURCE_S3", "harpy-documents-legacy")...)
	return append(lines, runtimeLine("HARPY_DOCUMENTS_ARCHIVE_RESERVATION_TTL", "24h")), nil
}

func writeAnalyticsMaterial(dir string) ([]string, error) {
	keyring, err := writeFCARTKeyring(dir, "analytics-export-keyring.json", "analytics-export-local", 12)
	if err != nil {
		return nil, err
	}
	spool, err := runtimeDirectory(dir, "analytics-export-spool")
	if err != nil {
		return nil, err
	}
	lines := artifactS3Lines("HARPY_ANALYTICS_DASHBOARD_EXPORT_S3", "harpy-analytics-exports")
	lines = append(lines,
		runtimeLine("HARPY_ANALYTICS_DASHBOARD_EXPORT_FCART_KEK_KEYRING_FILE", keyring),
		runtimeLine("HARPY_ANALYTICS_DASHBOARD_EXPORT_SPOOL_DIR", spool),
	)
	return lines, nil
}

func writeTransferMaterial(dir string) ([]string, error) {
	exportKey, err := writeFCARTKeyring(dir, "iam-export-keyring.json", "iam-export-local", 13)
	if err != nil {
		return nil, err
	}
	importKey, err := writeFCARTKeyring(dir, "iam-import-keyring.json", "iam-import-local", 14)
	if err != nil {
		return nil, err
	}
	spool, err := runtimeDirectory(dir, "iam-transfer-spool")
	if err != nil {
		return nil, err
	}
	lines := []string{runtimeLine("HARPY_IAM_SYSTEM_TRANSFER_SPOOL_DIR", spool)}
	lines = append(lines, artifactS3Lines("HARPY_IAM_SYSTEM_EXPORT_S3", "harpy-iam-exports")...)
	lines = append(lines, runtimeLine("HARPY_IAM_SYSTEM_EXPORT_FCART_KEK_KEYRING_FILE", exportKey))
	lines = append(lines, artifactS3Lines("HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3", "harpy-iam-imports")...)
	return append(lines, runtimeLine("HARPY_IAM_SYSTEM_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE", importKey)), nil
}

func writeFCARTKeyring(dir, name, kid string, marker byte) (string, error) {
	material, err := markedMaterial(marker)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(fcartKeyringDocument{
		Version:  1,
		Current:  fcartKeyDocument{KID: kid, Key: rawKey(material[:])},
		Retained: make([]fcartKeyDocument, 0),
	})
	if err != nil {
		return "", err
	}
	return writeRuntimeFile(dir, name, encoded)
}

func artifactS3Lines(prefix, bucket string) []string {
	return []string{
		runtimeLine(prefix+"_ENDPOINT", "http://127.0.0.1:9000"),
		runtimeLine(prefix+"_REGION", "us-east-1"),
		runtimeLine(prefix+"_BUCKET", bucket),
		runtimeLine(prefix+"_PATH_STYLE", "true"),
		runtimeLine(prefix+"_CREDENTIAL_SOURCE", "workload_identity"),
	}
}
