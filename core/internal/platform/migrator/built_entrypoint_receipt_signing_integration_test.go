//go:build integration

package migrator

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func builtEntrypointReceiptSignature(
	t *testing.T,
	signer attestation.Signer,
	contractVersion string,
	kind string,
	payload any,
	issuedAt time.Time,
) runtimev1.AttestationSignature {
	t.Helper()
	envelope, err := attestation.MarshalCanonical(map[string]any{
		"contract_version": contractVersion, "kind": kind, "payload": payload,
	})
	if err != nil {
		t.Fatalf("encode built entrypoint receipt envelope: %v", err)
	}
	return builtEntrypointAttestationSignature(t, signer, envelope, issuedAt)
}

func builtEntrypointReceiptDatabase(
	live builtEntrypointLiveIdentity,
) builtEntrypointReceiptDatabaseShape {
	return builtEntrypointReceiptDatabaseShape{
		DatabaseName:             builtEntrypointDatabaseName,
		DatabaseSystemIdentifier: live.systemID, DatabaseOID: live.databaseOID,
		DatabaseOrigin: "fresh", DatabaseOwner: runtimev1.DatabaseOwnerRole,
		DatabaseEncoding:       runtimev1.PostgresEncoding,
		DatabaseLocaleProvider: runtimev1.PostgresLocaleProvider,
		DatabaseCollation:      runtimev1.PostgresCollation,
		DatabaseCType:          runtimev1.PostgresCharacterType,
		DataChecksums:          runtimev1.PostgresDataChecksums,
	}
}

type builtEntrypointReceiptDatabaseShape struct {
	DatabaseName             string
	DatabaseSystemIdentifier string
	DatabaseOID              uint32
	DatabaseOrigin           string
	DatabaseOwner            string
	DatabaseEncoding         string
	DatabaseLocaleProvider   string
	DatabaseCollation        string
	DatabaseCType            string
	DataChecksums            bool
}
