package licenseauthority

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

func validIssueInput(
	claim RefreshClaim,
	versions []ComponentVersion,
) bool {
	if claim.OperationID == uuid.Nil || !validBinding(claim.Binding) ||
		claim.Nonce == [32]byte{} || len(versions) == 0 || len(versions) > 64 {
		return false
	}
	previous := ""
	for _, version := range versions {
		if !validText(version.Component, 64) || !validText(version.Version, 128) ||
			version.Component <= previous {
			return false
		}
		previous = version.Component
	}
	return sort.SliceIsSorted(versions, func(left, right int) bool {
		return versions[left].Component < versions[right].Component
	})
}

func validRefreshClaim(command RefreshCommand, claim RefreshClaim) bool {
	if claim.OperationID != command.OperationID || claim.LeaseToken == uuid.Nil ||
		claim.Binding != command.Binding || claim.DatabaseNow.IsZero() ||
		claim.DatabaseNow.Location() != time.UTC || claim.LeaseExpiresAt.IsZero() ||
		claim.LeaseExpiresAt.Location() != time.UTC ||
		!claim.DatabaseNow.Before(claim.LeaseExpiresAt) {
		return false
	}
	if claim.Head == nil {
		return true
	}
	return claim.Head.Sequence > 0 && digestPattern.MatchString(claim.Head.Digest) &&
		claim.Head.LicenseID == claim.Binding.LicenseID &&
		claim.Head.EnvironmentID == claim.Binding.EnvironmentID
}
