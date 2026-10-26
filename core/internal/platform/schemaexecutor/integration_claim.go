//go:build integration

package schemaexecutor

// AcceptIntegrationClaim exposes the opaque lease only to exact-database tests.
func AcceptIntegrationClaim(record ClaimRecord) (Lease, ValidatedPlan, error) {
	claim, err := acceptClaim(record)
	if err != nil {
		return Lease{}, ValidatedPlan{}, err
	}
	plan, err := ValidateClaim(claim)
	if err != nil {
		return Lease{}, ValidatedPlan{}, err
	}
	return claim.lease, plan, nil
}
