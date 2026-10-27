package postrestore

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

func (store *Store) Admission(ctx context.Context) (application.Admission, error) {
	var result application.Admission
	err := store.global(ctx, snapshotQuery, func(database database) error {
		err := database.QueryRow(ctx, admissionSQL).Scan(
			&result.State, &result.EnvironmentID, &result.RestoreOperationID)
		if store.rowNotFound(err) {
			result.State = "normal"
			return nil
		}
		return err
	})
	return result, err
}
