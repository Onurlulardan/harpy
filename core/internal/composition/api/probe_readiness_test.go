package api

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

func TestAllReadyRequiresEveryFixedObservation(t *testing.T) {
	if !allReady(health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}) {
		t.Fatal("fully ready observation rejected")
	}
	for _, states := range []health.States{
		{DatabaseRole: true, DeploymentPolicy: true, SchemaContract: true},
		{Database: true, DeploymentPolicy: true, SchemaContract: true},
		{Database: true, DatabaseRole: true, SchemaContract: true},
		{Database: true, DatabaseRole: true, DeploymentPolicy: true},
	} {
		if allReady(states) {
			t.Fatalf("partial observation accepted: %+v", states)
		}
	}
}

func TestLocalReadinessFailsClosed(t *testing.T) {
	if !localDependenciesReady(nil) ||
		localDependenciesReady([]func() bool{func() bool { return false }}) ||
		localDependenciesReady([]func() bool{nil}) ||
		!localDependenciesReady([]func() bool{func() bool { return true }}) {
		t.Fatal("local readiness did not fail closed")
	}
}
