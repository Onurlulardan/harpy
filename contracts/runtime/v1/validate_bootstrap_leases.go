package runtimev1

import (
	"fmt"
	"time"
)

const maximumBootstrapLeaseDuration = 15 * time.Minute

func validateBootstrapCredentialLeases(
	leases []BootstrapCredentialLease,
	authorityIssuedAt, authorityExpiresAt time.Time,
) error {
	if len(leases) != 2 {
		return fmt.Errorf("bootstrap authority must contain exactly two credential leases")
	}
	expectedPrincipals := []string{"harpy_cluster_admin", "harpy_extension_admin"}
	leaseIDs := make(map[string]struct{}, len(leases))
	for index, lease := range leases {
		if err := validateUUID(lease.LeaseID, "lease_id"); err != nil {
			return err
		}
		if _, duplicate := leaseIDs[lease.LeaseID]; duplicate {
			return fmt.Errorf("credential lease IDs must be unique")
		}
		leaseIDs[lease.LeaseID] = struct{}{}
		if lease.Principal != expectedPrincipals[index] ||
			lease.MaxDurationSeconds != int(maximumBootstrapLeaseDuration/time.Second) ||
			lease.MaxConnections != 1 ||
			lease.CleanupPolicy != "disable_login_clear_password_terminate_sessions" {
			return fmt.Errorf("credential lease %d contract is invalid", index)
		}
		notBefore, err := validateTimestamp(lease.NotBefore, "lease.not_before")
		if err != nil {
			return err
		}
		expiresAt, err := validateTimestamp(lease.ExpiresAt, "lease.expires_at")
		if err != nil || !notBefore.Before(expiresAt) ||
			expiresAt.Sub(notBefore) > maximumBootstrapLeaseDuration ||
			notBefore.Before(authorityIssuedAt) || expiresAt.After(authorityExpiresAt) {
			return fmt.Errorf("credential lease %d validity window is invalid", index)
		}
	}
	return nil
}
