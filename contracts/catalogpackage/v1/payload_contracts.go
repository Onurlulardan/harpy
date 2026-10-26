package catalogpackagev1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const (
	ManifestContract = "catalog-package.v1"

	ContractForms     = "forms.catalog-definition.v1"
	ContractWorkflow  = "workflow.catalog-definition.v1"
	ContractIAM       = "iam.catalog-template.v1"
	ContractAnalytics = "analytics.catalog-definition.v1"
	ContractCatalog   = "catalog.demo-data.v1"

	SourceContractHR = FirstPartySourceContractPrefix + "hr"
)

var payloadDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// ValidatePayloadText rejects invalid UTF-8 and the byte-decoding signatures
// found in the legacy V1 source before any catalog payload can be signed.
func ValidatePayloadText(document []byte) error {
	if !utf8.Valid(document) {
		return fmt.Errorf("catalog payload UTF-8 is invalid")
	}
	text := string(document)
	if strings.ContainsRune(text, '\uFFFD') {
		return fmt.Errorf("catalog payload contains replacement character")
	}
	for _, sequence := range []string{
		"Ã‡", "Ã§", "Ã–", "Ã¶", "Ãœ", "Ã¼",
		"Äž", "ÄŸ", "Ä°", "Ä±", "Åž", "ÅŸ",
	} {
		if strings.Contains(text, sequence) {
			return fmt.Errorf("catalog payload contains mojibake sequence %q", sequence)
		}
	}
	return nil
}

func ValidateCanonicalPayload(
	document []byte, expectedSHA256 string, maximumBytes int,
) error {
	if err := ValidatePayloadText(document); err != nil {
		return err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, maximumBytes)
	if err != nil || !bytes.Equal(canonical, document) ||
		!payloadDigestPattern.MatchString(expectedSHA256) {
		return fmt.Errorf("catalog payload canonical contract is invalid")
	}
	digest := sha256.Sum256(document)
	if hex.EncodeToString(digest[:]) != expectedSHA256 {
		return fmt.Errorf("catalog payload digest differs")
	}
	return nil
}
