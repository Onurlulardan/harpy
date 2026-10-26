package systemtransferartifact

import (
	"context"
	"errors"
	"strings"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

const importObjectSuffix = "/1.fcart"

func (runtime *Runtime) DeleteExactImportSource(
	ctx context.Context,
	key string,
	version string,
) (systemtransfer.SourceDeleteResult, error) {
	if runtime == nil || !runtimeInputContext(ctx) ||
		!validImportObjectIdentity(key, version) {
		return systemtransfer.SourceDeleteResult{},
			systemtransfer.ErrSourceCleanupInvalid
	}
	deleted, err := runtime.imports.DeleteExact(
		ctx, artifactobject.Identity{Key: key, Version: version},
	)
	if err != nil {
		if errors.Is(err, artifactobject.ErrOutcomeUnknown) {
			return systemtransfer.SourceDeleteResult{},
				systemtransfer.ErrSourceDeleteOutcomeUnknown
		}
		return systemtransfer.SourceDeleteResult{}, errors.Join(
			systemtransfer.ErrSourceCleanupUnavailable, err,
		)
	}
	result := systemtransfer.SourceDeleteResult{
		ProviderRequestID:    deleted.ProviderRequestID,
		DeleteResponseSHA256: deleted.DeleteResponseSHA256,
		AbsenceProbeSHA256:   deleted.AbsenceProbeSHA256,
		VerifiedAbsentAt:     deleted.VerifiedAbsentAt.UTC(),
	}
	if result.Validate() != nil {
		return systemtransfer.SourceDeleteResult{},
			systemtransfer.ErrSourceDeleteOutcomeUnknown
	}
	return result, nil
}

func validImportObjectIdentity(key, version string) bool {
	prefix := importPrefix + "/v1/"
	if len(version) < 1 || len(version) > 512 ||
		!strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, importObjectSuffix) {
		return false
	}
	identity := strings.TrimSuffix(strings.TrimPrefix(key, prefix), importObjectSuffix)
	sourceID, err := uuid.Parse(identity)
	return err == nil && sourceID.Version() == uuid.Version(7) &&
		key == prefix+sourceID.String()+importObjectSuffix
}

var _ systemtransfer.SourceObjectDeleter = (*Runtime)(nil)
