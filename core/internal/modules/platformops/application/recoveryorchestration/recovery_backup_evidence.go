package recoveryorchestration

import (
	"encoding/json"
	"fmt"
	"sort"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	iamapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api"
	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverykeymaterial"
)

const maximumRecoveryEvidenceBytes = 1024 * 1024

func buildKeyCustodyEvidence(
	environmentID, recoverySetID string,
	material recoverykeymaterial.Material,
) ([]byte, error) {
	return canonicalRecoveryEvidence(map[string]any{
		"schema":                   "recovery-key-custody-evidence.v1",
		"environment_id":           environmentID,
		"recovery_set_id":          recoverySetID,
		"provider_kind":            material.ProviderKind(),
		"provider_handle":          material.ProviderHandle(),
		"provider_generation":      material.Generation(),
		"provider_material_sha256": material.DocumentSHA256(),
		"dr_kek_id":                material.DRKeyID(),
		"recovery_key_id":          material.RecoveryKeyID(),
	})
}

func buildOwnerAnchorEvidence(
	environmentID, recoverySetID, ownerBindingID string,
	ownerBindingGeneration int64,
	anchors []iamapi.RecoveryAnchor,
) ([]byte, error) {
	if len(anchors) == 0 {
		return nil, fmt.Errorf("owner recovery anchors are unavailable")
	}
	sort.Slice(anchors, func(left, right int) bool {
		return anchors[left].AnchorID < anchors[right].AnchorID
	})
	for _, anchor := range anchors {
		if anchor.EnvironmentID != environmentID ||
			anchor.OwnerBindingID != ownerBindingID ||
			anchor.OwnerBindingGeneration != ownerBindingGeneration {
			return nil, fmt.Errorf("owner recovery anchor binding differs")
		}
	}
	return canonicalRecoveryEvidence(struct {
		Schema        string                  `json:"schema"`
		EnvironmentID string                  `json:"environment_id"`
		RecoverySetID string                  `json:"recovery_set_id"`
		Anchors       []iamapi.RecoveryAnchor `json:"anchors"`
	}{
		Schema:        "iam-owner-recovery-anchor-evidence.v1",
		EnvironmentID: environmentID, RecoverySetID: recoverySetID,
		Anchors: anchors,
	})
}

func buildProtectedAuthorityEvidence(
	environmentID, recoverySetID, trustSHA256 string,
	head authorityapi.RecoveryHead,
) ([]byte, error) {
	if head.EnvironmentID != environmentID ||
		head.DeploymentMode != "onpremise" ||
		head.RecoveryTrustBundleSHA256 != trustSHA256 ||
		len(head.AckBytes) == 0 || len(head.StateBytes) == 0 ||
		len(head.ProposalBytes) == 0 {
		return nil, fmt.Errorf("protected authority recovery head differs")
	}
	return canonicalRecoveryEvidence(struct {
		Schema        string                    `json:"schema"`
		EnvironmentID string                    `json:"environment_id"`
		RecoverySetID string                    `json:"recovery_set_id"`
		Head          authorityapi.RecoveryHead `json:"head"`
	}{
		Schema:        "protected-authority-recovery-evidence.v1",
		EnvironmentID: environmentID, RecoverySetID: recoverySetID,
		Head: head,
	})
}

func canonicalRecoveryEvidence(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(raw, maximumRecoveryEvidenceBytes)
}
