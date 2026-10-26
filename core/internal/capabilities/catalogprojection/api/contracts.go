package catalogprojection

import (
	"context"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const ProvisionOperation = "catalog.operation.resume"

type Materialization struct {
	Kind           string
	ModuleSlug     string
	Identity       string
	Document       []byte
	SemanticSHA256 [32]byte
	ItemCount      int
	AuxiliaryCount int
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
	ReadyAt         time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Decoder interface {
	Decode([]byte, contract.Component) (Materialization, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, error)
}

type Store interface {
	Project(context.Context, Projection) (Snapshot, error)
}
