package asyncmanifestv1

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

func TestWorkflowInstanceChangedProducerMatchesSchema(t *testing.T) {
	t.Parallel()
	document, err := os.ReadFile("../../events/v1/workflow-instance-changed-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	wantDigest := fmt.Sprintf("%x", sha256.Sum256(document))
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := manifest.Producers().Resolve(WorkflowInstanceChangedEventType, 1)
	if err != nil || policy.OwnerDomain != "workflow" ||
		policy.PayloadSchemaSHA256 != wantDigest ||
		policy.PayloadSchemaSHA256 != WorkflowInstanceChangedSchemaSHA256 {
		t.Fatalf("workflow producer = %+v: %v", policy, err)
	}
}
