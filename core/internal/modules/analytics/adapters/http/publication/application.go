package publicationhttp

import (
	"context"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
)

type Application interface {
	CreatePublication(context.Context, httptransport.Access, publicationapp.CreateRequest) (publicationapp.Snapshot, error)
	GetPublication(context.Context, httptransport.Access, uuid.UUID) (publicationapp.Snapshot, error)
	ListPublications(context.Context, httptransport.Access, int, *uuid.UUID) (publicationapp.Page, error)
	SubmitPublication(context.Context, httptransport.Access, publicationapp.TransitionRequest) (publicationapp.Snapshot, error)
	ApprovePublication(context.Context, httptransport.Access, publicationapp.TransitionRequest) (publicationapp.Snapshot, error)
	ActivatePublication(context.Context, httptransport.Access, publicationapp.ActivateRequest) (publicationapp.Activation, error)
	RevokePublication(context.Context, httptransport.Access, publicationapp.TransitionRequest) (publicationapp.Snapshot, error)
}

type PublicApplication interface {
	Execute(context.Context, publicationapp.PublicRequest) (publicationapp.PublicResult, error)
}
