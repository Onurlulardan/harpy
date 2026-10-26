package serviceprincipal

import (
	"strings"
	"testing"
)

func TestDecodeEvidenceIsStrictAndTyped(t *testing.T) {
	raw := []byte(strings.Replace(`{
      "principal":{"id":"71000000-0000-4000-8000-000000000001",
        "kind":"workflow","state":"active","authorization_version":3,"revision":2},
      "grant":{"id":"72000000-0000-4000-8000-000000000001","revision":1,
        "previous_revision":0,
        "principal_id":"71000000-0000-4000-8000-000000000001",
        "capability":"forms.record.update","scope_kind":"tenant","owner":"forms",
        "resource_kind":"record","resource_id":null,
        "valid_from":"2026-07-16T00:00:00Z","valid_until":null,"state":"active",
        "manifest_sha256":"$manifest"}}`, "$manifest", strings.Repeat("7a", 32), 1))
	principal, grant, err := decodeEvidence(raw)
	if err != nil {
		t.Fatal(err)
	}
	if principal.AuthorizationVersion != 3 || grant.Revision != 1 ||
		string(grant.Capability) != "forms.record.update" || grant.ManifestSHA256[0] != 0x7a {
		t.Fatalf("decoded evidence drifted: %#v / %#v", principal, grant)
	}
}

func TestDecodeEvidenceRejectsUnknownFields(t *testing.T) {
	if _, _, err := decodeEvidence([]byte(`{"principal":{},"grant":{},"extra":true}`)); err == nil {
		t.Fatal("unknown evidence field was accepted")
	}
}
