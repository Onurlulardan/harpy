package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

func invitationEventPayload(
	change invitation.AuditChange,
) ([]byte, string, error) {
	var artifactReference *string
	if change.DeliveryArtifactRef != "" {
		artifactReference = &change.DeliveryArtifactRef
	}
	raw, err := json.Marshal(struct {
		Command             string  `json:"command"`
		DeliveryArtifactRef *string `json:"delivery_artifact_ref"`
		Revision            int64   `json:"revision"`
		State               string  `json:"state"`
	}{
		Command: change.CommandKey, DeliveryArtifactRef: artifactReference,
		Revision: change.ResultRevision, State: string(change.ResultState),
	})
	if err != nil {
		return nil, "", err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		raw, eventsv1.MaximumCanonicalPayloadBytes,
	)
	if err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(canonical)
	return canonical, hex.EncodeToString(digest[:]), nil
}
