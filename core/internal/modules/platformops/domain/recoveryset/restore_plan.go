package recoveryset

import (
	"bytes"
	"encoding/json"
	"io"
)

func BuildRestorePlan(request RestoreRequest, recoverySetID, manifestSHA256 string) ([]byte, error) {
	request.RecoverySetID = recoverySetID
	if ValidateRestore(request) != nil ||
		!digestPattern.MatchString(manifestSHA256) {
		return nil, ErrRestoreGate
	}
	payload := RestorePlanPayload{Schema: RestorePlanSchema,
		RecoverySetID: recoverySetID, ManifestSHA256: manifestSHA256,
		Identity: request.Manifest.Identity, TargetDisposition: request.Disposition,
		SafetyRecoverySetID:     request.Safety.RecoverySetID,
		SafetyRecoverySetSHA256: request.Safety.RecoverySetSHA256,
		WaiverReason:            request.Safety.WaiverReason, ProbeSHA256: request.Safety.ProbeSHA256,
		DestructiveConfirmed: request.Safety.DestructiveConfirmed}
	return canonicalPayload(payload)
}

func ParseRestorePlan(document []byte) (RestorePlanPayload, error) {
	var payload RestorePlanPayload
	if strictPayload(document, &payload) != nil || !payload.valid() {
		return RestorePlanPayload{}, ErrRestoreGate
	}
	return payload, nil
}

func (payload RestorePlanPayload) valid() bool {
	if payload.Schema != RestorePlanSchema || !validUUID(payload.RecoverySetID) ||
		!digestPattern.MatchString(payload.ManifestSHA256) || !payload.Identity.valid() {
		return false
	}
	evidence := SafetyEvidence{RecoverySetID: payload.SafetyRecoverySetID,
		RecoverySetSHA256: payload.SafetyRecoverySetSHA256,
		WaiverReason:      payload.WaiverReason, ProbeSHA256: payload.ProbeSHA256,
		DestructiveConfirmed: payload.DestructiveConfirmed}
	switch payload.TargetDisposition {
	case TargetExistingReadable:
		return validUUID(evidence.RecoverySetID) &&
			digestPattern.MatchString(evidence.RecoverySetSHA256) &&
			evidence.WaiverReason == "" && evidence.ProbeSHA256 == "" && !evidence.DestructiveConfirmed
	case TargetEmpty:
		return validWaiver(evidence, "empty_no_database_or_policy", false)
	case TargetUnreadable:
		return validWaiver(evidence, "unreadable_integrity_probe", true)
	default:
		return false
	}
}

func canonicalPayload(value any) ([]byte, error) {
	return canonicalObject(value, MaximumManifestBytes)
}

func strictPayload(document []byte, target any) error {
	if len(document) == 0 || len(document) > MaximumManifestBytes {
		return ErrManifest
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil || decoder.Decode(new(any)) != io.EOF {
		return ErrManifest
	}
	canonical, err := canonicalPayload(target)
	if err != nil || !bytes.Equal(canonical, document) {
		return ErrManifest
	}
	return nil
}
