package local

import (
	"strings"
	"testing"
)

func TestLocalBuildKeepsImageAndBuilderCache(t *testing.T) {
	document := readContractText(t, "../../.taskfiles/local.yml")
	build := taskBlock(t, document, "build", "project")
	if !strings.Contains(build, "build foundation-local-release local-release") ||
		strings.Contains(build, "mcp-service") || strings.Contains(build, "ai-service") {
		t.Fatal("local:build must compile the local images without AI or MCP")
	}
	for _, forbidden := range []string{
		"docker builder prune", "docker image prune", "build --pull",
	} {
		if strings.Contains(build, forbidden) {
			t.Fatalf("local:build must not %q; use local:build:refresh or local:cleanup", forbidden)
		}
	}
}

func TestLocalCleanupReclaimsUnusedImages(t *testing.T) {
	document := readContractText(t, "../../.taskfiles/local-operations.yml")
	cleanup := taskBlock(t, document, "cleanup", "cleanup:images")
	for _, required := range []string{
		"docker builder prune --all -f",
		"docker system prune -a -f",
		"docker volume prune -f",
	} {
		if !strings.Contains(cleanup, required) {
			t.Fatalf("local:cleanup is missing %q", required)
		}
	}
	all := taskBlock(t, document, "cleanup:all", "cleanup:disk")
	for _, required := range []string{
		"docker builder prune --all -f",
		"docker system prune -a -f",
		"docker volume prune -f",
	} {
		if !strings.Contains(all, required) {
			t.Fatalf("local:cleanup:all is missing %q", required)
		}
	}
}

func TestLocalCleanupDiskCompactsWindowsVhdx(t *testing.T) {
	document := readContractText(t, "../../.taskfiles/local-operations.yml")
	block := taskBlock(t, document, "cleanup:disk", "build:refresh")
	for _, required := range []string{
		"docker system prune -a -f",
		"ops/local/scripts/compact-docker-vhdx.ps1",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("local:cleanup:disk is missing %q", required)
		}
	}
	body := readContractText(t, "scripts/compact-docker-vhdx.ps1")
	for _, required := range []string{
		`Join-Path $env:LOCALAPPDATA 'Docker\wsl\disk\docker_data.vhdx'`,
		"wsl --shutdown",
		"compact vdisk",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("compact-docker-vhdx.ps1 is missing %q", required)
		}
	}
}
