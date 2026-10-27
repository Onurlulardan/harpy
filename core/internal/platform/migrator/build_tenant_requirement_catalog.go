package migrator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type tenantRequirementCatalogDocument struct {
	ContractVersion string                                     `json:"contract_version"`
	MigrationRunID  string                                     `json:"migration_run_id"`
	ReleaseID       string                                     `json:"release_id"`
	EnvironmentID   string                                     `json:"environment_id"`
	DatabaseOrigin  string                                     `json:"database_origin"`
	MigrationHash   string                                     `json:"migration_manifest_sha256"`
	Requirements    []postgresstore.TenantMigrationRequirement `json:"requirements"`
	Targets         []postgresstore.TenantRequirementTarget    `json:"targets"`
}

var tenantCatalogNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func (runner *Runner) buildTenantRequirementCatalog(
	runID string,
	liveTargets []postgresstore.TenantMigrationTarget,
) (postgresstore.TenantRequirementCatalog, error) {
	if len(runner.tenantSelected) == 0 || len(runner.tenantSelected) > 10000 {
		return postgresstore.TenantRequirementCatalog{},
			fmt.Errorf("tenant requirement selection count is invalid")
	}
	requirements := make([]postgresstore.TenantMigrationRequirement, 0,
		len(runner.tenantSelected))
	for index, selected := range runner.tenantSelected {
		requirements = append(requirements, postgresstore.TenantMigrationRequirement{
			Ordinal: index + 1, Owner: selected.Owner, Version: selected.Version,
			Origin: string(selected.Origin), SelectedPath: selected.Path,
			SelectedHash: selected.SHA256,
		})
	}
	if len(liveTargets) > 10000 {
		return postgresstore.TenantRequirementCatalog{},
			fmt.Errorf("tenant requirement target count is invalid")
	}
	targets := make([]postgresstore.TenantRequirementTarget, len(liveTargets))
	for index, target := range liveTargets {
		if !tenantCatalogNamePattern.MatchString(target.SchemaName) ||
			(index > 0 && liveTargets[index-1].TenantID >= target.TenantID) {
			return postgresstore.TenantRequirementCatalog{},
				fmt.Errorf("tenant requirement target identity is invalid")
		}
		targets[index] = postgresstore.TenantRequirementTarget{
			Ordinal: index + 1, TenantID: target.TenantID,
			SchemaName: target.SchemaName,
		}
	}
	identity := runner.identity
	document := tenantRequirementCatalogDocument{
		ContractVersion: postgresstore.TenantRequirementCatalogVersion,
		MigrationRunID:  runID, ReleaseID: identity.releaseID,
		EnvironmentID: identity.environmentID, DatabaseOrigin: string(identity.databaseOrigin),
		MigrationHash: identity.manifestSHA256, Requirements: requirements,
		Targets: targets,
	}
	canonical, err := attestation.MarshalCanonicalTenantRequirementCatalog(document)
	if err != nil {
		return postgresstore.TenantRequirementCatalog{}, err
	}
	digest := sha256.Sum256(canonical)
	return postgresstore.TenantRequirementCatalog{
		ContractVersion: document.ContractVersion, MigrationRunID: runID,
		ReleaseID: identity.releaseID, EnvironmentID: identity.environmentID,
		DatabaseOrigin: document.DatabaseOrigin, MigrationHash: identity.manifestSHA256,
		CatalogHash: hex.EncodeToString(digest[:]), RequirementCount: len(requirements),
		TargetCount:  len(targets),
		Canonical:    canonical,
		Requirements: requirements,
		Targets:      targets,
	}, nil
}
