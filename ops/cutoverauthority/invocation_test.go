package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadInvocationSeparatesExternalAuthorityProfiles(t *testing.T) {
	directory := t.TempDir()
	adopt := validInvocationFixture(t, directory, "adopt_authority")
	path := writeInvocationFixture(t, directory, "adopt.json", adopt)
	loaded, err := loadInvocation(path)
	if err != nil {
		t.Fatalf("load adoption config: %v", err)
	}
	if loaded.Adopt == nil || loaded.IssueFinalizer != nil ||
		loaded.AuthorizeFinalization != nil ||
		loaded.Adopt.RootDSNFile == "" {
		t.Fatal("adoption profile was not isolated")
	}
	authorize := validInvocationFixture(t, directory, "authorize_finalization")
	path = writeInvocationFixture(t, directory, "authorize.json", authorize)
	loaded, err = loadInvocation(path)
	if err != nil {
		t.Fatalf("load authorization config: %v", err)
	}
	if loaded.Adopt != nil || loaded.AuthorizeFinalization == nil ||
		loaded.AuthorizeFinalization.ClusterAdminDSNFile == "" {
		t.Fatal("authorization profile was not isolated")
	}
	issue := validInvocationFixture(t, directory, "issue_finalizer")
	path = writeInvocationFixture(t, directory, "issue.json", issue)
	loaded, err = loadInvocation(path)
	if err != nil {
		t.Fatalf("load issuance config: %v", err)
	}
	if loaded.Adopt != nil || loaded.IssueFinalizer == nil ||
		loaded.AuthorizeFinalization != nil ||
		loaded.RecordFullRollout != nil ||
		loaded.IssueFinalizer.ClusterAdminDSNFile == "" {
		t.Fatal("issuance profile was not isolated")
	}
	full := validInvocationFixture(t, directory, "record_full_rollout")
	path = writeInvocationFixture(t, directory, "full.json", full)
	loaded, err = loadInvocation(path)
	if err != nil || loaded.RecordFullRollout == nil ||
		loaded.Adopt != nil || loaded.IssueFinalizer != nil ||
		loaded.AuthorizeFinalization != nil {
		t.Fatalf("full rollout profile was not isolated: %v", err)
	}
}

func TestLoadInvocationRejectsUnknownTrailingAndCrossProfileFields(t *testing.T) {
	directory := t.TempDir()
	value := validInvocationFixture(t, directory, "adopt_authority")
	value["unknown"] = true
	path := writeInvocationFixture(t, directory, "unknown.json", value)
	if _, err := loadInvocation(path); err == nil {
		t.Fatal("unknown config field accepted")
	}
	value = validInvocationFixture(t, directory, "adopt_authority")
	value["authorize_finalization"] = map[string]any{}
	path = writeInvocationFixture(t, directory, "cross.json", value)
	if _, err := loadInvocation(path); err == nil {
		t.Fatal("cross-profile config accepted")
	}
	path = writeInvocationFixture(t, directory, "trailing.json",
		validInvocationFixture(t, directory, "adopt_authority"))
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.WriteString("{}")
	_ = file.Close()
	if _, err := loadInvocation(path); err == nil {
		t.Fatal("trailing config document accepted")
	}
}

func validInvocationFixture(t *testing.T, directory, action string) map[string]any {
	t.Helper()
	value := map[string]any{
		"contract_version": invocationContract, "action": action,
		"environment": "test", "deadline_seconds": 60,
	}
	artifacts := map[string]any{}
	for _, key := range []string{"trust_root", "release_manifest", "release_signature",
		"bootstrap_authority_manifest", "bootstrap_authority_signature",
		"authority_operation", "authority_operation_signature"} {
		artifacts[key] = filepath.Join(directory, action+"-"+key)
	}
	if action == "adopt_authority" {
		artifacts["root_dsn_file"] = filepath.Join(directory, "root-dsn")
		artifacts["signing_key_file"] = filepath.Join(directory, "signer")
		artifacts["receipt_path"] = filepath.Join(directory, "receipt")
		artifacts["receipt_signature_path"] = filepath.Join(directory, "receipt-signature")
		value["adopt"] = artifacts
		return value
	}
	if action == "record_full_rollout" {
		artifacts["cluster_admin_dsn_file"] = filepath.Join(directory, "full-dsn")
		artifacts["completion_evidence"] = filepath.Join(directory, "full-evidence")
		artifacts["evidence_artifacts"] = []string{
			filepath.Join(directory, "full-artifact"),
		}
		value["record_full_rollout"] = artifacts
		return value
	}
	commonKeys := []string{"cluster_admin_dsn_file", "authority_receipt",
		"authority_receipt_signature", "credential_lease", "credential_lease_signature",
		"session_seal_receipt", "credential_cleanup_receipt", "terminal_evidence"}
	for _, key := range commonKeys {
		artifacts[key] = filepath.Join(directory, action+"-"+key)
	}
	if action == "issue_finalizer" {
		artifacts["finalizer_dsn_file"] = filepath.Join(directory, "issue-finalizer-dsn")
		artifacts["signing_key_file"] = filepath.Join(directory, "issue-signer")
		artifacts["finalizer_lease_id"] = "99999999-9999-4999-8999-999999999999"
		artifacts["finalizer_lease_path"] = filepath.Join(directory, "issued-lease")
		artifacts["finalizer_lease_signature_path"] = filepath.Join(directory, "issued-signature")
		value["issue_finalizer"] = artifacts
		return value
	}
	keys := []string{"finalizer_lease", "finalizer_lease_signature"}
	for _, key := range keys {
		artifacts[key] = filepath.Join(directory, "authorize-"+key)
	}
	value["authorize_finalization"] = artifacts
	return value
}

func writeInvocationFixture(
	t *testing.T, directory, name string, value map[string]any,
) string {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
