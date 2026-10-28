package main

func writeAuthorityMaterial(dir, mode string) ([]string, error) {
	manifest, err := writeDurableFile(dir, "installation.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	lines := []string{runtimeLine("HARPY_INSTALLATION_MANIFEST_PATH", manifest)}
	if mode == "saas" {
		saas, saasErr := writeSaaSAuthority(dir)
		if saasErr != nil {
			return nil, saasErr
		}
		return append(lines, saas...), nil
	}
	onPrem, err := writeOnPremAuthority(dir)
	if err != nil {
		return nil, err
	}
	return append(lines, onPrem...), nil
}

func writeSaaSAuthority(dir string) ([]string, error) {
	ack, err := writeDurableFile(dir, "ack-trust.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	release, err := writeDurableFile(dir, "release-trust.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	return []string{
		runtimeLine("HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH", ack),
		runtimeLine("HARPY_RELEASE_TRUST_ROOT_PATH", release),
		runtimeLine("HARPY_SAAS_AUTHORITY_REGISTRY_URL", "http://127.0.0.1:8443"),
	}, nil
}

func writeOnPremAuthority(dir string) ([]string, error) {
	recovery, err := writeDurableFile(dir, "recovery-trust.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	control, err := writeSigningKey(dir, "recovery-control.key", 31)
	if err != nil {
		return nil, err
	}
	ack, err := writeDurableFile(dir, "ack-trust.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	license, err := writeDurableFile(dir, "license.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	licenseTrust, err := writeDurableFile(dir, "license-trust.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	release, err := writeDurableFile(dir, "release-trust.json", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	journal, err := writeDurableFile(dir, "authority-head.jsonl", []byte("{}\n"))
	if err != nil {
		return nil, err
	}
	ackKey, err := writeSigningKey(dir, "authority-ack.key", 32)
	if err != nil {
		return nil, err
	}
	signingDir, err := runtimeDirectory(dir, "recovery-signing")
	if err != nil {
		return nil, err
	}
	return []string{
		runtimeLine("HARPY_RECOVERY_TRUST_BUNDLE_PATH", recovery),
		runtimeLine("HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE", control),
		runtimeLine("HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH", ack),
		runtimeLine("HARPY_LICENSE_SNAPSHOT_FILE", license),
		runtimeLine("HARPY_LICENSE_TRUST_BUNDLE_PATH", licenseTrust),
		runtimeLine("HARPY_RELEASE_TRUST_ROOT_PATH", release),
		runtimeLine("HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH", journal),
		runtimeLine("HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE", ackKey),
		runtimeLine("HARPY_RECOVERY_SIGNING_DIRECTORY", signingDir),
	}, nil
}

func writeSigningKey(dir, name string, marker byte) (string, error) {
	material, err := markedMaterial(marker)
	if err != nil {
		return "", err
	}
	return writeRuntimeFile(dir, name, material[:])
}
