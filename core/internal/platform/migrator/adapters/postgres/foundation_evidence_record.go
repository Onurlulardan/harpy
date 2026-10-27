package postgres

import "time"

type EnvironmentIdentityRecord struct {
	EnvironmentID            string
	DatabaseOrigin           string
	DatabaseSystemID         string
	DatabaseOID              uint32
	BootstrapReleaseID       string
	BootstrapAuthoritySHA256 string
}

type ReceiptEvidenceRecord struct {
	Kind              string
	RunID             string
	EnvironmentID     string
	ReleaseID         string
	DatabaseOrigin    string
	DatabaseSystemID  string
	DatabaseOID       uint32
	OperationDigest   string
	CanonicalDocument []byte
	CanonicalSHA256   string
	SignatureKeyID    string
	IssuedAt          time.Time
}

type ExtensionEvidenceRecord struct {
	Name              string
	Version           string
	SchemaName        string
	OwnerRole         string
	ControlFileSHA256 string
	Capabilities      []string
	DatabaseOrigin    string
	VerifiedRunID     string
}

type CoreSchemaContractRecord struct {
	DatabaseGeneration       int64
	MinimumRuntimeGeneration int64
	MigrationSHA256          string
	ExtensionSHA256          string
	Status                   string
	VerifiedRunID            string
	Revision                 int64
}

type DatabaseGenerationManifestRecord struct {
	EnvironmentID      string
	DatabaseGeneration int64
	MigrationSHA256    string
	ExtensionSHA256    string
}

type ReleaseGenerationRecord struct {
	MigrationRunID            string
	ReleaseID                 string
	EnvironmentID             string
	DatabaseOrigin            string
	DatabaseGeneration        int64
	MinimumRuntimeGeneration  int64
	RuntimeGeneration         int64
	MinimumDatabaseGeneration int64
	MigrationSHA256           string
	ExtensionSHA256           string
	PostgresImageDigest       string
	BootstrapReceiptRunID     string
	ExtensionReceiptRunID     string
	CleanupReceiptRunID       string
}
