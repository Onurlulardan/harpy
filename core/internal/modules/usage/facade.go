package usage

import (
	usagepostgres "github.com/Onurlulardan/harpy/core/internal/modules/usage/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
)

type Service = application.Service
type Store = application.Store
type Clock = application.Clock
type IDGenerator = application.IDGenerator
type RecordRequest = application.RecordRequest
type RecordResult = application.RecordResult
type Access = application.Access
type AccessInput = application.AccessInput
type DeploymentMode = application.DeploymentMode
type TenantPurpose = application.TenantPurpose
type TimeRangeRequest = application.TimeRangeRequest

var (
	ErrHTTPAuthentication = application.ErrAuthentication
	ErrHTTPForbidden      = application.ErrForbidden
	ErrHTTPUnavailable    = application.ErrUnavailable
)

type PostgresTransactionMode = usagepostgres.TransactionMode
type PostgresRow = usagepostgres.Row
type PostgresDatabase = usagepostgres.Database
type PostgresTransactionRunner = usagepostgres.TransactionRunner
type PostgresStore = usagepostgres.Store

const (
	PostgresCommandTransaction  = usagepostgres.CommandTransaction
	PostgresSnapshotTransaction = usagepostgres.SnapshotTransaction
)

func NewService(store Store, clock Clock, ids IDGenerator) (*Service, error) {
	return application.NewService(store, clock, ids)
}

func NewAccess(input AccessInput) (Access, error) { return application.NewAccess(input) }

func NewPostgresStoreForProfile(run PostgresTransactionRunner,
	rowNotFound func(error) bool,
) (*PostgresStore, error) {
	return usagepostgres.NewStoreForProfile(run, rowNotFound)
}
