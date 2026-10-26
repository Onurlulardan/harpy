package formsreference

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ID [16]byte

type Target struct{ DefinitionID, VersionID ID }

type Source struct {
	Kind          string
	ID, VersionID ID
	ComponentID   *ID
}

type Coordinator interface {
	BindVersion(
		context.Context,
		platformtx.TenantScope,
		Source,
		[]Target,
		time.Time,
	) error
	DecommissionSource(
		context.Context,
		platformtx.TenantScope,
		ID,
		*ID,
		time.Time,
	) error
	DecommissionComponent(
		context.Context,
		platformtx.TenantScope,
		ID,
		time.Time,
	) error
	CatalogSource(
		context.Context,
		platformtx.TenantScope,
		ID,
	) (Source, error)
}
