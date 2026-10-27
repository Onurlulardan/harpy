package recoveryorchestration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	iamapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api"
	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

type restoreKeyCustodyEvidence struct {
	Schema                 string `json:"schema"`
	EnvironmentID          string `json:"environment_id"`
	RecoverySetID          string `json:"recovery_set_id"`
	ProviderKind           string `json:"provider_kind"`
	ProviderHandle         string `json:"provider_handle"`
	ProviderGeneration     int64  `json:"provider_generation"`
	ProviderMaterialSHA256 string `json:"provider_material_sha256"`
	DRKEKID                string `json:"dr_kek_id"`
	RecoveryKeyID          string `json:"recovery_key_id"`
}

type restoreOwnerAnchorEvidence struct {
	Schema        string                  `json:"schema"`
	EnvironmentID string                  `json:"environment_id"`
	RecoverySetID string                  `json:"recovery_set_id"`
	Anchors       []iamapi.RecoveryAnchor `json:"anchors"`
}

type restoreProtectedAuthorityEvidence struct {
	Schema        string                    `json:"schema"`
	EnvironmentID string                    `json:"environment_id"`
	RecoverySetID string                    `json:"recovery_set_id"`
	Head          authorityapi.RecoveryHead `json:"head"`
}

func parseCanonicalRestoreEvidence(document []byte, target any) error {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, int(maximumRestoreEvidenceBytes),
	)
	if err != nil || !bytes.Equal(canonical, document) {
		return fmt.Errorf("restore evidence is not canonical")
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil || decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("restore evidence is invalid")
	}
	return nil
}

func validateOwnerAnchorEvidence(
	value restoreOwnerAnchorEvidence,
	identity recoveryset.EnvironmentIdentity,
	recoverySetID string,
) error {
	if value.Schema != "iam-owner-recovery-anchor-evidence.v1" ||
		value.EnvironmentID != identity.EnvironmentID ||
		value.RecoverySetID != recoverySetID || len(value.Anchors) == 0 {
		return fmt.Errorf("owner anchor recovery evidence differs")
	}
	for _, anchor := range value.Anchors {
		if anchor.EnvironmentID != identity.EnvironmentID ||
			anchor.TenantID != identity.PrimaryBindingID ||
			anchor.OwnerBindingID != identity.OwnerBindingID ||
			anchor.OwnerBindingGeneration != identity.OwnerBindingGeneration ||
			anchor.AnchorID == "" || anchor.OwnerUserID == "" ||
			len(anchor.CredentialID) == 0 || len(anchor.PublicKeyCOSE) == 0 {
			return fmt.Errorf("owner anchor recovery evidence differs")
		}
	}
	return nil
}

func validateProtectedAuthorityEvidence(
	value restoreProtectedAuthorityEvidence,
	identity recoveryset.EnvironmentIdentity,
	recoverySetID, trustSHA256 string,
) error {
	head := value.Head
	if value.Schema != "protected-authority-recovery-evidence.v1" ||
		value.EnvironmentID != identity.EnvironmentID ||
		value.RecoverySetID != recoverySetID ||
		head.EnvironmentID != identity.EnvironmentID ||
		head.DeploymentMode != "onpremise" ||
		head.RecoveryTrustBundleSHA256 != trustSHA256 ||
		head.ScopeKind != "environment" || head.ScopeID != identity.EnvironmentID ||
		head.Generation < 1 || head.RegistrySequence < 1 ||
		!matchesEvidenceDigest(head.StateSHA256, head.StateBytes) ||
		!matchesEvidenceDigest(head.ProposalSHA256, head.ProposalBytes) ||
		!matchesEvidenceDigest(head.AckSHA256, head.AckBytes) {
		return fmt.Errorf("protected authority recovery evidence differs")
	}
	return nil
}

func matchesEvidenceDigest(expected string, document []byte) bool {
	digest := sha256.Sum256(document)
	return expected == hex.EncodeToString(digest[:])
}
