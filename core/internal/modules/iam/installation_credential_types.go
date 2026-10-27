package iam

import (
	"errors"

	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
)

var ErrInstallationCredentialInvalid = errors.New(
	"IAM installation credential is invalid",
)

type PreparedInstallationCredential struct {
	value installationcredential.Credential
}

func (credential PreparedInstallationCredential) Valid() bool {
	return credential.value.Valid()
}

type InstallationCredentialInput struct {
	OperationID, EnvironmentID [16]byte
	TenantID, UserID           [16]byte
	Credential                 PreparedInstallationCredential
}

type InstallationRecoveryBinding struct {
	OperationID, TenantID, FirstOwnerUserID         [16]byte
	CurrentCredentialID, NextCredentialID           [16]byte
	CurrentAttemptID, NextAttemptID                 [16]byte
	CurrentCredentialVersion, NextCredentialVersion int64
	Cause, ManifestSHA256                           string
}
