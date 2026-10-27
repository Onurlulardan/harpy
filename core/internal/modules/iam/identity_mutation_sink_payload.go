package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
)

func identityChangeEventPayload(
	record changeapp.RecordedChange,
) ([]byte, string, error) {
	raw, err := json.Marshal(struct {
		Clock    int64  `json:"clock"`
		Command  string `json:"command"`
		Resource string `json:"resource"`
		Result   string `json:"result"`
		Revision int64  `json:"revision"`
	}{record.Clock, record.CommandKey, record.ResourceKind, record.ResultCode, record.Revision})
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
