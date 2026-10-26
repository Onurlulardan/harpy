package iamrestorev1

import (
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func BuildSecurityEvidence(value SecurityEvidence) ([]byte, error) {
	if !validSecurity(value) {
		return nil, fmt.Errorf("IAM restore security evidence is invalid")
	}
	return canonicalDocument(value)
}

func BuildCustodyManifest(value CustodyManifest) ([]byte, error) {
	if !validCustody(value) {
		return nil, fmt.Errorf("IAM custody manifest is invalid")
	}
	return canonicalDocument(value)
}

func BindRotationResult(value RotationResult) (RotationResult, error) {
	if value.ResultSHA256 != "" {
		return RotationResult{}, fmt.Errorf("IAM custody rotation result is already bound")
	}
	value.ResultSHA256 = rotationDigest(value)
	if !validRotation(value) {
		return RotationResult{}, fmt.Errorf("IAM custody rotation result is invalid")
	}
	return value, nil
}

func canonicalDocument(value any) ([]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
}
