package recoverysigning

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func TestReconcileFencesPartialTamperedAndWrongEnvironment(t *testing.T) {
	tests := []func(t *testing.T, target string){
		func(t *testing.T, target string) {
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
		},
		func(t *testing.T, target string) {
			mustPublished(t, target)
			path := filepath.Join(target, recoveryv1.ControlSignerFilename)
			if err := os.WriteFile(path, make([]byte, 32), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		func(t *testing.T, target string) {
			mustPublished(t, target)
			if err := os.WriteFile(filepath.Join(target, "extra"), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
	}
	for _, mutate := range tests {
		target := filepath.Join(t.TempDir(), "signing")
		mutate(t, target)
		if _, err := reconcile(context.Background(), target, fixtureEnvironmentID,
			fixtureDependencies(nil)); !errors.Is(err, ErrFenced) {
			t.Fatalf("fence error = %v", err)
		}
	}
	target := filepath.Join(t.TempDir(), "signing")
	mustPublished(t, target)
	if _, err := reconcile(context.Background(), target,
		"018f0000-0000-7000-8000-000000000002", fixtureDependencies(nil)); !errors.Is(err, ErrFenced) {
		t.Fatalf("wrong environment error = %v", err)
	}
}

func mustPublished(t *testing.T, target string) {
	t.Helper()
	if _, err := reconcile(context.Background(), target, fixtureEnvironmentID,
		fixtureDependencies(nil)); err != nil {
		t.Fatal(err)
	}
}
