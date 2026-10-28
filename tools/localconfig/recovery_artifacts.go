package main

import (
	"fmt"
	"time"
)

func addRecoveryArtifacts(set *artifactSet, _ time.Time) error {
	if set.options.profile != profileLocal {
		return nil
	}
	credentials, err := newRecoveryCredentials(set)
	if err != nil {
		return err
	}
	if err := addRecoveryCredentialArtifacts(set, credentials); err != nil {
		return err
	}
	if err := addRecoveryPostgresArtifacts(set); err != nil {
		return err
	}
	if err := addRecoveryStorageManifest(set); err != nil {
		return err
	}
	if err := addRecoveryProfileArtifacts(set); err != nil {
		return err
	}
	return nil
}

type recoveryCredential struct {
	purpose   string
	accessKey string
	secretKey string
}

func newRecoveryCredentials(set *artifactSet) ([]recoveryCredential, error) {
	purposes := []string{
		"write", "retention-delete", "restore-read",
		"documents-source-read", "documents-rehydrate-write",
	}
	used := map[string]bool{}
	for _, name := range []string{
		"s3-access-key-id.txt", "s3-secret-access-key.txt",
		"minio-root-access-key.txt", "minio-root-secret-key.txt",
	} {
		used[string(set.files[name])] = true
	}
	result := make([]recoveryCredential, 0, len(purposes))
	for _, purpose := range purposes {
		credential, err := uniqueRecoveryCredential(purpose, used)
		if err != nil {
			return nil, err
		}
		result = append(result, credential)
	}
	return result, nil
}

func uniqueRecoveryCredential(
	purpose string,
	used map[string]bool,
) (recoveryCredential, error) {
	for {
		accessKey, err := randomAccessKeyID()
		if err != nil {
			return recoveryCredential{}, err
		}
		secretKey, err := randomBase64(32)
		if err != nil {
			return recoveryCredential{}, err
		}
		if !used[accessKey] && !used[secretKey] {
			used[accessKey], used[secretKey] = true, true
			return recoveryCredential{
				purpose: purpose, accessKey: accessKey, secretKey: secretKey,
			}, nil
		}
	}
}

func recoveryCredentialByPurpose(
	credentials []recoveryCredential,
	purpose string,
) (recoveryCredential, error) {
	for _, credential := range credentials {
		if credential.purpose == purpose {
			return credential, nil
		}
	}
	return recoveryCredential{}, fmt.Errorf("recovery credential %s is missing", purpose)
}
