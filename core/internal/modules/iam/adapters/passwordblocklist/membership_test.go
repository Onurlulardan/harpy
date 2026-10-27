package passwordblocklist

import (
	"context"
	"errors"
	"strings"
	"testing"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func TestBlockedMatchesOnlyCanonicalDigest(t *testing.T) {
	path, digest := blocklistFixture(t, blocklistv1.KindCommon, "Café-password-value")
	blocklist, err := Open(path, digest, blocklistv1.KindCommon)
	if err != nil {
		t.Fatal(err)
	}
	defer blocklist.Close()
	blocked, err := blocklist.Blocked(context.Background(), []byte("Café-password-value"))
	if err != nil || !blocked {
		t.Fatalf("known password blocked=%t error=%v", blocked, err)
	}
	blocked, err = blocklist.Blocked(context.Background(), []byte("unrelated-password-value"))
	if err != nil || blocked {
		t.Fatalf("unknown password blocked=%t error=%v", blocked, err)
	}
}

func TestBlockedFailsClosedWithoutLeakingCandidate(t *testing.T) {
	path, digest := blocklistFixture(t, blocklistv1.KindCompromised, "known password")
	blocklist, err := Open(path, digest, blocklistv1.KindCompromised)
	if err != nil {
		t.Fatal(err)
	}
	decomposed := "Cafe\u0301-password-value"
	if _, err := blocklist.Blocked(context.Background(), []byte(decomposed)); !errors.Is(err, ErrCandidateInvalid) || strings.Contains(err.Error(), decomposed) {
		t.Fatalf("invalid candidate error leaked or changed: %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := blocklist.Blocked(cancelled, []byte("valid-password-value")); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("cancelled lookup error=%v", err)
	}
}
