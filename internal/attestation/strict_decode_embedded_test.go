package attestation

import (
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestDecodeStrictFlattensAnonymousContractMembers(t *testing.T) {
	original := cutoverv1.AuthorityCatalogDependencyReceipt{
		AuthorityCatalogDependency: cutoverv1.AuthorityCatalogDependency{
			Kind: "schema", SchemaName: "platform", ObjectName: "platform",
			Owner: "harpy_owner", DefinitionSHA256: stringDigest('a'),
			ACLSHA256: stringDigest('b'),
		},
		ObjectOID: 2200,
	}
	document, err := MarshalCanonical(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded cutoverv1.AuthorityCatalogDependencyReceipt
	if err := DecodeStrict(document, &decoded); err != nil {
		t.Fatalf("decode embedded authority dependency: %v", err)
	}
	if decoded != original {
		t.Fatalf("decoded dependency = %#v", decoded)
	}
}

func stringDigest(marker byte) string {
	value := make([]byte, 64)
	for index := range value {
		value[index] = marker
	}
	return string(value)
}
