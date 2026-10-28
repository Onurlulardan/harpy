package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestGeneratedConnectionFilesShareOnlyIntendedCredentials(t *testing.T) {
	set, _ := generatedSetForTest(t)
	accessKey := string(set.files["s3-access-key-id.txt"])
	secretKey := string(set.files["s3-secret-access-key.txt"])
	rootAccessKey := string(set.files["minio-root-access-key.txt"])
	rootSecretKey := string(set.files["minio-root-secret-key.txt"])
	if rootAccessKey == accessKey || rootSecretKey == secretKey {
		t.Fatal("MinIO root credentials overlap application S3 credentials")
	}
	wantMinIO := "MINIO_ROOT_USER=" + rootAccessKey + "\n" +
		"MINIO_ROOT_PASSWORD=" + rootSecretKey + "\n"
	if string(set.files["minio.env"]) != wantMinIO {
		t.Fatal("MinIO env file differs from dedicated root credentials")
	}
	smtpPassword := string(set.files["smtp-password.txt"])
	if string(set.files["smtp-auth.txt"]) != "harpy-local:"+smtpPassword+"\n" {
		t.Fatal("Mailpit auth file differs from worker SMTP credential")
	}
	assertDSN(t, set, "postgres-bootstrap-dsn.txt", "postgres-bootstrap-password.txt",
		"local_bootstrap_admin", "postgres")
	assertDSN(t, set, "postgres-runtime-dsn.txt", "postgres-runtime-password.txt",
		"harpy_v2_runtime", "harpy")
	assertDSN(t, set, "postgres-migrator-dsn.txt", "postgres-migrator-password.txt",
		"harpy_migrator", "harpy")
	assertDSN(t, set, "postgres-cluster-admin-maintenance-dsn.txt",
		"postgres-cluster-admin-password.txt", "harpy_cluster_admin", "postgres")
	assertDSN(t, set, "postgres-cluster-admin-target-dsn.txt",
		"postgres-cluster-admin-password.txt", "harpy_cluster_admin", "harpy")
	assertDSN(t, set, "postgres-extension-admin-dsn.txt",
		"postgres-extension-admin-password.txt", "harpy_extension_admin", "harpy")
	rabbit, err := url.Parse(string(set.files["rabbitmq-url.txt"]))
	if err != nil || rabbit.User.Username() != "harpy_worker" ||
		rabbit.Host != "rabbitmq:5672" || rabbit.Path != "/harpy" {
		t.Fatalf("RabbitMQ URL differs: %v", err)
	}
	password, _ := rabbit.User.Password()
	if password != string(set.files["rabbitmq-password.txt"]) {
		t.Fatal("RabbitMQ URL and password file differ")
	}
}

func assertDSN(
	t *testing.T,
	set *artifactSet,
	dsnFile, passwordFile, username, database string,
) {
	t.Helper()
	value, err := url.Parse(string(set.files[dsnFile]))
	if err != nil || value.Scheme != "postgres" || value.Host != "postgres:5432" ||
		strings.TrimPrefix(value.Path, "/") != database || value.User.Username() != username ||
		value.Query().Get("sslmode") != "disable" {
		t.Fatalf("invalid DSN %s: %v", dsnFile, err)
	}
	password, exists := value.User.Password()
	if !exists || password != string(set.files[passwordFile]) {
		t.Fatalf("DSN password differs for %s", dsnFile)
	}
}
