package local

import (
	"os"
	"strings"
	"testing"
)

func TestRecoveryComposeKeepsStorageAndCredentialsIsolated(t *testing.T) {
	content, err := os.ReadFile("compose.recovery.yml")
	if err != nil {
		t.Fatal(err)
	}
	compose := string(content)
	required := []string{
		"  recovery-minio:\n",
		"- recovery-minio-data:/data",
		"S3_OWNER_EXPECTED_BUCKET_OWNER: \"210987654321\"",
		"- local-secret-recovery-backup:/run/secrets:ro",
		"- local-secret-recovery-restore:/run/secrets:ro",
		"- local-secret-recovery-operator:/run/secrets:ro",
		"- local-restore-state:/run/restore-state",
		"entrypoint: [\"/app/core-migrator\", \"database-backup-executor\"]",
		"entrypoint: [\"/app/core-migrator\", \"database-restore\"]",
	}
	for _, value := range required {
		if !strings.Contains(compose, value) {
			t.Fatalf("recovery compose is missing %q", value)
		}
	}
	runtime, err := os.ReadFile("compose.runtime.yml")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(runtime), "local-secret-recovery") ||
		strings.Contains(string(runtime), "recovery-minio-data") {
		t.Fatal("normal API or worker compose received recovery authority")
	}
}

func TestRecoveryTaskHandshakeAndReconcileOrder(t *testing.T) {
	root, err := os.ReadFile("../../Taskfile.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(root),
		"\"local:recovery\": .taskfiles/local-recovery.yml") {
		t.Fatal("local recovery Taskfile namespace is not included")
	}
	content, err := os.ReadFile("../../.taskfiles/local-recovery.yml")
	if err != nil {
		t.Fatal(err)
	}
	document := string(content)
	for _, task := range []string{"restore", "rehydrate"} {
		next := "rehydrate"
		if task == "rehydrate" {
			next = "resume"
		}
		block := taskBlock(t, document, task, next)
		assertOrdered(t, block, []string{
			"reset-operation",
			"up -d --force-recreate database-",
			"wait-ready",
			"frontend:stop",
			"stop core-api core-worker",
			"ack-drain",
			"wait database-",
			"verify-completed",
			"restore-identity-reconcile",
			"record-reactivation",
			"restore-license-reconcile",
			"up -d --wait --wait-timeout 180 core-api",
			"task: :local:frontend",
			"show-reactivation",
		})
	}
	if strings.Contains(document, "task: local:") {
		t.Fatal("included recovery tasks must address root tasks with a leading colon")
	}
}

func TestMainLocalLifecycleIncludesRecoveryVolumes(t *testing.T) {
	content, err := os.ReadFile("../../.taskfiles/local.yml")
	if err != nil {
		t.Fatal(err)
	}
	document := string(content)
	if strings.Count(document, "-f ops/local/compose.recovery.yml") != 2 ||
		!strings.Contains(document, "database-backup-executor") ||
		!strings.Contains(document, "run --rm recovery-s3-init") {
		t.Fatal("main local lifecycle does not own the recovery stack")
	}
}

func assertOrdered(t *testing.T, document string, values []string) {
	t.Helper()
	position := 0
	for _, value := range values {
		next := strings.Index(document[position:], value)
		if next < 0 {
			t.Fatalf("%q is missing or out of order", value)
		}
		position += next + len(value)
	}
}
