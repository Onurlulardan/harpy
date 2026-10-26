package iamrestorev1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type rotationBody struct {
	OperationID        string   `json:"operation_id"`
	Status             string   `json:"status"`
	Generation         int64    `json:"generation"`
	SecurityEpoch      string   `json:"security_epoch"`
	PublicKIDs         []string `json:"public_kids"`
	RevokedGenerations []int64  `json:"revoked_generations"`
}

func rotationDigest(value RotationResult) string {
	body := rotationBody{OperationID: value.OperationID, Status: value.Status,
		Generation: value.Generation, SecurityEpoch: value.SecurityEpoch,
		PublicKIDs: value.PublicKIDs, RevokedGenerations: value.RevokedGenerations}
	document, err := json.Marshal(body)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}

// RotationDigest computes the provider result digest for manifest writers.
func RotationDigest(value RotationResult) string { return rotationDigest(value) }
