package recoveryorchestration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func restoreSafetyDigest(claim ports.RestoreExecutionClaim) (string, error) {
	raw, err := json.Marshal(struct {
		Schema                  string `json:"schema"`
		TargetDisposition       string `json:"target_disposition"`
		SafetyRecoverySetID     string `json:"safety_recovery_set_id,omitempty"`
		SafetyRecoverySetSHA256 string `json:"safety_recovery_set_sha256,omitempty"`
		WaiverReason            string `json:"waiver_reason,omitempty"`
		ProbeSHA256             string `json:"probe_sha256,omitempty"`
		DestructiveConfirmed    bool   `json:"destructive_confirmed"`
	}{
		Schema:                  "restore-safety-evidence.v1",
		TargetDisposition:       claim.TargetDisposition,
		SafetyRecoverySetID:     claim.SafetyRecoverySetID,
		SafetyRecoverySetSHA256: claim.SafetyRecoverySetSHA256,
		WaiverReason:            claim.WaiverReason, ProbeSHA256: claim.ProbeSHA256,
		DestructiveConfirmed: claim.DestructiveConfirmed,
	})
	if err != nil {
		return "", err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, 64*1024)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
