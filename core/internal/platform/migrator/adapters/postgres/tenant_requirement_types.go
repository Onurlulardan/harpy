package postgres

import "time"

const TenantRequirementCatalogVersion = "tenant-migration-requirement-catalog.v2"

type TenantMigrationRequirement struct {
	Ordinal      int    `json:"ordinal"`
	Owner        string `json:"owner"`
	Version      int64  `json:"logical_version"`
	Origin       string `json:"database_origin"`
	SelectedPath string `json:"selected_path"`
	SelectedHash string `json:"selected_sha256"`
}

type TenantRequirementTarget struct {
	Ordinal    int    `json:"ordinal"`
	TenantID   string `json:"tenant_id"`
	SchemaName string `json:"schema_name"`
}

type TenantRequirementCatalog struct {
	ContractVersion  string
	MigrationRunID   string
	ReleaseID        string
	EnvironmentID    string
	DatabaseOrigin   string
	MigrationHash    string
	CatalogHash      string
	RequirementCount int
	TargetCount      int
	Canonical        []byte
	Requirements     []TenantMigrationRequirement
	Targets          []TenantRequirementTarget
	RecordedAt       time.Time
}
