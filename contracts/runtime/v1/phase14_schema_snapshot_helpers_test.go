package runtimev1

func addPhase14RuntimeSnapshot(profile string, variables map[string]any) {
	switch profile {
	case "api":
		variables["DEPLOYMENT_MODE"] = "saas"
		variables["HARPY_SAAS_AUTHORITY_PROPOSER_KMS_KEY_ID"] =
			"gcp-kms://projects/harpy-prod/locations/europe-west1/keyRings/authority/" +
				"cryptoKeys/proposer/cryptoKeyVersions/7"
		variables["HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"] = "/authority/ack-trust.json"
	case "worker":
		variables["HARPY_LICENSE_SNAPSHOT_FILE"] = "/license/dev-snapshot.json"
		variables["HARPY_LICENSE_TRUST_BUNDLE_PATH"] = "/license/trust.json"
		variables["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] = "/recovery/trust.json"
		variables["HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"] = "/authority/ack-trust.json"
		variables["HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH"] = "/authority/head.jsonl"
		variables["HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE"] = "/authority/ack.key"
	}
}

func apiOnPremContractSnapshot() map[string]any {
	snapshot := profileSnapshot("api")
	variables := snapshot["variables"].(map[string]any)
	delete(variables, "DEPLOYMENT_MODE")
	delete(variables, "HARPY_SAAS_AUTHORITY_PROPOSER_KMS_KEY_ID")
	variables["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] = "/recovery/trust.json"
	variables["HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE"] = "/recovery/control.key"
	return snapshot
}

func selectWorkerSaaSContractSnapshot(variables map[string]any) {
	for _, key := range []string{
		"HARPY_LICENSE_SNAPSHOT_FILE", "HARPY_LICENSE_TRUST_BUNDLE_PATH",
		"HARPY_RECOVERY_TRUST_BUNDLE_PATH", "HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH",
		"HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE",
	} {
		delete(variables, key)
	}
	variables["HARPY_SAAS_AUTHORITY_REGISTRY_URL"] = "https://authority.invalid"
}
