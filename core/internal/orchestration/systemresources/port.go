package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type Port interface {
	Catalog(context.Context, Admission, publicresource.Resource) (CatalogResult, error)
	PrepareCreate(context.Context, publicresource.CreateInput) (publicresource.PreparedCreate, error)
	Create(context.Context, Admission, publicresource.PreparedCreate) (MutationResult, error)
	Get(context.Context, Admission, publicresource.GetInput) (ReadResult, error)
	Query(context.Context, Admission, publicresource.QueryInput) (QueryResult, error)
	Update(context.Context, Admission, publicresource.UpdateInput) (MutationResult, error)
	Delete(context.Context, Admission, publicresource.DeleteInput) (DeleteResult, error)
	ChangeState(context.Context, Admission, publicresource.StateInput) (MutationResult, error)
}
