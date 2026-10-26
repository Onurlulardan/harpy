package local

import (
	"os"
	"strings"
	"testing"
)

func TestLocalBuildKeepsImageAndBuilderCache(t *testing.T) {
	document, err := os.ReadFile("../../.taskfiles/local.yml")
	if err != nil {
		t.Fatalf("read local Taskfile: %v", err)
	}
	build := taskBlock(t, string(document), "build", "project")
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
	document, err := os.ReadFile("../../.taskfiles/local-operations.yml")
	if err != nil {
		t.Fatalf("read local operations Taskfile: %v", err)
	}
	cleanup := taskBlock(t, string(document), "cleanup", "cleanup:images")
	for _, required := range []string{
		"docker builder prune --all -f",
		"docker system prune -a -f",
		"docker volume prune -f",
	} {
		if !strings.Contains(cleanup, required) {
			t.Fatalf("local:cleanup is missing %q", required)
		}
	}
	all := taskBlock(t, string(document), "cleanup:all", "cleanup:disk")
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
	document, err := os.ReadFile("../../.taskfiles/local-operations.yml")
	if err != nil {
		t.Fatalf("read local operations Taskfile: %v", err)
	}
	block := taskBlock(t, string(document), "cleanup:disk", "build:refresh")
	for _, required := range []string{
		"docker system prune -a -f",
		"ops/local/scripts/compact-docker-vhdx.ps1",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("local:cleanup:disk is missing %q", required)
		}
	}
	script, err := os.ReadFile("scripts/compact-docker-vhdx.ps1")
	if err != nil {
		t.Fatalf("read compact script: %v", err)
	}
	body := string(script)
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
