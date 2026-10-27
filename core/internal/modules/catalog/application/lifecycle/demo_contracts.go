package lifecycle

import (
	"context"
	"reflect"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const (
	DemoStateInstalled   = "INSTALLED"
	DemoStateAbsent      = "ABSENT"
	DemoStateNotIncluded = "NOT_INCLUDED"
	DemoActionInstall    = "INSTALL"
	DemoActionRemove     = "REMOVE"
)

type DemoComponent struct {
	InstallationID, PackageID, ComponentID uuid.UUID
	ReleaseID, SourceRunID                 uuid.UUID
	PackageSlug, PackageVersion            string
	ContentSHA256, State                   string
	Document                               []byte
	RecordIDs                              []uuid.UUID
	Revision                               uint64
	UpdatedAt                              time.Time
}

func (value DemoComponent) Clone() DemoComponent {
	value.Document = append([]byte(nil), value.Document...)
	value.RecordIDs = append([]uuid.UUID(nil), value.RecordIDs...)
	return value
}

type DemoStatus struct {
	InstallationID uuid.UUID
	State          string
	Components     []DemoComponent
	UpdatedAt      time.Time
}

type DemoRemovalComponent struct {
	ComponentID, SourceRunID uuid.UUID
	RecordIDs                []uuid.UUID
}

type DemoRemovalPlan struct {
	InstallationID uuid.UUID
	Components     []DemoRemovalComponent
	PlanSHA256     string
}

type DemoOperationPlan struct {
	OperationID, InstallationID uuid.UUID
	Action, ExpectedPlanSHA256  string
	Components                  []DemoRemovalComponent
	CreatedAt                   time.Time
}

type DemoOwnerCommand struct {
	OperationID, InstallationID, PackageID, ComponentID uuid.UUID
	ReleaseID, SourceRunID                              uuid.UUID
	PackageVersion, ContentSHA256, IdempotencyKey       string
	ExpectedPlanSHA256                                  string
	Document                                            []byte
	RecordIDs                                           []uuid.UUID
}

type DemoOwnerResult struct {
	ReceiptID, SourceRunID uuid.UUID
	Installed              bool
	RecordIDs              []uuid.UUID
	CompletedAt            time.Time
}

type DemoOwner interface {
	Ready(context.Context, platformtx.TenantScope) error
	PlanRemoval(context.Context, platformtx.TenantScope, DemoOwnerCommand) (DemoOwnerResult, error)
	Install(context.Context, platformtx.TenantScope, DemoOwnerCommand) (DemoOwnerResult, error)
	Remove(context.Context, platformtx.TenantScope, DemoOwnerCommand) (DemoOwnerResult, error)
	Inspect(context.Context, platformtx.TenantScope, DemoOwnerCommand) (DemoOwnerResult, error)
}

type DemoStore interface {
	DemoComponents(context.Context, uuid.UUID) ([]DemoComponent, error)
	RecordDemoOperationPlan(context.Context, DemoOperationPlan) (DemoOperationPlan, error)
	ReadDemoOperationPlan(context.Context, uuid.UUID) (DemoOperationPlan, error)
	ApplyDemoResult(context.Context, DemoResultMutation) (DemoComponent, error)
}

type DemoResultMutation struct {
	OperationID, ReceiptID, InstallationID, ComponentID uuid.UUID
	Action, ContentSHA256, IdempotencyKey               string
	SourceRunID                                         uuid.UUID
	RecordIDs                                           []uuid.UUID
	ExpectedRevision                                    uint64
	CompletedAt                                         time.Time
}

func nilDemoOwner(value DemoOwner) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
