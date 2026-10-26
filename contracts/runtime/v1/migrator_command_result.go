package runtimev1

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

const MigratorCommandResultVersion = "migrator-command-result.v1"

var (
	migratorOwnerPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	migratorPathPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9_./-]*\.sql$`)
)

type MigratorCommandResult struct {
	ContractVersion string                 `json:"contract_version"`
	Success         bool                   `json:"success"`
	Command         string                 `json:"command"`
	Plan            *MigrationPlanResult   `json:"plan,omitempty"`
	Verify          *MigrationVerifyResult `json:"verify,omitempty"`
	Up              *MigrationUpResult     `json:"up,omitempty"`
}

type MigrationPlanResult struct {
	CurrentVersion int64                `json:"current_version"`
	Pending        []MigrationPlanEntry `json:"pending"`
}

type MigrationPlanEntry struct {
	Owner                        string `json:"owner"`
	Scope                        string `json:"scope"`
	Version                      int64  `json:"version"`
	TransactionMode              string `json:"transaction_mode"`
	Phase                        string `json:"phase"`
	TargetDatabaseGeneration     int64  `json:"target_database_generation"`
	NextMinimumRuntimeGeneration int64  `json:"next_minimum_runtime_generation"`
	Origin                       string `json:"origin"`
	Path                         string `json:"path"`
	SHA256                       string `json:"sha256"`
}

type MigrationVerifyResult struct {
	DatabaseVersion          int64  `json:"database_version"`
	DatabaseGeneration       int64  `json:"database_generation"`
	MinimumRuntimeGeneration int64  `json:"minimum_runtime_generation"`
	VerifiedRunID            string `json:"verified_run_id"`
}

type MigrationUpResult struct {
	RunID           string  `json:"run_id"`
	AppliedVersions []int64 `json:"applied_versions"`
	DatabaseVersion int64   `json:"database_version"`
}

func (result MigratorCommandResult) Validate() error {
	if result.ContractVersion != MigratorCommandResultVersion || !result.Success {
		return fmt.Errorf("migrator result identity is invalid")
	}
	branches := boolCount(result.Plan != nil, result.Verify != nil, result.Up != nil)
	if branches != 1 {
		return fmt.Errorf("migrator result must carry one command branch")
	}
	switch result.Command {
	case "plan":
		return validatePlanResult(result)
	case "verify":
		return validateVerifyResult(result)
	case "up":
		return validateUpResult(result)
	default:
		return fmt.Errorf("migrator result command is invalid")
	}
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func validMigratorUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validMigratorDigest(value string) bool {
	return validateDigest(value, "sha256") == nil
}
