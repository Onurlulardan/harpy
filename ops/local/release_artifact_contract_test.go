package local

import (
	"strings"
	"testing"
)

func TestLocalReleaseArtifactServiceWritesReleaseVolume(t *testing.T) {
	compose := readContractText(t, "compose.tools.yml")
	service := composeServiceBlock(t, compose, "local-release-artifact")
	required := []string{
		`restart: "no"`,
		`entrypoint: ["/app/local-release", "artifact"]`,
		"release-state:/run/release\n",
		"network_mode: none",
	}
	for _, fragment := range required {
		if !strings.Contains(service, fragment) {
			t.Fatalf("local-release-artifact missing %q", fragment)
		}
	}
	if strings.Contains(service, "release-state:/run/release:ro") {
		t.Fatal("release volume is mounted read-only")
	}
	if strings.Contains(service, "local-release prepare") {
		t.Fatal("artifact service invokes local-release prepare")
	}
}
