package runtimev1

import "fmt"

const BinarySchemaContractVersion = "database-schema-contract.v1"

type BinarySchemaContract struct {
	ContractVersion           string `json:"contract_version"`
	Component                 string `json:"component"`
	RuntimeGeneration         int64  `json:"runtime_generation"`
	MinimumDatabaseGeneration int64  `json:"minimum_database_generation"`
	MigrationManifestSHA256   string `json:"migration_manifest_sha256"`
	ExtensionManifestSHA256   string `json:"extension_manifest_sha256"`
	PostgresImageDigest       string `json:"postgres_image_digest"`
}

func (contract BinarySchemaContract) Validate() error {
	if contract.ContractVersion != BinarySchemaContractVersion || contract.Component != "core" {
		return fmt.Errorf("binary schema contract identity is invalid")
	}
	if contract.RuntimeGeneration < 1 || contract.MinimumDatabaseGeneration < 1 {
		return fmt.Errorf("binary schema contract generation is invalid")
	}
	if err := validateDigest(contract.MigrationManifestSHA256, "migration_manifest_sha256"); err != nil {
		return err
	}
	if err := validateDigest(contract.ExtensionManifestSHA256, "extension_manifest_sha256"); err != nil {
		return err
	}
	if contract.PostgresImageDigest != PostgresImageDigest {
		return fmt.Errorf("binary schema contract PostgreSQL image is invalid")
	}
	return nil
}
