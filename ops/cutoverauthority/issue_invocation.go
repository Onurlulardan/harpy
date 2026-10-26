package main

import (
	"fmt"

	"github.com/google/uuid"
)

type finalizationPrerequisitePaths struct {
	authorityArtifactPaths
	AuthorityReceipt          string `json:"authority_receipt"`
	AuthorityReceiptSignature string `json:"authority_receipt_signature"`
	CredentialLease           string `json:"credential_lease"`
	CredentialLeaseSignature  string `json:"credential_lease_signature"`
	SessionSealReceipt        string `json:"session_seal_receipt"`
	CredentialCleanupReceipt  string `json:"credential_cleanup_receipt"`
	TerminalEvidence          string `json:"terminal_evidence"`
}

type finalizerIssueInvocation struct {
	ClusterAdminDSNFile string `json:"cluster_admin_dsn_file"`
	FinalizerDSNFile    string `json:"finalizer_dsn_file"`
	finalizationPrerequisitePaths
	SigningKeyFile              string `json:"signing_key_file"`
	FinalizerLeaseID            string `json:"finalizer_lease_id"`
	FinalizerLeasePath          string `json:"finalizer_lease_path"`
	FinalizerLeaseSignaturePath string `json:"finalizer_lease_signature_path"`
}

func (value finalizerIssueInvocation) validate() error {
	leaseID, err := uuid.Parse(value.FinalizerLeaseID)
	if err != nil || leaseID == uuid.Nil {
		return fmt.Errorf("finalizer lease identity is invalid")
	}
	paths := value.finalizationPrerequisitePaths.all()
	paths = append(paths, value.ClusterAdminDSNFile, value.FinalizerDSNFile,
		value.SigningKeyFile,
		value.FinalizerLeasePath, value.FinalizerLeaseSignaturePath)
	return validateAbsoluteDistinctPaths(paths)
}

func (value finalizationPrerequisitePaths) all() []string {
	paths := value.authorityArtifactPaths.all()
	return append(paths, value.AuthorityReceipt, value.AuthorityReceiptSignature,
		value.CredentialLease, value.CredentialLeaseSignature,
		value.SessionSealReceipt, value.CredentialCleanupReceipt,
		value.TerminalEvidence)
}
