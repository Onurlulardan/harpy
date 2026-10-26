package usage

import (
	usageadoptionpg "github.com/Onurlulardan/harpy/core/internal/modules/usage/adapters/postgres/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/google/uuid"
)

type AdoptionCoordinator = adoption.Coordinator
type AdoptionSource = adoption.Source
type AdoptionTarget = adoption.Target
type AdoptionClock = adoption.Clock
type AdoptionExecuteRequest = adoption.ExecuteRequest
type AdoptionRun = adoption.Run
type AdoptionManifest = adoption.Manifest
type AdoptionSourceKind = adoption.SourceKind
type AdoptionPostgresDatabase = usageadoptionpg.Database
type AdoptionPostgresRunner = usageadoptionpg.TransactionRunner
type AdoptionPostgresStore = usageadoptionpg.Store
type AdoptionLegacyRow = usageadoptionpg.Row
type AdoptionLegacyRows = usageadoptionpg.Rows
type AdoptionLegacyDatabase = usageadoptionpg.LegacyDatabase
type AdoptionLegacyRunner = usageadoptionpg.LegacyTransactionRunner
type AdoptionLegacySourceImplementation = usageadoptionpg.LegacySource
type AdoptionZeroSourceImplementation = adoption.Zero

const (
	AdoptionZeroSource  = adoption.ZeroSource
	AdoptionLegacyTrace = adoption.LegacyTrace
)

func NewAdoptionCoordinator(source AdoptionSource, target AdoptionTarget,
	clock AdoptionClock,
) (*AdoptionCoordinator, error) {
	return adoption.NewCoordinator(source, target, clock)
}

func NewAdoptionPostgresStore(run AdoptionPostgresRunner) (*AdoptionPostgresStore, error) {
	return usageadoptionpg.NewStoreForAdopter(run)
}

func NewAdoptionLegacySourceForProfile(run AdoptionLegacyRunner, schema, relation string,
	environmentID, defaultTenant uuid.UUID,
) (*AdoptionLegacySourceImplementation, error) {
	return usageadoptionpg.NewLegacySourceForProfile(
		run, schema, relation, environmentID, defaultTenant)
}

func NewAdoptionZeroSource(snapshotID string) (*AdoptionZeroSourceImplementation, error) {
	return adoption.NewZeroSource(snapshotID)
}
