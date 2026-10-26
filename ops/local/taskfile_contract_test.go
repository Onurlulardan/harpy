package local

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestLocalUpPreservesInstallationAdmissionOrder(t *testing.T) {
	document, err := os.ReadFile("../../.taskfiles/local.yml")
	if err != nil {
		t.Fatalf("read local Taskfile: %v", err)
	}
	block := taskBlock(t, string(document), "up", "status")
	actual := make([]string, 0)
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- task: ") {
			actual = append(actual, strings.TrimPrefix(line, "- task: "))
		}
	}
	expected := []string{
		"prepare",
		"infra",
		"bootstrap",
		"migrate",
		"installation:artifacts",
		"license",
		"tenant-executor",
		"initialize",
		"services",
		"verify",
		"frontend",
		"status",
	}
	if !slices.Equal(actual, expected) {
		t.Fatalf("local:up order = %v, want %v", actual, expected)
	}
}

func TestLocalInfraReloadsClamAVAfterSignatureUpdate(t *testing.T) {
	document, err := os.ReadFile("../../.taskfiles/local.yml")
	if err != nil {
		t.Fatalf("read local Taskfile: %v", err)
	}
	block := taskBlock(t, string(document), "infra", "bootstrap")
	required := []string{
		"local-kms mailpit",
		"run --rm --no-deps clamav-update",
		"--no-deps --force-recreate clamav",
		"run --rm --no-deps infra-check",
	}
	position := 0
	for _, fragment := range required {
		next := strings.Index(block[position:], fragment)
		if next < 0 {
			t.Fatalf("local:infra is missing or misorders %q", fragment)
		}
		position += next + len(fragment)
	}
	if strings.Contains(block, "local-kms mailpit clamav") {
		t.Fatal("ClamAV must not start before the signature updater completes")
	}
}

func TestLocalBuildUsesCacheUnlessRefreshIsExplicit(t *testing.T) {
	document, err := os.ReadFile("../../.taskfiles/local.yml")
	if err != nil {
		t.Fatalf("read local Taskfile: %v", err)
	}
	build := taskBlock(t, string(document), "build", "project")
	if strings.Contains(build, "build --pull") {
		t.Fatal("normal local:build must not require registry access")
	}
	if !strings.Contains(
		build,
		"build foundation-local-release local-release",
	) {
		t.Fatal("normal local:build must build the local images in one cache-sharing bake")
	}
	if strings.Contains(string(document), "compose.satellites.yml") ||
		strings.Contains(build, "mcp-service") || strings.Contains(build, "ai-service") {
		t.Fatal("local COMPOSE still builds AI or MCP satellites")
	}
	inventory, err := os.ReadFile("../../.taskfiles/local-inventory.yml")
	if err != nil {
		t.Fatalf("read local inventory Taskfile: %v", err)
	}
	if strings.Contains(string(inventory), "compose.satellites.yml") {
		t.Fatal("local inventory COMPOSE still includes AI or MCP satellites")
	}

	operations, err := os.ReadFile("../../.taskfiles/local-operations.yml")
	if err != nil {
		t.Fatalf("read local operations Taskfile: %v", err)
	}
	refresh := taskBlock(t, string(operations), "build:refresh", "check")
	if !strings.Contains(
		refresh,
		"build --pull foundation-local-release local-release",
	) {
		t.Fatal("explicit build refresh must pull the two release images together")
	}
}

func TestLocalDockerfilePinsFrontendAndBaseImages(t *testing.T) {
	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("read local Dockerfile: %v", err)
	}
	lines := strings.Split(string(content), "\n")
	const frontend = "# syntax=docker/dockerfile:1.12@sha256:" +
		"93bfd3b68c109427185cd78b4779fc82b484b0b7618e36d0f104d4d801e66d25"
	if len(lines) == 0 || lines[0] != frontend {
		t.Fatal("local Dockerfile frontend must use the pinned BuildKit digest")
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "FROM ") && !strings.Contains(line, "@sha256:") {
			t.Fatalf("local Dockerfile base image is not digest-pinned: %q", line)
		}
	}
}

func TestDestructiveLocalOperationsRunFromRepositoryRoot(t *testing.T) {
	document, err := os.ReadFile("../../.taskfiles/local-operations.yml")
	if err != nil {
		t.Fatalf("read local operations Taskfile: %v", err)
	}
	destroy := taskBlock(t, string(document), "destroy", "cleanup")
	for _, required := range []string{
		`dir: "{{.ROOT_DIR}}"`,
		"down --volumes --remove-orphans",
		"go run ./tools/localconfig",
	} {
		if !strings.Contains(destroy, required) {
			t.Fatalf("local:destroy is missing %q", required)
		}
	}
}

func taskBlock(t *testing.T, document, name, next string) string {
	t.Helper()
	startMarker := "\n  " + name + ":\n"
	endMarker := "\n  " + next + ":\n"
	start := strings.Index(document, startMarker)
	if start < 0 {
		t.Fatalf("task %q is missing", name)
	}
	start += len(startMarker)
	end := strings.Index(document[start:], endMarker)
	if end < 0 {
		t.Fatalf("task %q boundary is missing", next)
	}
	return document[start : start+end]
}
