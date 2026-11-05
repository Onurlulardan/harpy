package main

import (
	"errors"
	"testing"
)

func TestParseTenantIDAcceptsCanonicalUUID(t *testing.T) {
	const tenantID = "00112233-4455-6677-8899-aabbccddeeff"
	for _, output := range []string{tenantID + "\n", tenantID + "\r\n", tenantID} {
		got, err := parseTenantID([]byte(output))
		if err != nil || got != tenantID {
			t.Fatalf("output=%q id=%q err=%v", output, got, err)
		}
	}
}

func TestParseTenantIDRejectsAnythingElse(t *testing.T) {
	samples := []string{
		"",
		"\n",
		"not-a-uuid\n",
		"00112233-4455-6677-8899-AABBCCDDEEFF\n",
		"00000000-0000-0000-0000-000000000000\n",
		"00112233-4455-6677-8899-aabbccddeeff\nextra\n",
		" 00112233-4455-6677-8899-aabbccddeeff\n",
		"00112233-4455-6677-8899-aabbccddeeff \n",
	}
	for _, sample := range samples {
		got, err := parseTenantID([]byte(sample))
		if !errors.Is(err, errInvalidTenant) || got != "" {
			t.Fatalf("sample=%q id=%q err=%v", sample, got, err)
		}
	}
}
