package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStaticInventoryRejectsLegacyDeploymentAuthority(t *testing.T) {
	workspace := staticWorkspaceFixture(t)
	compose := filepath.Join(workspace, "backend-v2", "docker-compose.prod.yml")
	writeInventoryFixture(t, compose, "services:\n  api:\n    environment:\n      SUPER_ADMIN_PASSWORD: unsafe\n", 0o600)
	if _, err := runStatic([]string{"-workspace", workspace}); err == nil {
		t.Fatal("legacy deployment secret was accepted")
	}
	writeInventoryFixture(t, compose, "services:\n  api:\n    environment:\n      DEPLOYMENT_MODE: onpremise\n", 0o600)
	if _, err := runStatic([]string{"-workspace", workspace}); err != nil {
		t.Fatal(err)
	}
}

func TestStaticInventoryRejectsBooleanAndBrowserPersistence(t *testing.T) {
	workspace := staticWorkspaceFixture(t)
	source := filepath.Join(workspace, "frontend-v3", "src", "access.ts")
	for _, content := range []string{
		"export const allowed = user.is_super_admin\n",
		"sessionStorage.setItem('initial-admin-password', password)\n",
	} {
		writeInventoryFixture(t, source, content, 0o600)
		if _, err := runStatic([]string{"-workspace", workspace}); err == nil {
			t.Fatalf("unsafe source was accepted: %s", content)
		}
	}
}

func staticWorkspaceFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, path := range []string{
		"backend-v2/go.mod", "frontend-v3/package.json",
	} {
		writeInventoryFixture(t, filepath.Join(root, filepath.FromSlash(path)), "fixture", 0o600)
	}
	if err := os.MkdirAll(filepath.Join(root, "frontend-v3", "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	return root
}
