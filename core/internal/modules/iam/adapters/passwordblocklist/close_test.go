package passwordblocklist

import (
	"context"
	"errors"
	"testing"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func TestCloseWipesLoadedArtifactAndFailsClosed(t *testing.T) {
	path, digest := blocklistFixture(t, blocklistv1.KindCommon, "known password")
	blocklist, err := Open(path, digest, blocklistv1.KindCommon)
	if err != nil {
		t.Fatal(err)
	}
	retained := blocklist.content
	if err := blocklist.Close(); err != nil {
		t.Fatal(err)
	}
	for _, value := range retained {
		if value != 0 {
			t.Fatal("close did not wipe loaded artifact")
		}
	}
	if _, err := blocklist.Blocked(context.Background(), []byte("known password")); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("closed blocklist lookup error=%v", err)
	}
	if err := blocklist.Close(); err != nil {
		t.Fatalf("idempotent close error=%v", err)
	}
}
