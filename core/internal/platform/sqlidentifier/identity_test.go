package sqlidentifier_test

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestIdentityIsExactOpaqueAndDomainAware(t *testing.T) {
	token, _ := sqlidentifier.MintDynamic("rel_example")
	if !sqlidentifier.Matches(token, "rel_example") ||
		sqlidentifier.Matches(token, "rel_other") {
		t.Fatal("opaque token exact match contract failed")
	}
	first, err := sqlidentifier.Fingerprint(token)
	second, secondErr := sqlidentifier.Fingerprint(token)
	other, _ := sqlidentifier.MintDynamic("rel_other")
	third, thirdErr := sqlidentifier.Fingerprint(other)
	if err != nil || secondErr != nil || thirdErr != nil ||
		first == [32]byte{} || first != second || first == third {
		t.Fatal("opaque token fingerprint contract failed")
	}
}

func TestIdentityRejectsZeroToken(t *testing.T) {
	if sqlidentifier.Matches(sqlidentifier.Token{}, "rel_example") {
		t.Fatal("zero token matched")
	}
	if _, err := sqlidentifier.Fingerprint(sqlidentifier.Token{}); err == nil {
		t.Fatal("zero token fingerprinted")
	}
}
