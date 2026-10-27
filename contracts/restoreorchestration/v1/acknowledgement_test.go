package restoreorchestrationv1

import (
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAcknowledgementBindsDrainAndSwitchAuthority(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	environmentID := uuid.NewString()
	operationID := uuid.NewString()
	attemptID := mustUUIDv7(t)
	previousID := uuid.NewString()
	newID := uuid.NewString()
	now := time.Now().UTC().Truncate(time.Microsecond)
	trust, err := BuildTrust(TrustInput{
		EnvironmentID:           environmentID,
		DeploymentBindingSHA256: strings.Repeat("a", 64),
		Generation:              1, KeyID: "host.g1", PublicKey: public,
	})
	if err != nil {
		t.Fatal(err)
	}
	drain := AcknowledgementInput{
		Purpose: PurposeDrain, EnvironmentID: environmentID,
		OperationID: operationID, AttemptID: attemptID,
		PlanSHA256: strings.Repeat("b", 64), Sequence: 1,
		IssuedAt: now, ExpiresAt: now.Add(time.Minute),
		PreviousDatabaseSystemID:      previousID,
		PreviousDatabaseLocatorSHA256: strings.Repeat("c", 64),
		NewDatabaseLocatorSHA256:      strings.Repeat("d", 64),
		StoppedServices:               []string{"core-api", "core-worker", "frontend-v3"},
	}
	artifact, err := BuildAcknowledgement(drain, trust, private)
	if err != nil || artifact.Purpose() != PurposeDrain {
		t.Fatalf("drain acknowledgement: %v", err)
	}
	switchInput := drain
	switchInput.Purpose = PurposeSwitch
	switchInput.Sequence = 2
	switchInput.NewDatabaseSystemID = newID
	switchInput.StoppedServices = nil
	switchInput.SwitchOutcome = "switched_compare_and_swap"
	if _, err = BuildAcknowledgement(switchInput, trust, private); err != nil {
		t.Fatalf("switch acknowledgement: %v", err)
	}
	switchInput.PlanSHA256 = strings.Repeat("e", 64)
	if _, err = BuildAcknowledgement(switchInput, trust, private); err != nil {
		t.Fatal(err)
	}
}

func TestAcknowledgementRejectsCrossPurposeAndTamper(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Now().UTC().Truncate(time.Microsecond)
	trust, _ := BuildTrust(TrustInput{
		EnvironmentID:           uuid.NewString(),
		DeploymentBindingSHA256: strings.Repeat("a", 64),
		Generation:              1, KeyID: "host.g1", PublicKey: public,
	})
	input := AcknowledgementInput{
		Purpose: PurposeDrain, EnvironmentID: trust.EnvironmentID(),
		OperationID: uuid.NewString(), AttemptID: mustUUIDv7(t),
		PlanSHA256: strings.Repeat("b", 64), Sequence: 1,
		IssuedAt: now, ExpiresAt: now.Add(time.Minute),
		PreviousDatabaseSystemID:      uuid.NewString(),
		PreviousDatabaseLocatorSHA256: strings.Repeat("c", 64),
		NewDatabaseLocatorSHA256:      strings.Repeat("d", 64),
		StoppedServices:               []string{"core-api", "core-worker", "frontend-v3"},
	}
	artifact, err := BuildAcknowledgement(input, trust, private)
	if err != nil {
		t.Fatal(err)
	}
	expected := Expected{
		Purpose: PurposeSwitch, EnvironmentID: input.EnvironmentID,
		OperationID: input.OperationID, AttemptID: input.AttemptID,
		PlanSHA256:                    input.PlanSHA256,
		PreviousDatabaseSystemID:      input.PreviousDatabaseSystemID,
		PreviousDatabaseLocatorSHA256: input.PreviousDatabaseLocatorSHA256,
		NewDatabaseLocatorSHA256:      input.NewDatabaseLocatorSHA256,
	}
	if _, err = ParseAcknowledgement(
		artifact.CanonicalBytes(), trust, expected, now,
	); err == nil {
		t.Fatal("cross-purpose acknowledgement was accepted")
	}
}

func mustUUIDv7(t *testing.T) string {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value.String()
}
