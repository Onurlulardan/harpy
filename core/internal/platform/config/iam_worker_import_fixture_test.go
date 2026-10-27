package config

import (
	"strings"
	"testing"
	"time"
)

func TestIAMWorkerRequiresImportInvitationAuthority(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	entries := validWorkerEnvironmentAt(now)
	withoutOneTime := make([]string, 0, len(entries)-1)
	for _, entry := range entries {
		if !strings.HasPrefix(entry, iamOneTimeKeyringEnvironmentKey+"=") {
			withoutOneTime = append(withoutOneTime, entry)
		}
	}
	if _, err := loadWorkerAt(withoutOneTime, now); err == nil {
		t.Fatal("worker accepted a missing one-time invitation keyring")
	}
	withoutDelivery := make([]string, 0, len(entries)-1)
	for _, entry := range entries {
		if !strings.HasPrefix(entry, iamDeliveryKMSKeyIDEnvironmentKey+"=") {
			withoutDelivery = append(withoutDelivery, entry)
		}
	}
	if _, err := loadWorkerAt(withoutDelivery, now); err != nil {
		t.Fatal(err)
	}
}

func TestIAMWorkerSeparatesDirectoryAndInvitationKeyMaterial(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	entries := withEnvironment(validWorkerEnvironmentAt(now),
		iamOneTimeKeyringEnvironmentKey+"="+
			validIAMHMACKeyringJSON(now, iamPurposeOneTime, 22),
	)
	if _, err := loadWorkerAt(entries, now); err == nil {
		t.Fatal("worker accepted shared directory and one-time key material")
	}
}

func TestIAMWorkerExposesValidatedImportInvitationKeys(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	worker, err := loadWorkerAt(validWorkerEnvironmentAt(now), now)
	if err != nil {
		t.Fatal(err)
	}
	iam, err := worker.IAM()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = iam.OneTimeTokenKeyring(); err != nil {
		t.Fatal(err)
	}
	if _, err = iam.DeliveryKMSKeyID(); err != nil {
		t.Fatal(err)
	}
}
