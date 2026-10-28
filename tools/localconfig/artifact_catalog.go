package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"
)

const manifestFileName = "manifest.json"

var actorBridgeDomains = []string{
	"analytics", "billing", "catalog", "documents",
	"forms", "notifications", "work", "workflow",
}

var baseArtifactNames = []string{
	"api.env",
	"ai-context-public-keys.json",
	"ai-core-credential",
	"ai-credential-private-key",
	"ai-runtime.json",
	"analytics-dashboard-export-fcart-kek-keyring.json",
	"actor-bridge-client.crt",
	"actor-bridge-client.key",
	"actor-bridge-public-keys.json",
	"actor-bridge-server-ca.crt",
	"actor-bridge-server-ca.key",
	"actor-bridge-signing-seed",
	"actor-bridge-targets.json",
	"catalog-package-public-keyring.json",
	"catalog-package-signing-seed",
	"documents-kek-keyring.json",
	"forms-export-fcart-kek-keyring.json",
	"forms-import-source-fcart-kek-keyring.json",
	"iam-access-token-keyring.json",
	"iam-auth-throttle-keyring.json",
	"iam-csrf-keyring.json",
	"iam-login-directory-keyring.json",
	"iam-one-time-token-keyring.json",
	"iam-password-blocklists.json",
	"iam-password-common.bin",
	"iam-password-compromised.bin",
	"iam-recovery-code-keyring.json",
	"iam-refresh-token-keyring.json",
	"iam-session-network-keyring.json",
	"iam-system-export-fcart-kek-keyring.json",
	"iam-system-import-source-fcart-kek-keyring.json",
	"invitation-template.json",
	"kms-master-key.bin",
	"kms-server.crt",
	"kms-server.key",
	"local-ca.crt",
	"local-ca.key",
	"mailpit-server.crt",
	"mailpit-server.key",
	"mcp-core-credential",
	"mcp-runtime.json",
	"migrator.env",
	"minio.env",
	"minio-root-access-key.txt",
	"minio-root-secret-key.txt",
	"notifications-consumer-keyring.json",
	"notifications-consumer-manifest.json",
	"notifications-consumer-signing-seed",
	"notifications-kek-keyring.json",
	"operator.env",
	"postgres-bootstrap-password.txt",
	"postgres-bootstrap-dsn.txt",
	"postgres-cluster-admin-maintenance-dsn.txt",
	"postgres-cluster-admin-password.txt",
	"postgres-cluster-admin-target-dsn.txt",
	"postgres-extension-admin-dsn.txt",
	"postgres-extension-admin-password.txt",
	"postgres-migrator-dsn.txt",
	"postgres-migrator-password.txt",
	"postgres-runtime-dsn.txt",
	"postgres-runtime-password.txt",
	"postgres-reactivation-observer-password.txt",
	"rabbitmq-password.txt",
	"rabbitmq-url.txt",
	"rabbitmq.conf",
	"rest-query-cursor-keyring.json",
	"s3-access-key-id.txt",
	"s3-secret-access-key.txt",
	"schema-executor.env",
	"smtp-auth.txt",
	"smtp-password.txt",
	"tenant-migration-executor.env",
	"worker.env",
}

var localWorkloadArtifactNames = []string{
	"gcp-subject-token-hmac-key.bin",
	"gcp-subject-token.jwt",
	"gcp-wif.json",
	"local-authority-ack-signing-seed",
	"local-authority-release-signing-seed",
}

var forgeGuardWorkloadArtifactNames = []string{
	"api-gcp-wif.json",
	"api-workload-signing-seed",
	"host-workload-public-keys.json",
	"kms-access-token-signing-key.bin",
	"worker-gcp-wif.json",
	"worker-workload-signing-seed",
}

type artifactSet struct {
	files            map[string][]byte
	expiresAt        time.Time
	actorBridgeKeyID string
	options          generationOptions
	billingActorID   string
}

func newArtifactSet() *artifactSet {
	return newArtifactSetFor(localGenerationOptions())
}

func newArtifactSetFor(options generationOptions) *artifactSet {
	return &artifactSet{files: make(map[string][]byte), options: options}
}

func expectedArtifactNames() []string {
	names := append([]string(nil), baseArtifactNames...)
	names = append(names, localWorkloadArtifactNames...)
	names = append(names, recoveryArtifactNames...)
	for _, domain := range actorBridgeDomains {
		prefix := "actor-bridge-server-" + domain
		names = append(names, prefix+".crt", prefix+".key")
	}
	sort.Strings(names)
	return names
}

func (set *artifactSet) expectedArtifactNames() []string {
	names := expectedArtifactNames()
	if set.options.profile == profileForgeGuard {
		for _, localName := range localWorkloadArtifactNames {
			names = removeArtifactName(names, localName)
		}
		for _, recoveryName := range recoveryArtifactNames {
			names = removeArtifactName(names, recoveryName)
		}
		names = append(names, forgeGuardWorkloadArtifactNames...)
		names = append(names, forgeGuardArtifactNames(set.options.licenseKind)...)
		sort.Strings(names)
	}
	return names
}

func removeArtifactName(names []string, target string) []string {
	result := names[:0]
	for _, name := range names {
		if name != target {
			result = append(result, name)
		}
	}
	return result
}

func (set *artifactSet) add(name string, content []byte) error {
	if name == "" || name != filepath.Base(name) || name == manifestFileName ||
		len(content) == 0 {
		return fmt.Errorf("invalid artifact %q", name)
	}
	if _, exists := set.files[name]; exists {
		return fmt.Errorf("duplicate artifact %q", name)
	}
	set.files[name] = append([]byte(nil), content...)
	return nil
}

func (set *artifactSet) trackExpiry(value time.Time) {
	if value.IsZero() {
		return
	}
	if set.expiresAt.IsZero() || value.Before(set.expiresAt) {
		set.expiresAt = value.UTC().Truncate(time.Second)
	}
}

func (set *artifactSet) requireComplete() error {
	expected := set.expectedArtifactNames()
	if len(set.files) != len(expected) || set.expiresAt.IsZero() {
		return fmt.Errorf("artifact set is incomplete")
	}
	for _, name := range expected {
		if _, exists := set.files[name]; !exists {
			return fmt.Errorf("artifact set is missing %q", name)
		}
	}
	return nil
}
