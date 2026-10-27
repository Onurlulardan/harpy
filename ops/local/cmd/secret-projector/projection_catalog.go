package main

import "io/fs"

const (
	runtimeUserID = 65532
	rabbitMQUID   = 100
	rabbitMQGID   = 101
)

type projectedFile struct {
	name string
	mode fs.FileMode
}

type targetSpec struct {
	name  string
	uid   int
	gid   int
	dir   fs.FileMode
	files []projectedFile
}

func files(names ...string) []projectedFile {
	result := make([]projectedFile, 0, len(names))
	for _, name := range names {
		result = append(result, projectedFile{name: name, mode: 0o400})
	}
	return result
}

func secretTargets() []targetSpec {
	return []targetSpec{
		{name: "postgres", dir: 0o500, files: files("postgres-bootstrap-password.txt")},
		{name: "rabbitmq", uid: rabbitMQUID, gid: rabbitMQGID,
			dir: 0o500, files: files("rabbitmq.conf")},
		{name: "s3-init", dir: 0o500, files: files(
			"minio-root-access-key.txt", "minio-root-secret-key.txt",
			"s3-access-key-id.txt", "s3-secret-access-key.txt")},
		{name: "recovery-minio", dir: 0o500, files: files(
			"recovery-minio-root-access-key.txt",
			"recovery-minio-root-secret-key.txt")},
		{name: "recovery-s3-init", dir: 0o500, files: files(
			"minio-root-access-key.txt", "minio-root-secret-key.txt",
			"recovery-minio-root-access-key.txt",
			"recovery-minio-root-secret-key.txt", "recovery-s3-users.env")},
		runtimeTarget("recovery-backup",
			"database-backup-executor.env", "postgres-backup-dsn.txt",
			"recovery-storage-manifest.v1.json",
			"recovery-write-credential.json",
			"recovery-retention-delete-credential.json",
			"recovery-documents-source-read-credential.json"),
		runtimeTarget("recovery-restore",
			"database-restore.env", "database-rehydrate.env",
			"postgres-restore-dsn.txt", "recovery-storage-manifest.v1.json",
			"recovery-restore-read-credential.json",
			"recovery-documents-rehydrate-write-credential.json"),
		runtimeTarget("recovery-operator",
			"operator-restore-identity.env", "operator-restore-license.env",
			"postgres-runtime-dsn.txt"),
		runtimeTarget("kms", "kms-master-key.bin", "kms-server.crt", "kms-server.key"),
		runtimeTarget("mailpit", "mailpit-server.crt", "mailpit-server.key", "smtp-auth.txt"),
		runtimeTarget("infra-check", "local-ca.crt"),
		runtimeTarget("local-release", "postgres-bootstrap-dsn.txt",
			"postgres-runtime-password.txt", "postgres-migrator-password.txt",
			"postgres-backup-password.txt", "postgres-restore-password.txt",
			"postgres-reactivation-observer-password.txt",
			"catalog-package-signing-seed",
			"catalog-package-public-keyring.json"),
		runtimeTarget("migrator", "postgres-migrator-dsn.txt",
			"catalog-package-public-keyring.json"),
		runtimeTarget("schema", "postgres-migrator-dsn.txt"),
		runtimeTarget("operator", "postgres-runtime-dsn.txt",
			"iam-login-directory-keyring.json",
			"local-authority-ack-signing-seed",
			"local-authority-release-signing-seed"),
		runtimeTarget("api", apiSecretNames()...),
		runtimeTarget("worker", workerSecretNames()...),
	}
}

func runtimeTarget(name string, names ...string) targetSpec {
	return targetSpec{name: name, uid: runtimeUserID, gid: runtimeUserID,
		dir: 0o500, files: files(names...)}
}

func stateTargets() []targetSpec {
	return []targetSpec{
		{name: "foundation-release", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "release", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "forms-export", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "analytics-dashboard-export", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "forms-import", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "iam-transfer", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "ai-system-knowledge", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "mailpit-data", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "installation", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "dev-authority", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "recovery-signing", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "recovery-key-state", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
		{name: "restore-state", uid: runtimeUserID, gid: runtimeUserID, dir: 0o700},
	}
}
