package config

import (
	"testing"
	"time"
)

func TestIAMActiveIssuanceUsesClosedOpenBoundary(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	access := validIAMAccessKeyringJSON(start.Add(24*time.Hour), 10)
	hmac := validIAMHMACKeyringJSON(start.Add(24*time.Hour), iamPurposeRefresh, 20)
	if _, ok := parseIAMAccessKeyring(Secret{value: access}, start); !ok {
		t.Fatal("access active boundary was not inclusive")
	}
	if _, ok := parseIAMAccessKeyring(
		Secret{value: access}, start.Add(30*24*time.Hour),
	); ok {
		t.Fatal("access sign-until boundary was inclusive")
	}
	if _, ok := parseIAMHMACKeyring(Secret{value: hmac}, iamPurposeRefresh, start); !ok {
		t.Fatal("HMAC active boundary was not inclusive")
	}
	if _, ok := parseIAMHMACKeyring(
		Secret{value: hmac}, iamPurposeRefresh, start.Add(90*24*time.Hour),
	); ok {
		t.Fatal("HMAC issue-until boundary was inclusive")
	}
}

func TestIAMRetainedVerificationMustExtendPastStartup(t *testing.T) {
	now := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	access := iamAccessDocument(now, 10,
		iamRetiredAccessDocument(9, "2026-07-01T00:00:00Z", "2026-07-02T00:00:00Z"))
	if _, ok := parseIAMAccessKeyring(Secret{value: access}, now); ok {
		t.Fatal("access publish-until equal to startup was accepted")
	}
	hmac := iamHMACDocument(now, iamPurposeCSRF, 21,
		iamRetainedHMACDocument(
			20, "2026-04-02T00:00:00Z", "2026-07-01T00:00:00Z",
			`"2026-07-02T00:00:00Z"`,
		),
	)
	if _, ok := parseIAMHMACKeyring(Secret{value: hmac}, iamPurposeCSRF, now); ok {
		t.Fatal("HMAC verify-until equal to startup was accepted")
	}
}
