package buildinfo

import (
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestReadAcceptsDevelopmentDefaults(t *testing.T) {
	info, err := Read(runtimev1.ServiceCoreAPI, runtimev1.EnvironmentDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	version := info.Version()
	if version.Version != "0.0.0-dev" || version.Commit != "unknown" {
		t.Fatalf("unexpected development provenance: %#v", version)
	}
}

func TestReadRejectsUnpinnedToolchain(t *testing.T) {
	_, err := read(
		runtimev1.ServiceCoreWorker,
		runtimev1.EnvironmentTest,
		"1.0.0",
		strings.Repeat("a", 40),
		"2026-07-13T00:00:00Z",
		"go1.26.4",
	)
	if err == nil || !strings.Contains(err.Error(), "pinned") {
		t.Fatalf("expected toolchain rejection, got %v", err)
	}
}

func TestReadRequiresProductionProvenance(t *testing.T) {
	_, err := read(
		runtimev1.ServiceCoreAPI,
		runtimev1.EnvironmentProduction,
		"0.0.0-dev",
		"unknown",
		"1970-01-01T00:00:00Z",
		runtimev1.GoVersion,
	)
	if err == nil || !strings.Contains(err.Error(), "commit") {
		t.Fatalf("expected production provenance rejection, got %v", err)
	}
}

func TestReadAcceptsReleaseProvenance(t *testing.T) {
	info, err := read(
		runtimev1.ServiceCoreMigrator,
		runtimev1.EnvironmentStaging,
		"2.3.4-rc.1+build.7",
		strings.Repeat("f", 40),
		"2026-07-13T00:00:00Z",
		runtimev1.GoVersion,
	)
	if err != nil {
		t.Fatal(err)
	}
	if info.Service() != runtimev1.ServiceCoreMigrator {
		t.Fatalf("unexpected service: %s", info.Service())
	}
}
