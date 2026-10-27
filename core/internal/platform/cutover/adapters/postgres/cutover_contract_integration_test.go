//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestCutoverApplyAndFinalizationContractsAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	roles := []string{
		"harpy_adopter", "harpy_cutover_decoder", "harpy_cutover_finalizer",
	}
	for _, role := range roles {
		assertCutoverRole(t, ctx, database.admin, role, false, false, 0)
	}
	adoptionExpiresAt := time.Now().UTC().Add(5 * time.Minute).Truncate(time.Microsecond)
	setAdoptionRoleLease(t, ctx, database.admin, adoptionExpiresAt)
	assertCutoverRole(t, ctx, database.admin, "harpy_adopter", true, false, 2)
	assertCutoverRole(t, ctx, database.admin, "harpy_cutover_decoder", true, true, 2)
	verifyCutoverRolePrivilegeBoundary(t, ctx, database, "harpy_adopter")
	verifyDecoderPrivilegeBoundary(t, ctx, database)
	first := connectCutoverRole(t, ctx, database.dsn, "harpy_adopter")
	second := connectCutoverRole(t, ctx, database.dsn, "harpy_adopter")
	decoder := connectCutoverRole(t, ctx, database.dsn, "harpy_cutover_decoder")
	seedCutoverOperation(t, ctx, database.admin, adoptionExpiresAt)
	setCutoverRoleLease(t, ctx, database.admin, "adoption_seal")
	verifySealedAdoptionRoleDenials(
		t, ctx, database, first, decoder, adoptionExpiresAt,
	)
	verifyCompositeApplyContract(t, ctx, database, first, second)
	_ = first.Close(ctx)
	_ = second.Close(ctx)
	_ = decoder.Close(ctx)

	setCutoverRoleLease(t, ctx, database.admin, "adoption_cleanup")
	assertCutoverRole(t, ctx, database.admin, "harpy_adopter", false, false, 0)
	assertCutoverRole(t, ctx, database.admin, "harpy_cutover_decoder", false, false, 0)
	assertCutoverRoleValidityReset(t, ctx, database.admin, "harpy_adopter")
	assertCutoverRoleValidityReset(t, ctx, database.admin, "harpy_cutover_decoder")
	verifyFinalizationContract(t, ctx, database)

	setCutoverRoleLease(t, ctx, database.admin, "finalizer_cleanup")
	assertCutoverRole(t, ctx, database.admin, "harpy_cutover_finalizer", false, false, 0)
}
