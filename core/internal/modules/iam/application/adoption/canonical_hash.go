package adoption

import (
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const maximumCanonicalAdoptionBytes = 64 * 1024 * 1024

func canonicalSHA256(value any) (string, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal canonical input: %w", err)
	}
	digest, err := canonicaljsonv1.DigestHex(document, maximumCanonicalAdoptionBytes)
	if err != nil {
		return "", fmt.Errorf("digest canonical input: %w", err)
	}
	return digest, nil
}
