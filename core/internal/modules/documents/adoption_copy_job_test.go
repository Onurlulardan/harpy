package documents

import (
	"strings"
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
)

func TestDocumentWorkerProcessorsCoverReleaseJobs(t *testing.T) {
	policies := asyncmanifestv1.DocumentsJobPolicies()
	covered := map[string]bool{}
	for _, kind := range documentWorkerProcessorKinds() {
		covered[kind] = true
	}
	if len(covered) != len(policies) {
		t.Fatalf("processor kinds = %d, catalog jobs = %d", len(covered), len(policies))
	}
	for _, policy := range policies {
		if !covered[policy.Kind] {
			t.Fatalf("release job %s has no Documents worker processor", policy.Kind)
		}
	}
}

func TestAdoptionCopyReferenceRoundTrip(t *testing.T) {
	tenant := jobcontract.ID(jobTestTenant)
	run := jobcontract.ID(jobTestResource)
	version := jobcontract.ID(jobTestID)
	reference, digest, err := jobcontract.AdoptionCopy(tenant, run, version)
	if err != nil || len(digest) != 64 {
		t.Fatal(err)
	}
	parsed, err := jobcontract.ParseAdoptionCopy(reference)
	if err != nil || parsed.TenantID != tenant || parsed.RunID != run ||
		parsed.SourceVersionID != version {
		t.Fatalf("parsed = %#v, %v", parsed, err)
	}
	if jobcontract.AdoptionCopyEffectKey(run, version) !=
		"documents:adoption-copy:"+run.String()+":"+version.String() {
		t.Fatal("adoption copy effect key drifted")
	}
}

func TestMissingDocumentWorkerProcessorNamesTheKind(t *testing.T) {
	missing := missingDocumentWorkerProcessors(nil)
	if !strings.Contains(missing, "documents.adoption.copy") ||
		!strings.Contains(missing, "documents.archive.build") {
		t.Fatalf("missing processors = %s", missing)
	}
}
