package main

import (
	"bytes"
	"testing"
)

func TestGeneratedArtifactKEKKeyringsUseRuntimeCanonicalOrder(t *testing.T) {
	set, _ := generatedSetForTest(t)
	prefix := []byte(`{"version":1,"current":`)
	suffix := []byte(`,"retained":[]}`)
	for _, target := range artifactKEKTargets {
		document := set.files[target.file]
		if !bytes.HasPrefix(document, prefix) || !bytes.HasSuffix(document, suffix) {
			t.Fatalf("%s does not use runtime canonical field order", target.file)
		}
	}
}
