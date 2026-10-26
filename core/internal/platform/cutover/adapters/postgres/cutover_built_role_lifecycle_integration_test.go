//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestBuiltCutoverRoleLifecycleUsesPG18SCRAM(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	expiresAt := time.Now().UTC().Add(2 * time.Minute).Truncate(time.Microsecond)

	leaseBuiltCutoverRoles(t, ctx, database, expiresAt)
	assertBuiltLeaseCatalog(
		t, ctx, database.admin, "harpy_cutover_decoder", expiresAt, true, 2,
	)
	assertBuiltLeaseCatalog(
		t, ctx, database.admin, "harpy_adopter", expiresAt, false, 1,
	)
	assertBuiltRoleConnectionAccepted(t, ctx, database, "harpy_cutover_decoder")
	assertBuiltRoleConnectionAccepted(t, ctx, database, "harpy_adopter")
	adoptionApplication := cutoverAdoptionApplicationName(integrationRunID)
	assertBuiltRoleApplicationName(
		t, ctx, database, "harpy_cutover_decoder", adoptionApplication,
	)
	assertBuiltRoleApplicationName(t, ctx, database, "harpy_adopter", adoptionApplication)
	sealBuiltCutoverRoles(t, ctx, database)
	assertBuiltSealedRole(t, ctx, database.admin, "harpy_cutover_decoder", expiresAt)
	assertBuiltSealedRole(t, ctx, database.admin, "harpy_adopter", expiresAt)
	cleanupBuiltAdoptionRoles(t, ctx, database)

	leaseBuiltFinalizerRole(t, ctx, database, expiresAt)
	assertBuiltLeaseCatalog(
		t, ctx, database.admin, "harpy_cutover_finalizer", expiresAt, false, 1,
	)
	assertBuiltRoleConnectionAccepted(t, ctx, database, "harpy_cutover_finalizer")
	assertBuiltRoleApplicationName(
		t, ctx, database, "harpy_cutover_finalizer",
		cutoverFinalizerApplicationName(integrationRunID),
	)
	cleanupBuiltFinalizerRole(t, ctx, database)

	assertExpiredBuiltRolePasswordsDenied(t, ctx, database)
}
