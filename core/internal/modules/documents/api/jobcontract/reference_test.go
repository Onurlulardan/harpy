package jobcontract

import (
	"testing"
)

func TestReferencesRoundTripWithCanonicalPayloadDigest(t *testing.T) {
	tenant := mustID(t, "018f0000-0000-7000-8000-000000000101")
	resource := mustID(t, "018f0000-0000-7000-8000-000000000102")
	tests := []struct {
		build func(ID, ID) (string, string, error)
		parse func(string) (Reference, error)
	}{
		{Archive, ParseArchive}, {Deletion, ParseDeletion},
	}
	for _, test := range tests {
		reference, digest, err := test.build(tenant, resource)
		if err != nil || len(digest) != 64 {
			t.Fatalf("build = %q, %q, %v", reference, digest, err)
		}
		parsed, err := test.parse(reference)
		if err != nil || parsed.TenantID != tenant || parsed.ResourceID != resource {
			t.Fatalf("parse = %#v, %v", parsed, err)
		}
	}
	reference, digest, err := PurgePage(tenant, resource, resource)
	parsed, parseErr := ParsePurge(reference)
	if err != nil || parseErr != nil || len(digest) != 64 ||
		parsed.ResourceID != resource || parsed.AfterDeletionID != resource {
		t.Fatalf("purge page = %q %q %#v, %v %v", reference, digest, parsed, err, parseErr)
	}
	uploadReference, uploadDigest, err := UploadReconcile(tenant, resource, 7)
	uploadParsed, uploadParseErr := ParseUploadReconcile(uploadReference)
	if err != nil || uploadParseErr != nil || len(uploadDigest) != 64 ||
		uploadParsed.ResourceID != resource || uploadParsed.Generation != 7 {
		t.Fatalf("upload reconcile = %q %q %#v, %v %v", uploadReference,
			uploadDigest, uploadParsed, err, uploadParseErr)
	}
}

func mustID(t *testing.T, value string) ID {
	t.Helper()
	id, err := ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
