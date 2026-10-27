package installationv1

import "testing"

func TestBuildProducesStrictRoundTrippableManifest(t *testing.T) {
	manifest, err := Build(BuildInput{
		EnvironmentID:  "018f0000-0000-7000-8000-000000000001",
		DeploymentMode: DeploymentOnPremise,
		PrimaryTenant:  TenantMetadata{Name: "Harpy Local", Slug: "harpy"},
		FirstAdmin: AdminMetadata{
			DisplayName: "Local Instance Owner", Email: "superadmin@harpy.io",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(manifest.CanonicalBytes())
	if err != nil || parsed.SHA256() != manifest.SHA256() ||
		parsed.RequestSHA256() != manifest.RequestSHA256() {
		t.Fatalf("built manifest did not round trip: parsed=%+v err=%v", parsed, err)
	}
}
