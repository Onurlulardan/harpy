package licenseauthority

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

type runtimeGrantWire struct {
	LicenseID        uuid.UUID               `json:"license_id"`
	EnvironmentID    uuid.UUID               `json:"environment_id"`
	SnapshotSequence int64                   `json:"snapshot_sequence"`
	SnapshotSHA256   string                  `json:"snapshot_sha256"`
	State            licenseapp.RuntimeState `json:"state"`
	EntitlementKey   string                  `json:"entitlement_key"`
	ValueType        string                  `json:"value_type"`
	BooleanValue     *bool                   `json:"boolean_value"`
	IntegerValue     *int64                  `json:"integer_value"`
	Unlimited        bool                    `json:"unlimited"`
	Channel          string                  `json:"channel"`
	DatabaseNow      time.Time               `json:"database_now"`
}

func (storage *store) LoadLicenseRuntimeEntitlements(
	ctx context.Context,
	environmentID uuid.UUID,
) ([]licenseapp.RuntimeGrant, error) {
	if !storage.usable(ctx) {
		return nil, licenseapp.ErrAuthorityState
	}
	var document string
	if err := storage.database.QueryRow(
		ctx, readRuntimeSQL, environmentID,
	).Scan(&document); err != nil {
		return nil, fmt.Errorf("read billing license runtime entitlements: %w", err)
	}
	wire := []runtimeGrantWire{}
	if err := json.Unmarshal([]byte(document), &wire); err != nil {
		return nil, fmt.Errorf("decode billing license runtime entitlements: %w", err)
	}
	grants := make([]licenseapp.RuntimeGrant, len(wire))
	for index, grant := range wire {
		grants[index] = licenseapp.RuntimeGrant{
			LicenseID:        grant.LicenseID,
			EnvironmentID:    grant.EnvironmentID,
			SnapshotSequence: grant.SnapshotSequence,
			SnapshotSHA256:   grant.SnapshotSHA256,
			State:            grant.State,
			EntitlementKey:   grant.EntitlementKey,
			ValueType:        grant.ValueType,
			BooleanValue:     grant.BooleanValue,
			IntegerValue:     grant.IntegerValue,
			Unlimited:        grant.Unlimited,
			Channel:          grant.Channel,
			DatabaseNow:      grant.DatabaseNow.UTC(),
		}
	}
	return grants, nil
}
