package systemresourcelayouts

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PublishInput struct {
	Resource            publicresource.Resource
	RevisionID          string
	ExpectedRowRevision int64
}

func (application *Application) Publish(
	ctx context.Context,
	admission publicresource.Admission,
	input PublishInput,
) (layoutapi.Revision, error) {
	kind, err := layoutResource(input.Resource)
	actor, actorErr := admission.Actor()
	if err != nil || actorErr != nil || input.RevisionID == "" ||
		input.ExpectedRowRevision < 1 ||
		input.ExpectedRowRevision == maximumLayoutRowRevision {
		return layoutapi.Revision{}, ErrContract
	}
	var result layoutapi.Revision
	err = application.runCommand(
		ctx, admission, publicresource.LayoutPublishCatalogOperation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			current, currentErr := application.layout.GetLayout(
				scoped, scope, input.RevisionID,
			)
			if currentErr != nil {
				return currentErr
			}
			if current.Resource != kind {
				return layoutapi.ErrNotFound
			}
			if !publishStateMatches(current, input.ExpectedRowRevision) {
				return layoutapi.ErrInvalid
			}
			catalog, currentErr := application.iam.CatalogForLayoutPublish(
				scoped, scope, actor, input.Resource,
			)
			if currentErr != nil {
				return currentErr
			}
			presentation, currentErr := catalog.Presentation()
			if currentErr != nil {
				return currentErr
			}
			result, currentErr = application.layout.PublishPresentation(
				scoped, scope, layoutapi.PublishPresentationInput{
					RevisionID:          input.RevisionID,
					ExpectedRowRevision: input.ExpectedRowRevision,
					Catalog:             presentation,
				},
			)
			return currentErr
		},
	)
	return result, err
}

func publishStateMatches(revision layoutapi.Revision, expected int64) bool {
	return revision.State == layoutapi.RevisionDraft &&
		revision.RowRevision == expected ||
		revision.State == layoutapi.RevisionPublished &&
			revision.RowRevision == expected+1
}

const maximumLayoutRowRevision = int64(^uint64(0) >> 1)
