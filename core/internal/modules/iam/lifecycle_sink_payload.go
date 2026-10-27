package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func lifecycleEventPayload(change lifecycle.Change) ([]byte, string, error) {
	raw, err := json.Marshal(struct {
		AggregateKind string `json:"aggregate_kind"`
		Command       string `json:"command"`
		Revision      int64  `json:"revision"`
	}{change.AggregateKind, change.CommandKey, change.Revision})
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
