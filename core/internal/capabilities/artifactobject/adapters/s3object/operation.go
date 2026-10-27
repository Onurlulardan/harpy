package s3object

import (
	"context"
	"errors"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) operationError(ctx context.Context) error {
	if store == nil || nilDependency(store.api) || nilDependency(store.clock) ||
		store.bucket == "" || store.prefix == "" || store.contract == "" ||
		store.maximumBytes < 1 {
		return ErrConfiguration
	}
	if ctx == nil {
		return ErrObjectInput
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(artifactobject.ErrUnavailable, err)
	}
	return nil
}
