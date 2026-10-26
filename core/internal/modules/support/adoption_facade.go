package support

import (
	postgresadoption "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/postgres/adoption"
	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
	"github.com/google/uuid"
)

type AdoptionSourceKind = application.SourceKind
type AdoptionVerifiedLocator = application.VerifiedLocator
type AdoptionSourceManifest = application.SourceManifest
type AdoptionContactSourceRow = application.ContactSourceRow
type AdoptionBugSourceRow = application.BugSourceRow
type AdoptionContactPage = application.ContactPage
type AdoptionBugPage = application.BugPage
type AdoptionSource = application.Source
type AdoptionTarget = application.Target
type AdoptionRun = application.Run
type AdoptionRunStatus = application.RunStatus
type AdoptionStartRequest = application.StartRequest
type AdoptionProcessRequest = application.ProcessRequest
type AdoptionClock = application.Clock
type AdoptionWorker = application.Worker
type AdoptionDatabase = postgresadoption.Database
type AdoptionTransactionMode = postgresadoption.TransactionMode
type AdoptionTransactionRunner = postgresadoption.TransactionRunner

const (
	AdoptionContactGlobal = application.ContactGlobal
	AdoptionBugTenant     = application.BugTenant
	AdoptionRunPlanned    = application.RunPlanned
	AdoptionRunApplying   = application.RunApplying
	AdoptionRunCompleted  = application.RunCompleted
)

func NewAdoptionWorker(source AdoptionSource, target AdoptionTarget,
	clock AdoptionClock,
) (*AdoptionWorker, error) {
	return application.NewWorker(source, target, clock)
}

func NewAdoptionPostgresTarget(run AdoptionTransactionRunner) (AdoptionTarget, error) {
	return postgresadoption.NewStoreForAdopter(run)
}

func NewAdoptionVerifiedLocator(kind AdoptionSourceKind, tenantID uuid.UUID,
	schema, relation string, relationOID uint32, tenantSlug, tenantName string,
	attestationSHA256 [32]byte,
) (AdoptionVerifiedLocator, error) {
	return application.NewVerifiedLocator(kind, tenantID, schema, relation, relationOID,
		tenantSlug, tenantName, attestationSHA256)
}

func AdoptionEmptySourceRoot() [32]byte { return application.EmptySourceRoot() }

func AdoptionFoldSHA256(previous, page [32]byte) [32]byte {
	return application.FoldSHA256(previous, page)
}

func AdoptionContactPageSHA256(rows []AdoptionContactSourceRow) ([32]byte, error) {
	return application.ContactSourcePageSHA256(rows)
}

func AdoptionBugPageSHA256(rows []AdoptionBugSourceRow) ([32]byte, error) {
	return application.BugSourcePageSHA256(rows)
}
