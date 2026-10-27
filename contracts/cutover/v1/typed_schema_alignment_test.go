package cutoverv1

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestTypedCutoverDocumentsMatchPublishedSchemas(t *testing.T) {
	leaseSHA := strings.Repeat("a", 64)
	sealSHA := strings.Repeat("b", 64)
	tests := []struct {
		schema string
		value  any
	}{
		{"credential-lease.schema.json", validTypedCredentialLease()},
		{"credential-lease-signature.schema.json", CredentialLeaseSignature{
			ContractVersion: "cutover-credential-lease-signature.v1",
			ManifestSHA256:  leaseSHA, Signature: validTypedSignature(),
		}},
		{"session-seal-receipt.schema.json", validTypedSessionSeal(leaseSHA)},
		{"credential-cleanup-receipt.schema.json", validTypedCleanup(leaseSHA, sealSHA)},
		{"change-batch.schema.json", validTypedChangeBatch()},
		{"terminal-evidence.schema.json", validTypedTerminalEvidence()},
		{"finalizer-lease.schema.json", validTypedFinalizerLease(leaseSHA, sealSHA)},
		{"finalizer-lease-signature.schema.json", FinalizerLeaseSignature{
			ContractVersion: "cutover-finalizer-lease-signature.v1",
			ManifestSHA256:  leaseSHA, Signature: validTypedSignature(),
		}},
	}
	for _, test := range tests {
		t.Run(test.schema, func(t *testing.T) {
			document, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
			if err != nil {
				t.Fatal(err)
			}
			if err := compileCutoverSchema(t, test.schema).Validate(decoded); err != nil {
				t.Fatalf("typed document violates schema: %v", err)
			}
		})
	}
}
