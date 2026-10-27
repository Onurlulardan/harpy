package demomaterialization

import (
	"context"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const ProvisionOperation = "catalog.operation.resume"

type Materialization struct {
	ModuleSlug     string
	SourceContract string
	Document       []byte
	SemanticSHA256 [32]byte
	FormCount      int
	RowCount       int
}

func (value Materialization) Clone() Materialization {
	value.Document = append([]byte(nil), value.Document...)
	return value
}

type Projection struct {
	ProposedReceiptID uuid.UUID
	PackageID         uuid.UUID
	PackageSlug       string
	PackageVersion    string
	ReleaseID         uuid.UUID
	Component         contract.Component
	Materialization   Materialization
	IdempotencyKey    string
	RequestSHA256     [32]byte
	OccurredAt        time.Time
}

type Snapshot struct {
	ReceiptID       uuid.UUID
	PackageID       uuid.UUID
	PackageSlug     string
	PackageVersion  string
	ReleaseID       uuid.UUID
	Component       contract.Component
	Materialization Materialization
	IdempotencyKey  string
	RequestSHA256   [32]byte
	State           string
	Ready           bool
	RetryAt         time.Time
	CompletedAt     time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, error)
}

type Store interface {
	Project(context.Context, Projection) (Snapshot, error)
}
