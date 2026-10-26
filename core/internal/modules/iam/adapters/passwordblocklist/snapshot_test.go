package passwordblocklist

import (
	"context"
	"os"
	"testing"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func TestLoadedSnapshotDoesNotReadMutablePathDuringLookup(t *testing.T) {
	path, digest := blocklistFixture(t, blocklistv1.KindCommon, "known password")
	blocklist, err := Open(path, digest, blocklistv1.KindCommon)
	if err != nil {
		t.Fatal(err)
	}
	defer blocklist.Close()
	if err := os.WriteFile(path, []byte("mutated after activation"), 0o600); err != nil {
		t.Fatal(err)
	}
	blocked, err := blocklist.Blocked(context.Background(), []byte("known password"))
	if err != nil || !blocked {
		t.Fatalf("verified startup snapshot changed: blocked=%t error=%v", blocked, err)
	}
}
