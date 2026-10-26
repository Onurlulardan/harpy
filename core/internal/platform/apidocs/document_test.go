package apidocs

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestValidateDocumentRejectsInvalidArtifacts(t *testing.T) {
	tests := map[string]string{
		"empty":           "",
		"malformed":       `{`,
		"trailing":        testDocument + `{}`,
		"wrong version":   `{"openapi":"3.0.3","info":{"title":"API","version":"v1"},"paths":{}}`,
		"missing title":   `{"openapi":"3.1.1","info":{"version":"v1"},"paths":{}}`,
		"blank title":     `{"openapi":"3.1.1","info":{"title":"  ","version":"v1"},"paths":{}}`,
		"missing release": `{"openapi":"3.1.1","info":{"title":"API"},"paths":{}}`,
		"blank release":   `{"openapi":"3.1.1","info":{"title":"API","version":"\t"},"paths":{}}`,
		"missing paths":   `{"openapi":"3.1.1","info":{"title":"API","version":"v1"}}`,
		"null paths":      `{"openapi":"3.1.1","info":{"title":"API","version":"v1"},"paths":null}`,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, err := validateDocument([]byte(raw)); err == nil {
				t.Fatal("invalid artifact passed validation")
			}
		})
	}
}

func TestValidateDocumentEnforcesSizeAndOwnsValidatedBytes(t *testing.T) {
	prefix := `{"openapi":"3.1.1","info":{"title":"API","version":"v1"},"paths":{},"padding":"`
	suffix := `"}`
	paddingBytes := maximumDocumentBytes - len(prefix) - len(suffix)
	raw := []byte(prefix + strings.Repeat("x", paddingBytes) + suffix)
	if len(raw) != maximumDocumentBytes {
		t.Fatalf("fixture size=%d want=%d", len(raw), maximumDocumentBytes)
	}

	digest := sha256.Sum256(raw)
	wantETag := `"` + hex.EncodeToString(digest[:]) + `"`
	validated, etag, err := validateDocument(raw)
	if err != nil {
		t.Fatalf("maximum-size artifact failed validation: %v", err)
	}
	if etag != wantETag {
		t.Fatalf("etag=%q want=%q", etag, wantETag)
	}
	raw[0] = '['
	if validated[0] != '{' {
		t.Fatal("validated artifact aliases caller-owned bytes")
	}
	if _, _, err = validateDocument(append(validated, '\n')); err == nil {
		t.Fatal("oversized artifact passed validation")
	}
}
