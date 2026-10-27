package releasepublication

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func validOwnerPayloads(
	manifest packagecontract.Manifest,
	payloads []catalogpackagev1.OwnerPayload,
) bool {
	if len(payloads) == 0 || len(payloads) != len(manifest.Components) {
		return false
	}
	byComponent := make(map[string]catalogpackagev1.OwnerPayload, len(payloads))
	for _, payload := range payloads {
		canonical, err := canonicaljsonv1.CanonicalizeObject(
			payload.Document, catalogpackagev1.MaximumPayloadBytes,
		)
		digest := sha256.Sum256(payload.Document)
		if err != nil || !bytes.Equal(canonical, payload.Document) ||
			hex.EncodeToString(digest[:]) != payload.ContentSHA256 {
			return false
		}
		if _, duplicate := byComponent[payload.ComponentID]; duplicate {
			return false
		}
		byComponent[payload.ComponentID] = payload
	}
	for _, component := range manifest.Components {
		payload, found := byComponent[component.ID.String()]
		if !found || payload.Owner != component.Owner ||
			payload.Contract != component.Contract ||
			payload.ContentSHA256 != component.ContentHash {
			return false
		}
		delete(byComponent, component.ID.String())
	}
	return len(byComponent) == 0
}
