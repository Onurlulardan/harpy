package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestIAMHMACKeyringAcceptsFiniteAndReceiptRetention(t *testing.T) {
	now := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	finite := iamHMACDocument(now, iamPurposeRefresh, 20,
		iamRetainedHMACDocument(
			19, "2026-04-02T00:00:00Z", "2026-07-01T00:00:00Z",
			`"2026-09-30T00:00:00Z"`,
		),
	)
	if ring, ok := parseIAMHMACKeyring(Secret{value: finite}, iamPurposeRefresh, now); !ok || ring.Validate() != nil {
		t.Fatal("valid finite-retention keyring rejected")
	}
	receiptGated := iamHMACDocument(now, iamPurposeDirectory, 22,
		iamRetainedHMACDocument(
			18, "2026-04-02T00:00:00Z", "2026-07-01T00:00:00Z", "null",
		),
	)
	if _, ok := parseIAMHMACKeyring(Secret{value: receiptGated}, iamPurposeDirectory, now); !ok {
		t.Fatal("valid receipt-gated keyring rejected")
	}
}

func TestIAMHMACKeyringRejectsContractDrift(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	valid := validIAMHMACKeyringJSON(now, iamPurposeRefresh, 20)
	tests := map[string]string{
		"wrong purpose": strings.Replace(valid, `"refresh_token_v1"`, `"csrf_v1"`, 1),
		"duplicate":     strings.Replace(valid, `"version":1`, `"version":1,"version":1`, 1),
		"unknown":       strings.Replace(valid, `"retained":[]`, `"retained":[],"extra":0`, 1),
		"expired":       validIAMHMACKeyringJSON(now.Add(90*24*time.Hour), iamPurposeRefresh, 20),
		"future":        validIAMHMACKeyringJSON(now.Add(-91*24*time.Hour), iamPurposeRefresh, 20),
		"padded material": strings.Replace(
			valid, iamFixtureMaterial(20), iamFixtureMaterial(20)+"=", 1,
		),
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseIAMHMACKeyring(Secret{value: document}, iamPurposeRefresh, now); ok {
				t.Fatal("invalid HMAC keyring accepted")
			}
		})
	}
}

func TestIAMHMACKeyringEnforcesPurposeRetentionMode(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	finiteNull := iamHMACDocument(now, iamPurposeRefresh, 20,
		iamRetainedHMACDocument(19, "2026-04-02T00:00:00Z", "2026-07-01T00:00:00Z", "null"))
	receiptTimestamp := iamHMACDocument(now, iamPurposeDirectory, 22,
		iamRetainedHMACDocument(19, "2026-04-02T00:00:00Z", "2026-07-01T00:00:00Z", `"2026-12-01T00:00:00Z"`))
	shortRetention := iamHMACDocument(now, iamPurposeRefresh, 20,
		iamRetainedHMACDocument(19, "2026-04-02T00:00:00Z", "2026-07-01T00:00:00Z", `"2026-09-29T00:00:00Z"`))
	for _, test := range []struct {
		document string
		purpose  iamHMACPurpose
	}{{finiteNull, iamPurposeRefresh}, {receiptTimestamp, iamPurposeDirectory}, {shortRetention, iamPurposeRefresh}} {
		if _, ok := parseIAMHMACKeyring(Secret{value: test.document}, test.purpose, now); ok {
			t.Fatal("invalid retention mode accepted")
		}
	}
}

func TestIAMHMACKeyringRequiresContiguousUniqueKeys(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	retention := `"2026-12-01T00:00:00Z"`
	tests := []string{
		iamHMACDocument(now, iamPurposeRefresh, 20,
			iamRetainedHMACDocument(19, "2026-04-01T00:00:00Z", "2026-06-30T00:00:00Z", retention)),
		iamHMACDocument(now, iamPurposeRefresh, 20,
			iamRetainedHMACDocument(20, "2026-04-02T00:00:00Z", "2026-07-01T00:00:00Z", retention)),
	}
	for _, document := range tests {
		if _, ok := parseIAMHMACKeyring(Secret{value: document}, iamPurposeRefresh, now); ok {
			t.Fatal("invalid retained HMAC chain accepted")
		}
	}
}

func iamHMACDocument(
	now time.Time,
	purpose iamHMACPurpose,
	marker byte,
	retained ...string,
) string {
	base := validIAMHMACKeyringJSON(now, purpose, marker)
	return strings.Replace(base, `"retained":[]`, `"retained":[`+strings.Join(retained, ",")+`]`, 1)
}

func iamRetainedHMACDocument(
	marker byte,
	activatedAt string,
	issueUntil string,
	verifyUntil string,
) string {
	return fmt.Sprintf(
		`{"kid":%q,"key":%q,"activated_at":%q,"issue_until":%q,"verify_until":%s}`,
		iamFixtureKID(marker), iamFixtureMaterial(marker), activatedAt, issueUntil, verifyUntil,
	)
}
