package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecondSetupKeepsInstalledMaterial(t *testing.T) {
	root := t.TempDir()
	out := runOK(t, root, []string{"--defaults"}, "")
	env := readEnv(t, root, "single-tenant")
	assertNoLeak(t, out, env)
	secrets := filepath.Join(root, "deploy", "single-tenant", "secrets")
	assertStdoutOmitsFile(t, out, filepath.Join(secrets, "runtime", "cursor-keyring.json"))
	assertStdoutOmitsFile(t, out, filepath.Join(secrets, "release", "trust-root.json"))
	first := secretFingerprints(t, root, "single-tenant")
	runOK(t, root, []string{"--defaults"}, "")
	assertSameFingerprints(t, first, secretFingerprints(t, root, "single-tenant"))

	install := filepath.Join(secrets, "runtime", "installation.json")
	if err := os.WriteFile(install, []byte("{\"kept\":true}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	kept := fileFingerprint(t, install)
	cursor := filepath.Join(secrets, "runtime", "cursor-keyring.json")
	access := filepath.Join(secrets, "runtime", "access-token-keyring.json")
	trust := filepath.Join(secrets, "release", "trust-root.json")
	accessBefore := fileFingerprint(t, access)
	if err := os.Remove(cursor); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(trust); err != nil {
		t.Fatal(err)
	}
	runOK(t, root, []string{"--defaults"}, "")
	if _, err := os.Stat(cursor); err != nil {
		t.Fatal("missing runtime file was not recreated")
	}
	if _, err := os.Stat(trust); err != nil {
		t.Fatal("incomplete release set was not rebuilt")
	}
	if fileFingerprint(t, access) != accessBefore {
		t.Fatal("partial retry rotated an existing runtime key")
	}
	if fileFingerprint(t, install) != kept {
		t.Fatal("partial retry rewrote installation state")
	}

	beforeRotate := secretFingerprints(t, root, "single-tenant")
	rotatedOut := runOK(t, root, []string{"--defaults", "--rotate-secrets"}, "")
	rotatedEnv := readEnv(t, root, "single-tenant")
	assertNoLeak(t, rotatedOut, rotatedEnv)
	assertRotationMatrix(t, beforeRotate, secretFingerprints(t, root, "single-tenant"))
	if fileFingerprint(t, install) != kept {
		t.Fatal("rotate-secrets rewrote installation state")
	}
}

func TestRefreshReleaseKeepsRuntimeKeys(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults"}, "")
	base := filepath.Join(root, "deploy", "single-tenant", "secrets")
	access := filepath.Join(base, "runtime", "access-token-keyring.json")
	trust := filepath.Join(base, "release", "trust-root.json")
	accessBefore := fileFingerprint(t, access)
	trustBefore := fileFingerprint(t, trust)
	runOK(t, root, []string{"--defaults", "--refresh-release"}, "")
	if fileFingerprint(t, access) != accessBefore {
		t.Fatal("refresh-release rotated a runtime key")
	}
	if fileFingerprint(t, trust) == trustBefore {
		t.Fatal("refresh-release kept the release trust root")
	}
}

func TestActorBridgePairRewritesTogether(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults"}, "")
	base := filepath.Join(root, "deploy", "single-tenant", "secrets", "runtime")
	seed := filepath.Join(base, "actor-bridge-seed")
	public := filepath.Join(base, "actor-bridge-public-keys.json")
	seedBefore := fileFingerprint(t, seed)
	if err := os.Remove(public); err != nil {
		t.Fatal(err)
	}
	runOK(t, root, []string{"--defaults"}, "")
	if _, err := os.Stat(public); err != nil {
		t.Fatal(err)
	}
	if fileFingerprint(t, seed) == seedBefore {
		t.Fatal("deleted public key left the old seed in place")
	}
}

func TestMultiTenantSetupKeepsMaterial(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults", "--profile", "multi-tenant"}, "")
	first := secretFingerprints(t, root, "multi-tenant")
	runOK(t, root, []string{"--defaults", "--profile", "multi-tenant"}, "")
	assertSameFingerprints(t, first, secretFingerprints(t, root, "multi-tenant"))
}

func assertStdoutOmitsFile(t *testing.T, out, path string) {
	t.Helper()
	body := mustRead(t, path)
	if body != "" && strings.Contains(out, body) {
		t.Fatalf("stdout leaked %s", path)
	}
}
