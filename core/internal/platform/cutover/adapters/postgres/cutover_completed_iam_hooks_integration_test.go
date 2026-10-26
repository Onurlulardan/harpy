//go:build integration

package postgres

import (
	"context"
	"testing"
)

func completedIAMSystemResourcesHooks(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) protectedFinalizationHooks {
	t.Helper()
	return protectedFinalizationHooks{beforeAuthorization: func(
		adoption builtCutoverLeaseArtifacts, evidence finalizationEvidence,
	) {
		prepareIAMSystemResourcesQuiescing(
			t, ctx, database, adoption.identity.scopeID, evidence.terminalSHA256,
		)
		for _, domain := range []string{"iam", "system_resource_layout"} {
			seedIAMSystemResourcesAcceptedZeroInventory(
				t, ctx, database, adoption.identity.scopeID,
				domain, evidence.terminalSHA256,
			)
		}
	}}
}
