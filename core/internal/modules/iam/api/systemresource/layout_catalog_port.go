package systemresource

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const LayoutPublishCatalogOperation = "system_resource_layout.publish"

// LayoutCatalogPort exposes only the IAM catalog reads required by the
// system-resource presentation owner. Command use is restricted to publication.
type LayoutCatalogPort interface {
	Catalog(
		context.Context,
		transaction.TenantScope,
		ActorEvidence,
		Resource,
	) (FieldCatalog, error)
	CatalogForLayoutPublish(
		context.Context,
		transaction.TenantScope,
		ActorEvidence,
		Resource,
	) (FieldCatalog, error)
}
