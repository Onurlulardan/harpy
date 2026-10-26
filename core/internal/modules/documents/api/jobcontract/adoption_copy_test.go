package jobcontract

import (
	"crypto/sha256"
	"testing"
)

func TestAdoptionCopyReferenceRoundTrip(t *testing.T) {
	tenantID := mustID(t, "11111111-1111-4111-8111-111111111111")
	runID := mustID(t, "018f0000-0000-7000-8000-000000000701")
	versionID := mustID(t, "018f0000-0000-7000-8000-000000000703")
	manifest := sha256.Sum256([]byte("documents adoption manifest"))
	reference, digest, err := AdoptionCopy(tenantID, runID, versionID, 17, manifest)
	parsed, parseErr := ParseAdoptionCopy(reference)
	if err != nil || parseErr != nil || len(digest) != 64 ||
		parsed.TenantID != tenantID || parsed.ResourceID != runID ||
		parsed.SourceVersionID != versionID || parsed.CutoverGeneration != 17 ||
		parsed.ManifestSHA256 != manifest {
		t.Fatalf("adoption reference mismatch: %+v %s %v/%v", parsed, digest, err, parseErr)
	}
}
