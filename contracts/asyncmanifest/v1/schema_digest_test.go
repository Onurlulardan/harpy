package asyncmanifestv1

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

func schemaDigest(t *testing.T, path string) string {
	t.Helper()
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	document = bytes.ReplaceAll(document, []byte("\r\n"), []byte("\n"))
	return fmt.Sprintf("%x", sha256.Sum256(document))
}
