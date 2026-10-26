package recoverysigning

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

const fixtureEnvironmentID = "018f0000-0000-7000-8000-000000000001"

func TestReconcilePublishesThenVerifiesWithoutRegeneration(t *testing.T) {
	target := filepath.Join(t.TempDir(), "signing")
	first, err := reconcile(context.Background(), target, fixtureEnvironmentID,
		fixtureDependencies(nil))
	if err != nil || !first.IsValid() || first.Reused {
		t.Fatalf("first reconcile = %#v, %v", first, err)
	}
	before, err := os.ReadFile(filepath.Join(target, recoveryv1.ControlSignerFilename))
	if err != nil {
		t.Fatal(err)
	}
	second, err := reconcile(context.Background(), target, fixtureEnvironmentID,
		fixtureDependencies(bytes.NewReader(bytes.Repeat([]byte{9}, 96))))
	if err != nil || !second.Reused || second.TrustBundleSHA256 != first.TrustBundleSHA256 {
		t.Fatalf("second reconcile = %#v, %v", second, err)
	}
	after, _ := os.ReadFile(filepath.Join(target, recoveryv1.ControlSignerFilename))
	if !bytes.Equal(before, after) {
		t.Fatal("completed signing set was regenerated")
	}
}

func TestReconcileResumesInterruptedStaging(t *testing.T) {
	target := filepath.Join(t.TempDir(), "signing")
	deps := fixtureDependencies(nil)
	deps.beforePublish = func() error { return errors.New("interrupt") }
	if _, err := reconcile(context.Background(), target, fixtureEnvironmentID, deps); !errors.Is(err, ErrInterrupted) {
		t.Fatalf("interrupt error = %v", err)
	}
	if _, err := os.Stat(target + ".staging"); err != nil {
		t.Fatal("interrupted staging evidence is missing")
	}
	result, err := reconcile(context.Background(), target, fixtureEnvironmentID,
		fixtureDependencies(nil))
	if err != nil || !result.IsValid() {
		t.Fatalf("resume = %#v, %v", result, err)
	}
}

func fixtureDependencies(random *bytes.Reader) dependencies {
	if random == nil {
		entropy := append(bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 32)...)
		entropy = append(entropy, bytes.Repeat([]byte{3}, 32)...)
		random = bytes.NewReader(entropy)
	}
	return dependencies{random: random, now: func() time.Time {
		return time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	}}
}
