package passwordblocklist

import (
	"errors"
	"os"
	"testing"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func TestOpenVerifiesContentAddressKindAndBound(t *testing.T) {
	path, digest := blocklistFixture(t, blocklistv1.KindCommon, "known password")
	opened, err := Open(path, digest, blocklistv1.KindCommon)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()

	wrongDigest := digest
	wrongDigest[0] ^= 0xff
	if _, err := Open(path, wrongDigest, blocklistv1.KindCommon); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("content-address mismatch error=%v", err)
	}
	if _, err := Open(path, digest, blocklistv1.KindCompromised); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("cross-kind artifact error=%v", err)
	}

	oversized, _ := blocklistFixture(t, blocklistv1.KindCommon, "other password")
	maximum := int64(16 + blocklistv1.KindCommon.MaximumEntries()*32)
	if err := os.Truncate(oversized, maximum+1); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(oversized, digest, blocklistv1.KindCommon); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("oversized artifact error=%v", err)
	}
}
