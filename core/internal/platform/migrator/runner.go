package migrator

import (
	"database/sql"
	"io/fs"
	"log/slog"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const (
	platformMigrationScope   = "platform"
	platformMigrationLockKey = "harpy:v2:migrate:platform"
	runLifecycleReadyVersion = int64(20260713013200)
	requiredCleanupTimeout   = 5 * time.Second
)

type Runner struct {
	database         *sql.DB
	platformFiles    fs.FS
	selected         []postgresv1.SelectedMigration
	historyAliases   map[int64]historySupersession
	tenantDatabase   func() (*sql.DB, error)
	tenantFiles      fs.FS
	tenantSelected   []postgresv1.SelectedMigration
	tenantAliases    map[int64]historySupersession
	identity         runIdentity
	evidence         foundationEvidence
	schemaAuthority  schemaContractAuthority
	releaseFinalizer ReleaseFinalizer
	cleanupTimeout   time.Duration
	logger           *slog.Logger
}

type UpRequest struct {
	RunID string
}

type UpResult struct {
	RunID           string
	AppliedVersions []int64
	DatabaseVersion int64
}

type PlanResult struct {
	CurrentVersion int64
	Pending        []postgresv1.SelectedMigration
}

type VerifyResult struct {
	DatabaseVersion          int64
	DatabaseGeneration       int64
	MinimumRuntimeGeneration int64
	VerifiedRunID            string
}
