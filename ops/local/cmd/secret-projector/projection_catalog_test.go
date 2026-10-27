package main

import (
	"strings"
	"testing"
)

func TestProjectionCatalogHasUniqueTargetsAndFiles(t *testing.T) {
	targets := make(map[string]bool)
	for _, target := range secretTargets() {
		if targets[target.name] {
			t.Fatalf("duplicate target %q", target.name)
		}
		targets[target.name] = true
		files := make(map[string]bool)
		for _, file := range target.files {
			if files[file.name] {
				t.Fatalf("target %s repeats %s", target.name, file.name)
			}
			files[file.name] = true
		}
	}
}

func TestRuntimeTargetsExcludePrivilegedMaterial(t *testing.T) {
	for _, target := range secretTargets() {
		if target.name != "api" && target.name != "worker" {
			continue
		}
		forbidden := map[string]bool{
			"local-ca.key": true, "kms-master-key.bin": true,
			"postgres-bootstrap-dsn.txt": true,
			"postgres-migrator-dsn.txt":  true,
			"minio-root-access-key.txt":  true,
			"minio-root-secret-key.txt":  true,
		}
		for _, file := range target.files {
			if forbidden[file.name] {
				t.Fatalf("target %s received privileged artifact %s", target.name, file.name)
			}
		}
	}
}

func TestRabbitMQProjectionUsesContainerIdentity(t *testing.T) {
	for _, target := range secretTargets() {
		if target.name == "rabbitmq" {
			if target.uid != rabbitMQUID || target.gid != rabbitMQGID {
				t.Fatalf("rabbitmq owner = %d:%d", target.uid, target.gid)
			}
			return
		}
	}
	t.Fatal("rabbitmq projection target is missing")
}

func TestReleaseStateTargetsAreSeparate(t *testing.T) {
	states := make(map[string]bool)
	for _, target := range stateTargets() {
		states[target.name] = true
	}
	if !states["foundation-release"] || !states["release"] {
		t.Fatal("foundation and current release state must use separate volumes")
	}
}

func TestRuntimeTargetsNeverReceiveRecoveryAuthority(t *testing.T) {
	for _, target := range secretTargets() {
		if target.name != "api" && target.name != "worker" {
			continue
		}
		for _, file := range target.files {
			if strings.HasPrefix(file.name, "recovery-") ||
				file.name == "recovery-storage-manifest.v1.json" {
				t.Fatalf("%s received recovery authority %s", target.name, file.name)
			}
		}
	}
}

func TestRecoveryCredentialProjectionIsExact(t *testing.T) {
	expected := map[string]string{
		"recovery-write-credential.json":                     "recovery-backup",
		"recovery-retention-delete-credential.json":          "recovery-backup",
		"recovery-documents-source-read-credential.json":     "recovery-backup",
		"recovery-restore-read-credential.json":              "recovery-restore",
		"recovery-documents-rehydrate-write-credential.json": "recovery-restore",
	}
	found := make(map[string]string, len(expected))
	for _, target := range secretTargets() {
		for _, file := range target.files {
			if owner, tracked := expected[file.name]; tracked {
				if found[file.name] != "" || owner != target.name {
					t.Fatalf("%s projected to %s, want only %s", file.name, target.name, owner)
				}
				found[file.name] = target.name
			}
		}
	}
	if len(found) != len(expected) {
		t.Fatalf("purpose credential projection count = %d, want %d", len(found), len(expected))
	}
}

func TestRecoveryStateTargetsAreDedicated(t *testing.T) {
	states := map[string]bool{}
	for _, target := range stateTargets() {
		states[target.name] = true
	}
	if !states["recovery-key-state"] || !states["restore-state"] {
		t.Fatal("recovery key and restore progress need separate state volumes")
	}
}
