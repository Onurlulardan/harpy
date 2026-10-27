package migrator

import (
	"testing"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func TestEnvironmentIdentityTreatsReleaseAsImmutableBootstrapEvidence(t *testing.T) {
	bootstrap := postgresstore.EnvironmentIdentityRecord{
		EnvironmentID:  "018f0000-0000-7000-8000-000000000001",
		DatabaseOrigin: "fresh", DatabaseSystemID: "12345", DatabaseOID: 16384,
		BootstrapReleaseID:       "018f0000-0000-7000-8000-000000000002",
		BootstrapAuthoritySHA256: integrationIdentityDigest('a'),
	}
	current := bootstrap
	current.BootstrapReleaseID = "018f0000-0000-7000-8000-000000000003"
	current.BootstrapAuthoritySHA256 = integrationIdentityDigest('b')
	if !sameEnvironmentIdentity(bootstrap, current) {
		t.Fatal("later release was treated as the bootstrap release")
	}
	current = bootstrap
	current.BootstrapAuthoritySHA256 = integrationIdentityDigest('b')
	if sameEnvironmentIdentity(bootstrap, current) {
		t.Fatal("bootstrap authority fork was accepted")
	}
	current = bootstrap
	current.DatabaseOID++
	if sameEnvironmentIdentity(bootstrap, current) {
		t.Fatal("physical database identity fork was accepted")
	}
}

func integrationIdentityDigest(marker byte) string {
	value := make([]byte, 64)
	for index := range value {
		value[index] = marker
	}
	return string(value)
}
