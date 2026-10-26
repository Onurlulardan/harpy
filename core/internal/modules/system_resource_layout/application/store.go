package application

import (
	"context"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

type CreateStoreInput struct {
	ID               uuid.UUID
	Resource         layoutapi.ResourceKind
	DefinitionJSON   string
	DefinitionSHA256 [32]byte
	Now              time.Time
}

type UpdateStoreInput struct {
	ID                  uuid.UUID
	ExpectedRowRevision int64
	DefinitionJSON      string
	DefinitionSHA256    [32]byte
	Now                 time.Time
}

type PublishStoreInput struct {
	ID                     uuid.UUID
	ExpectedRowRevision    int64
	SourceDefinitionSHA256 [32]byte
	CompilerVersion        string
	CatalogContract        string
	CatalogJSON            []byte
	CatalogSHA256          [32]byte
	PresentationContract   string
	PresentationJSON       []byte
	PresentationSHA256     [32]byte
	Now                    time.Time
}

type PublishedPresentation struct {
	Revision               layoutapi.Revision
	Generation             int64
	SourceDefinitionSHA256 [32]byte
	CompilerVersion        string
	CatalogContract        string
	CatalogJSON            []byte
	CatalogSHA256          [32]byte
	PresentationContract   string
	PresentationJSON       []byte
	PresentationSHA256     [32]byte
	EvidenceComplete       bool
}

type Store interface {
	Create(context.Context, CreateStoreInput) (layoutapi.Revision, error)
	Get(context.Context, uuid.UUID) (layoutapi.Revision, error)
	List(context.Context, layoutapi.ResourceKind) ([]layoutapi.Revision, error)
	Update(context.Context, UpdateStoreInput) (layoutapi.Revision, error)
	Delete(context.Context, uuid.UUID, int64) error
	Publish(context.Context, PublishStoreInput) (PublishedPresentation, error)
	Resolve(context.Context, layoutapi.ResourceKind) (PublishedPresentation, error)
}
