package main

import (
	"fmt"
	"strings"
)

func addRecoveryCredentialArtifacts(
	set *artifactSet,
	credentials []recoveryCredential,
) error {
	rootAccess, rootSecret, err := newRecoveryRootCredential(set, credentials)
	if err != nil {
		return err
	}
	rootFiles := map[string]string{
		"recovery-minio-root-access-key.txt": rootAccess,
		"recovery-minio-root-secret-key.txt": rootSecret,
	}
	for name, value := range rootFiles {
		if err := set.add(name, []byte(value)); err != nil {
			return err
		}
	}
	var users strings.Builder
	for _, credential := range credentials {
		name := "recovery-" + credential.purpose + "-credential.json"
		if err := addJSON(set, name, map[string]string{
			"schema":            "recovery-s3-credential.v1",
			"access_key_id":     credential.accessKey,
			"secret_access_key": credential.secretKey,
		}); err != nil {
			return err
		}
		key := strings.ToUpper(strings.ReplaceAll(credential.purpose, "-", "_"))
		fmt.Fprintf(&users, "%s_ACCESS_KEY_ID=%s\n", key, credential.accessKey)
		fmt.Fprintf(&users, "%s_SECRET_ACCESS_KEY=%s\n", key, credential.secretKey)
	}
	return set.add("recovery-s3-users.env", []byte(users.String()))
}

func newRecoveryRootCredential(
	set *artifactSet,
	credentials []recoveryCredential,
) (string, string, error) {
	used := map[string]bool{
		string(set.files["s3-access-key-id.txt"]):      true,
		string(set.files["s3-secret-access-key.txt"]):  true,
		string(set.files["minio-root-access-key.txt"]): true,
		string(set.files["minio-root-secret-key.txt"]): true,
	}
	for _, credential := range credentials {
		used[credential.accessKey], used[credential.secretKey] = true, true
	}
	for {
		accessKey, err := randomAccessKeyID()
		if err != nil {
			return "", "", err
		}
		secretKey, err := randomBase64(32)
		if err != nil {
			return "", "", err
		}
		if !used[accessKey] && !used[secretKey] {
			return accessKey, secretKey, nil
		}
	}
}
