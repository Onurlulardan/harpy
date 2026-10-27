package dashboardexportartifact

import (
	"context"
	"errors"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (runtime *Runtime) Delete(
	ctx context.Context, draft exportapp.ArtifactDraft, objectVersion string,
) (exportapp.DeleteReceipt, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil ||
		draft.ObjectKey == "" || objectVersion == "" {
		return exportapp.DeleteReceipt{}, exportapp.ErrInvalid
	}
	result, err := runtime.objects.DeleteExact(ctx, artifactobject.Identity{
		Key: draft.ObjectKey, Version: objectVersion})
	if errors.Is(err, artifactobject.ErrOutcomeUnknown) {
		return exportapp.DeleteReceipt{}, errors.Join(exportapp.ErrOutcomeUnknown, err)
	}
	if err != nil || result.ProviderRequestID == "" || result.VerifiedAbsentAt.IsZero() {
		return exportapp.DeleteReceipt{}, errors.Join(exportapp.ErrUnavailable, err)
	}
	return exportapp.DeleteReceipt{ProviderRequestID: result.ProviderRequestID,
		DeleteResponseSHA256: result.DeleteResponseSHA256,
		AbsenceSHA256:        result.AbsenceProbeSHA256,
		VerifiedAbsentAt:     result.VerifiedAbsentAt.UTC()}, nil
}
