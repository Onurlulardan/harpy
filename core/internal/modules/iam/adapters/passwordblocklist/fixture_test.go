package passwordblocklist

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func blocklistFixture(
	t *testing.T,
	kind blocklistv1.Kind,
	candidates ...string,
) (string, [32]byte) {
	t.Helper()
	records := make([][32]byte, 0, len(candidates))
	for _, candidate := range candidates {
		records = append(records, sha256.Sum256([]byte(candidate)))
	}
	content, err := blocklistv1.Encode(kind, records)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), kind.String()+"-v1.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	clear(content)
	return path, digest
}
