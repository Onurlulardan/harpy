package jobcontract

import (
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
)

func TestWorkJobReferencesRoundTripAndPinDigest(t *testing.T) {
	tenant := "019b2ee0-12e3-7a1a-8d21-5ca5797f2810"
	target := "019b2ee0-12e4-7a1a-8d21-5ca5797f2811"
	kinds := []string{
		asyncmanifestv1.WorkProjectionRepairJobKind,
		asyncmanifestv1.WorkPurgeExecuteJobKind,
		asyncmanifestv1.WorkRecurrenceMaterializeJobKind,
		asyncmanifestv1.WorkReminderDispatchJobKind,
	}
	for _, kind := range kinds {
		reference, digest, err := Build(kind, tenant, target)
		if err != nil || len(digest) != 64 {
			t.Fatalf("Build(%s) = %q, %q, %v", kind, reference, digest, err)
		}
		parsed, err := Parse(kind, reference)
		if err != nil || parsed.TenantID != tenant || parsed.TargetID != target {
			t.Fatalf("Parse(%s) = %#v, %v", kind, parsed, err)
		}
		rebuilt, rebuiltDigest, err := Build(kind, parsed.TenantID, parsed.TargetID)
		if err != nil || rebuilt != reference || rebuiltDigest != digest {
			t.Fatalf("rebuilt %s = %q, %q, %v", kind, rebuilt, rebuiltDigest, err)
		}
	}
}

func TestWorkJobReferenceRejectsKindConfusion(t *testing.T) {
	tenant := "019b2ee0-12e3-7a1a-8d21-5ca5797f2810"
	target := "019b2ee0-12e4-7a1a-8d21-5ca5797f2811"
	reference, _, err := Build(asyncmanifestv1.WorkPurgeExecuteJobKind, tenant, target)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse(asyncmanifestv1.WorkProjectionRepairJobKind, reference); err == nil {
		t.Fatal("kind-confused reference was accepted")
	}
}
