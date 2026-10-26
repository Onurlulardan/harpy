package platformops

import (
	"context"
	"errors"
	"testing"
)

type installationRecoveryRow struct {
	replay bool
	err    error
}

func (row installationRecoveryRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	*(destinations[0].(*bool)) = row.replay
	return nil
}

func TestInitializeInstallationRecoveryIdentityEnforcesModeShape(t *testing.T) {
	input := installationRecoveryFixture("onpremise")
	called := 0
	executor := InstallationActivationExecutor{QueryRow: func(
		context.Context, string, ...any,
	) InstallationActivationRow {
		called++
		return installationRecoveryRow{}
	}}
	if err := InitializeInstallationRecoveryIdentity(context.Background(), executor, input); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("query calls = %d, want 1", called)
	}
	input.DeploymentMode = "saas"
	if !errors.Is(InitializeInstallationRecoveryIdentity(
		context.Background(), executor, input,
	), ErrInstallationRecoveryIdentityInvalid) {
		t.Fatal("SaaS accepted an on-premise recovery pin")
	}
	input.SigningGeneration = 0
	input.TrustBundleSHA256 = ""
	input.RecoverySigningPresent = false
	if err := InitializeInstallationRecoveryIdentity(context.Background(), executor, input); err != nil {
		t.Fatal(err)
	}
}

func installationRecoveryFixture(mode string) InstallationRecoveryIdentityInput {
	return InstallationRecoveryIdentityInput{
		OperationID: [16]byte{1}, EnvironmentID: [16]byte{2}, TenantID: [16]byte{3},
		SchemaName: "t_primary", DeploymentMode: mode,
		SigningGeneration: 1, TrustBundleSHA256: repeatDigest("a"),
		RecoverySigningPresent: true, ManifestSHA256: repeatDigest("b"),
	}
}

func repeatDigest(value string) string {
	result := ""
	for len(result) < 64 {
		result += value
	}
	return result
}
