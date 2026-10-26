package platformlicense

import (
	"context"
	"encoding/json"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

func (store *Store) ReadLicense(
	ctx context.Context,
	environmentID uuid.UUID,
) (platformlicenseapp.License, error) {
	if !store.usable(ctx) {
		return platformlicenseapp.License{}, platformlicenseapp.ErrUnavailable
	}
	value := platformlicenseapp.License{}
	var features, limits string
	err := store.database.QueryRow(ctx, readLicenseSQL, environmentID).Scan(
		&value.LicenseID, &value.EnvironmentID, &value.SnapshotSequence,
		&value.SnapshotSHA256, &value.State, &value.ReleaseChannel,
		&features, &limits, &value.DatabaseNow,
	)
	if err != nil {
		return platformlicenseapp.License{}, classify(err)
	}
	if json.Unmarshal([]byte(features), &value.Features) != nil ||
		json.Unmarshal([]byte(limits), &value.Limits) != nil {
		return platformlicenseapp.License{}, platformlicenseapp.ErrUnavailable
	}
	value.DatabaseNow = value.DatabaseNow.UTC()
	return value, nil
}
