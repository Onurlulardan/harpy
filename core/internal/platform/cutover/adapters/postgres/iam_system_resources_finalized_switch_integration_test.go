//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMSystemResourcesProtectedSwitchAgainstPostgreSQL184(t *testing.T) {
	for index, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
			defer cancel()
			database := startCutoverIntegrationDatabaseForOrigin(t, ctx, origin)
			prepareIAMSystemResourcesAsyncSurfaces(t, ctx, database, origin)
			verifyIAMSystemResourcesProtectedSwitch(
				t, ctx, database, 120+index,
			)
		})
	}
}

func verifyIAMSystemResourcesProtectedSwitch(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	identityIndex int,
) {
	t.Helper()
	authorityBinary := buildCutoverAuthorityBinary(t, ctx)
	migratorBinary := buildCutoverMigratorBinary(t, ctx)
	directory := t.TempDir()
	now := time.Now().UTC()
	signer := prepareBuiltCutoverSigner(t, directory, now)
	seedReleaseReadiness(t, ctx, database.admin)
	addLastTenantCheckpoint(t, ctx, database.admin)
	assertObserverRequiresAdoptedCatalog(t, ctx, database)
	authority := buildBuiltAuthorityArtifacts(t, ctx, database, signer, directory, now)
	authority = adoptBuiltCutoverAuthority(
		t, ctx, authorityBinary, database, signer, authority, directory,
	)
	hooks := iamSystemResourcesFinalizationHooks(t, ctx, database)
	verifyProtectedFinalizationState(
		t, ctx, database, authorityBinary, migratorBinary, signer,
		authority, identityIndex, "completed", true, hooks,
	)
}

func iamSystemResourcesFinalizationHooks(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) protectedFinalizationHooks {
	t.Helper()
	return protectedFinalizationHooks{
		beforeAuthorization: func(
			adoption builtCutoverLeaseArtifacts, evidence finalizationEvidence,
		) {
			prepareIAMSystemResourcesQuiescing(
				t, ctx, database, adoption.identity.scopeID, evidence.terminalSHA256,
			)
			seedIsolatedIAMSystemResourcesGroup(t, ctx, database)
			seedIAMSystemResourcesAcceptedZeroInventory(
				t, ctx, database, adoption.identity.scopeID,
				"iam", evidence.terminalSHA256,
			)
		},
		afterAuthorization: func(
			adoption builtCutoverLeaseArtifacts, evidence finalizationEvidence,
		) {
			assertMissingLayoutZeroInventoryBlocksSwitch(t, ctx, database, evidence)
			seedIAMSystemResourcesAcceptedZeroInventory(
				t, ctx, database, adoption.identity.scopeID,
				"system_resource_layout", evidence.terminalSHA256,
			)
			assertIAMSystemResourcesFinalizationRollback(t, ctx, database, evidence)
		},
		afterVerification: func(
			adoption builtCutoverLeaseArtifacts, evidence finalizationEvidence,
		) {
			raceIAMSystemResourcesFinalization(t, ctx, database, evidence)
			assertIAMSystemResourcesFinalizedSwitch(
				t, ctx, database, adoption.identity.scopeID, evidence,
			)
		},
	}
}
