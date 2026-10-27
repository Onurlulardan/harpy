package api

import (
	"context"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
)

const (
	layoutHTTPActionRead   = "read"
	layoutHTTPActionUpdate = "update"
)

type systemResourceLayoutHTTPAccessVerifier interface {
	Verify(context.Context, layouts.LayoutAccessRequest) (
		layouts.LayoutAccessDecision, error,
	)
}

type systemResourceLayoutHTTPApplication interface {
	Create(context.Context, layouts.Admission, layouts.Resource,
		string, layouts.Definition) (layouts.Revision, error)
	Get(context.Context, layouts.Admission, layouts.Resource,
		string) (layouts.Revision, error)
	List(context.Context, layouts.Admission, layouts.Resource) (
		[]layouts.Revision, error,
	)
	Update(context.Context, layouts.Admission, layouts.Resource,
		layouts.UpdateInput) (layouts.Revision, error)
	Delete(context.Context, layouts.Admission, layouts.Resource,
		layouts.DeleteInput) error
	Publish(context.Context, layouts.Admission, layouts.PublishInput) (
		layouts.Revision, error,
	)
	Resolve(context.Context, layouts.Admission, layouts.Resource) (
		layouts.PresentationDocument, error,
	)
}
