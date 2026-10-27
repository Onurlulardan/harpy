package recoveryset

func BuildRestoreReceipt(payload RestoreReceiptPayload) ([]byte, error) {
	if !payload.valid() {
		return nil, ErrRestoreGate
	}
	return canonicalPayload(payload)
}

func ParseRestoreReceipt(document []byte) (RestoreReceiptPayload, error) {
	var payload RestoreReceiptPayload
	if strictPayload(document, &payload) != nil || !payload.valid() {
		return RestoreReceiptPayload{}, ErrRestoreGate
	}
	return payload, nil
}

func (payload RestoreReceiptPayload) valid() bool {
	previousValid := payload.PreviousDatabaseSystemID == "" ||
		validUUID(payload.PreviousDatabaseSystemID)
	return payload.Schema == RestoreReceiptSchema &&
		validUUID(payload.OperationID) && validUUID(payload.RecoverySetID) &&
		digestPattern.MatchString(payload.PlanSHA256) &&
		digestPattern.MatchString(payload.ManifestSHA256) && payload.Identity.valid() &&
		previousValid && validUUID(payload.NewDatabaseSystemID) &&
		payload.NewDatabaseSystemID != payload.PreviousDatabaseSystemID &&
		validDigests(payload.PreviousDatabaseLocatorSHA256,
			payload.NewDatabaseLocatorSHA256, payload.DrainAckSHA256) &&
		digestPattern.MatchString(payload.SwitchAckSHA256) &&
		payload.PreviousDatabaseLocatorSHA256 != payload.NewDatabaseLocatorSHA256 &&
		validDigests(payload.ObjectMerkleSHA256, payload.LocatorRemapSHA256,
			payload.SafetyEvidenceSHA256) && !payload.Seeded && !payload.OwnerChanged &&
		!payload.TenantReplaced && payload.SwitchOutcome == "switched_same_environment"
}
