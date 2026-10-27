package iamrestorev1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func ParseSecurityEvidence(document []byte) (SecurityEvidence, error) {
	var value SecurityEvidence
	if err := decodeCanonical(document, &value); err != nil || !validSecurity(value) {
		return SecurityEvidence{}, fmt.Errorf("IAM restore security evidence is invalid")
	}
	return value, nil
}

func ParseCustodyManifest(document []byte) (CustodyManifest, error) {
	var value CustodyManifest
	if err := decodeCanonical(document, &value); err != nil || !validCustody(value) {
		return CustodyManifest{}, fmt.Errorf("IAM custody manifest is invalid")
	}
	return value, nil
}

func decodeCanonical(document []byte, target any) error {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return fmt.Errorf("document is not canonical")
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(target); err != nil {
		return err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return fmt.Errorf("document has trailing input")
	}
	return nil
}
