package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const invocationContract = "external-cutover-authority-config.v1"

type invocation struct {
	ContractVersion       string                    `json:"contract_version"`
	Action                string                    `json:"action"`
	Environment           string                    `json:"environment"`
	DeadlineSeconds       int                       `json:"deadline_seconds"`
	Adopt                 *adoptionInvocation       `json:"adopt"`
	IssueFinalizer        *finalizerIssueInvocation `json:"issue_finalizer"`
	AuthorizeFinalization *authorizationInvocation  `json:"authorize_finalization"`
	RecordFullRollout     *fullRolloutInvocation    `json:"record_full_rollout"`
}

type authorityArtifactPaths struct {
	TrustRoot                   string `json:"trust_root"`
	ReleaseManifest             string `json:"release_manifest"`
	ReleaseSignature            string `json:"release_signature"`
	BootstrapAuthority          string `json:"bootstrap_authority_manifest"`
	BootstrapAuthoritySignature string `json:"bootstrap_authority_signature"`
	AuthorityOperation          string `json:"authority_operation"`
	AuthorityOperationSignature string `json:"authority_operation_signature"`
}

type adoptionInvocation struct {
	RootDSNFile string `json:"root_dsn_file"`
	authorityArtifactPaths
	SigningKeyFile       string `json:"signing_key_file"`
	ReceiptPath          string `json:"receipt_path"`
	ReceiptSignaturePath string `json:"receipt_signature_path"`
}

type authorizationInvocation struct {
	ClusterAdminDSNFile string `json:"cluster_admin_dsn_file"`
	finalizationPrerequisitePaths
	FinalizerLease          string `json:"finalizer_lease"`
	FinalizerLeaseSignature string `json:"finalizer_lease_signature"`
}

func loadInvocation(path string) (invocation, error) {
	if !filepath.IsAbs(path) {
		return invocation{}, fmt.Errorf("config path must be absolute")
	}
	document, err := os.ReadFile(path)
	if err != nil || len(document) == 0 || len(document) > 64*1024 {
		return invocation{}, fmt.Errorf("config document is unavailable or invalid")
	}
	var value invocation
	decoder := json.NewDecoder(bytesReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return invocation{}, fmt.Errorf("config document is invalid")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return invocation{}, fmt.Errorf("config document is invalid")
	}
	if err := value.validate(); err != nil {
		return invocation{}, err
	}
	return value, nil
}

func (value invocation) validate() error {
	if value.ContractVersion != invocationContract || value.DeadlineSeconds < 1 ||
		value.DeadlineSeconds > 900 || !validEnvironment(value.Environment) {
		return fmt.Errorf("config header is invalid")
	}
	switch value.Action {
	case "adopt_authority":
		if value.Adopt == nil || value.IssueFinalizer != nil ||
			value.AuthorizeFinalization != nil || value.RecordFullRollout != nil {
			return fmt.Errorf("adoption config profile is invalid")
		}
		return value.Adopt.validate()
	case "issue_finalizer":
		if value.Adopt != nil || value.IssueFinalizer == nil ||
			value.AuthorizeFinalization != nil || value.RecordFullRollout != nil ||
			value.DeadlineSeconds > 120 {
			return fmt.Errorf("finalizer issuance config profile is invalid")
		}
		return value.IssueFinalizer.validate()
	case "authorize_finalization":
		if value.Adopt != nil || value.IssueFinalizer != nil ||
			value.AuthorizeFinalization == nil || value.RecordFullRollout != nil ||
			value.DeadlineSeconds > 120 {
			return fmt.Errorf("finalization authorization profile is invalid")
		}
		return value.AuthorizeFinalization.validate()
	case "record_full_rollout":
		if value.Adopt != nil || value.IssueFinalizer != nil ||
			value.AuthorizeFinalization != nil || value.RecordFullRollout == nil ||
			value.DeadlineSeconds > 120 {
			return fmt.Errorf("full rollout config profile is invalid")
		}
		return value.RecordFullRollout.validate()
	default:
		return fmt.Errorf("config action is invalid")
	}
}

func validEnvironment(value string) bool {
	return value == "development" || value == "test" ||
		value == "staging" || value == "production"
}

func (value adoptionInvocation) validate() error {
	paths := value.authorityArtifactPaths.all()
	paths = append(paths, value.RootDSNFile, value.SigningKeyFile,
		value.ReceiptPath, value.ReceiptSignaturePath)
	return validateAbsoluteDistinctPaths(paths)
}

func (value authorizationInvocation) validate() error {
	paths := value.finalizationPrerequisitePaths.all()
	paths = append(paths, value.ClusterAdminDSNFile, value.FinalizerLease,
		value.FinalizerLeaseSignature)
	return validateAbsoluteDistinctPaths(paths)
}

func (value authorityArtifactPaths) all() []string {
	return []string{value.TrustRoot, value.ReleaseManifest, value.ReleaseSignature,
		value.BootstrapAuthority, value.BootstrapAuthoritySignature,
		value.AuthorityOperation, value.AuthorityOperationSignature}
}

func validateAbsoluteDistinctPaths(paths []string) error {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" || !filepath.IsAbs(path) {
			return fmt.Errorf("all artifact paths must be absolute")
		}
		clean := filepath.Clean(path)
		if _, exists := seen[clean]; exists {
			return fmt.Errorf("artifact paths must be distinct")
		}
		seen[clean] = struct{}{}
	}
	return nil
}
