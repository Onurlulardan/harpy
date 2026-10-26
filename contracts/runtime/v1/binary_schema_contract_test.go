package runtimev1

import (
	"strings"
	"testing"
)

func TestBinarySchemaContractValidation(t *testing.T) {
	digest := strings.Repeat("a", 64)
	valid := BinarySchemaContract{
		ContractVersion: BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: digest, ExtensionManifestSHA256: digest,
		PostgresImageDigest: PostgresImageDigest,
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*BinarySchemaContract)
	}{
		{name: "identity", mutate: func(value *BinarySchemaContract) { value.Component = "forms" }},
		{name: "runtime generation", mutate: func(value *BinarySchemaContract) { value.RuntimeGeneration = 0 }},
		{name: "database generation", mutate: func(value *BinarySchemaContract) { value.MinimumDatabaseGeneration = 0 }},
		{name: "migration digest", mutate: func(value *BinarySchemaContract) { value.MigrationManifestSHA256 = "A" + digest[1:] }},
		{name: "extension digest", mutate: func(value *BinarySchemaContract) { value.ExtensionManifestSHA256 = "" }},
		{name: "image", mutate: func(value *BinarySchemaContract) { value.PostgresImageDigest = "sha256:" + digest }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := valid
			test.mutate(&candidate)
			if err := candidate.Validate(); err == nil {
				t.Fatal("invalid binary schema contract accepted")
			}
		})
	}
}
