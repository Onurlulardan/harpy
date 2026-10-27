package billing

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	"github.com/google/uuid"
)

func TestDevelopmentAuthorityIsSignedBoundAndIdempotent(t *testing.T) {
	environmentID := uuid.Must(uuid.NewV7())
	manifest, err := installationv1.Build(installationv1.BuildInput{
		EnvironmentID: environmentID.String(), DeploymentMode: installationv1.DeploymentOnPremise,
		PrimaryTenant: installationv1.TenantMetadata{Name: "Local", Slug: "local"},
		FirstAdmin: installationv1.AdminMetadata{
			DisplayName: "Owner", Email: "owner@example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	if err = ReconcileDevelopmentAuthority(root, manifest.CanonicalBytes(), now); err != nil {
		t.Fatal(err)
	}
	snapshot := filepath.Join(root, developmentAuthorityCurrent, developmentSnapshotFile)
	before, err := os.ReadFile(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err = ReconcileDevelopmentAuthority(
		root, manifest.CanonicalBytes(), now.Add(time.Hour),
	); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(snapshot)
	if err != nil || string(before) != string(after) {
		t.Fatal("development authority was regenerated")
	}
}
