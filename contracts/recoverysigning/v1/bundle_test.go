package recoverysigningv1

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBuildAndParsePurposeSeparatedBundle(t *testing.T) {
	created := time.Date(2026, 7, 22, 10, 0, 0, 123000, time.UTC)
	bundle, err := Build("018f0000-0000-7000-8000-000000000001", created,
		fixtureMaterial())
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(bundle.CanonicalBytes())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.SHA256() != bundle.SHA256() || parsed.Generation() != 1 ||
		!parsed.GeneratedAt().Equal(created) {
		t.Fatal("recovery signing bundle round trip drifted")
	}
	for _, purpose := range orderedPurposes {
		if _, found := parsed.PublicKey(purpose); !found || Filename(purpose) == "" {
			t.Fatalf("purpose %q is missing", purpose)
		}
	}
}

func TestParseRejectsTamperFormattingUnknownAndPurposeReuse(t *testing.T) {
	bundle, err := Build("018f0000-0000-7000-8000-000000000001",
		time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC), fixtureMaterial())
	if err != nil {
		t.Fatal(err)
	}
	valid := bundle.CanonicalBytes()
	tests := [][]byte{
		append([]byte(" "), valid...),
		bytes.Replace(valid, []byte(`"generation":1`), []byte(`"generation":2`), 1),
		bytes.Replace(valid, []byte(`"schema":`), []byte(`"unknown":true,"schema":`), 1),
		bytes.Replace(valid, []byte(PurposeRestoreReceipt), []byte(PurposeRecoveryControl), 1),
		[]byte(strings.Repeat("x", MaximumBundleBytes+1)),
	}
	for _, document := range tests {
		if _, err := Parse(document); err == nil {
			t.Fatal("invalid recovery signing bundle accepted")
		}
	}
}

func fixtureMaterial() []PublicMaterial {
	return []PublicMaterial{
		{Purpose: PurposeRecoveryManifest, PublicKey: [32]byte{1}},
		{Purpose: PurposeRecoveryControl, PublicKey: [32]byte{2}},
		{Purpose: PurposeRestoreReceipt, PublicKey: [32]byte{3}},
	}
}
