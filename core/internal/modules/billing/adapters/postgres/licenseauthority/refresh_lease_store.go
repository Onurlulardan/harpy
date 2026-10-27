package licenseauthority

import (
	"context"
	"fmt"
	"time"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

const refreshLeaseSeconds = 60

func (storage *store) AcquireRefreshLease(
	ctx context.Context,
	command licenseapp.RefreshCommand,
	versions []licenseapp.ComponentVersion,
) (licenseapp.RefreshClaim, error) {
	if !storage.usable(ctx) {
		return licenseapp.RefreshClaim{}, licenseapp.ErrAuthorityState
	}
	components, err := encodeComponentVersions(versions)
	if err != nil {
		return licenseapp.RefreshClaim{}, fmt.Errorf("encode license components: %w", err)
	}
	claim := licenseapp.RefreshClaim{}
	var delivery string
	var nonce []byte
	var headFound bool
	var headSequence int64
	var headDigest string
	err = storage.database.QueryRow(ctx, claimRefreshSQL,
		command.OperationID, command.Binding.LicenseID,
		command.Binding.LicenseeOrganizationID, command.Binding.EnvironmentID,
		string(command.Binding.DeliveryKind), components, refreshLeaseSeconds,
	).Scan(
		&claim.OperationID, &claim.LeaseToken, &claim.Binding.LicenseID,
		&claim.Binding.LicenseeOrganizationID, &claim.Binding.EnvironmentID,
		&delivery, &nonce, &headFound, &headSequence, &headDigest,
		&claim.DatabaseNow, &claim.LeaseExpiresAt,
	)
	if err != nil {
		return licenseapp.RefreshClaim{}, fmt.Errorf("claim license refresh: %w", err)
	}
	if len(nonce) != len(claim.Nonce) {
		return licenseapp.RefreshClaim{}, licenseapp.ErrAuthorityState
	}
	copy(claim.Nonce[:], nonce)
	claim.Binding.DeliveryKind = licenseapp.DeliveryKind(delivery)
	claim.DatabaseNow = claim.DatabaseNow.In(time.UTC)
	claim.LeaseExpiresAt = claim.LeaseExpiresAt.In(time.UTC)
	if headFound {
		claim.Head = &licenseapp.AcceptedHead{
			LicenseID:     claim.Binding.LicenseID,
			EnvironmentID: claim.Binding.EnvironmentID,
			Sequence:      headSequence,
			Digest:        headDigest,
		}
	}
	return claim, nil
}
